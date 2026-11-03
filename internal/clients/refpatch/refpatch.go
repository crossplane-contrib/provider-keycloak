// Package refpatch keeps crossplane-runtime's reference resolver from pruning
// fields out of a managed resource's spec.
//
// crossplane-runtime's managed.APISimpleReferenceResolver persists the values
// it resolved by sending a JSON merge patch computed from the *diff* of the
// current resolution pass as a server-side apply (SSA) request, using the field
// manager "managed.crossplane.io/api-simple-reference-resolver" and
// ForceOwnership. With SSA, an apply request declares the complete set of
// fields the manager wants to own: every field the manager applied in an
// earlier request but omits now is removed from the object, unless another
// manager also owns it. Because each pass only contains its own diff, values
// resolved in different reconciles evict each other. ForceOwnership also takes
// sole ownership of atomic lists (e.g. Client.authenticationFlowBindingOverrides)
// away from the manager that authored them (e.g. Argo CD), so the next resolver
// apply that does not contain the list drops it from spec.forProvider
// entirely (crossplane-contrib/provider-keycloak#750).
//
// The patch the resolver computes is a valid RFC 7386 JSON merge patch, so
// Client sends it as such: a merge patch only touches the fields it contains,
// never prunes others, and removes the keys it sets to null instead of
// failing schema validation. The object's resourceVersion is added as a
// precondition so a patch computed from a stale cached object is rejected
// with a conflict (the managed reconciler requeues) instead of overwriting a
// concurrent change to a list.
package refpatch

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FieldOwnerAPISimpleRefResolver is the field manager crossplane-runtime's
// managed.APISimpleReferenceResolver uses for its patches.
const FieldOwnerAPISimpleRefResolver = "managed.crossplane.io/api-simple-reference-resolver"

const (
	errGetPatchData   = "cannot get the reference resolver's patch data"
	errDecodePatch    = "cannot decode the reference resolver's patch"
	errEncodePatch    = "cannot encode the reference resolver's merge patch"
	errMetadataObject = "metadata of the reference resolver's patch is not an object"
)

// Client wraps a client.Client and sends the server-side apply patches of
// crossplane-runtime's reference resolver as JSON merge patches. All other
// requests are passed through unchanged.
type Client struct {
	client.Client
}

// NewClient returns a Client wrapping c.
func NewClient(c client.Client) *Client {
	return &Client{Client: c}
}

// NewClientFunc returns a client.NewClientFunc that wraps the clients created
// by newClient (client.New when nil) with a Client. It is meant to be passed
// as the NewClient option of a controller-runtime manager.
func NewClientFunc(newClient client.NewClientFunc) client.NewClientFunc {
	if newClient == nil {
		newClient = client.New
	}
	return func(config *rest.Config, options client.Options) (client.Client, error) {
		c, err := newClient(config, options)
		if err != nil {
			return nil, err
		}
		return NewClient(c), nil
	}
}

// Patch patches obj. A server-side apply patch of the reference resolver is
// converted into a JSON merge patch with a resourceVersion precondition.
func (c *Client) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if patch.Type() != types.ApplyPatchType {
		return c.Client.Patch(ctx, obj, patch, opts...)
	}
	po := &client.PatchOptions{}
	po.ApplyOptions(opts)
	if po.FieldManager != FieldOwnerAPISimpleRefResolver {
		return c.Client.Patch(ctx, obj, patch, opts...)
	}

	data, err := patch.Data(obj)
	if err != nil {
		return errors.Wrap(err, errGetPatchData)
	}
	data, err = withResourceVersion(data, obj.GetResourceVersion())
	if err != nil {
		return err
	}

	// Force is only valid for apply patches and is rejected by the API server
	// for merge patches; the field manager and dry-run options are kept.
	mergeOpts := []client.PatchOption{client.FieldOwner(po.FieldManager)}
	if len(po.DryRun) > 0 {
		mergeOpts = append(mergeOpts, client.DryRunAll)
	}
	return c.Client.Patch(ctx, obj, client.RawPatch(types.MergePatchType, data), mergeOpts...)
}

// withResourceVersion sets metadata.resourceVersion in the JSON merge patch
// data so the API server rejects the patch with a conflict if the object was
// modified after it has been read.
func withResourceVersion(data []byte, resourceVersion string) ([]byte, error) {
	if resourceVersion == "" {
		return data, nil
	}
	p := map[string]any{}
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, errors.Wrap(err, errDecodePatch)
	}
	md := map[string]any{}
	if v, ok := p["metadata"]; ok && v != nil {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, errors.New(errMetadataObject)
		}
		md = m
	}
	md["resourceVersion"] = resourceVersion
	p["metadata"] = md
	out, err := json.Marshal(p)
	return out, errors.Wrap(err, errEncodePatch)
}
