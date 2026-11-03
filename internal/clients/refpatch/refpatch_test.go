package refpatch

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
)

type patchCall struct {
	patchType types.PatchType
	data      map[string]any
	opts      client.PatchOptions
}

func recordingClient(t *testing.T, calls *[]patchCall) *test.MockClient {
	t.Helper()
	return &test.MockClient{
		MockPatch: func(_ context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			raw, err := patch.Data(obj)
			if err != nil {
				t.Fatalf("cannot get patch data: %v", err)
			}
			data := map[string]any{}
			if err := json.Unmarshal(raw, &data); err != nil {
				t.Fatalf("cannot decode patch data %q: %v", raw, err)
			}
			po := client.PatchOptions{}
			po.ApplyOptions(opts)
			*calls = append(*calls, patchCall{patchType: patch.Type(), data: data, opts: po})
			return nil
		},
	}
}

func object(resourceVersion string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("openidclient.keycloak.crossplane.io/v1alpha2")
	u.SetKind("Client")
	u.SetName("client")
	u.SetResourceVersion(resourceVersion)
	return u
}

// resolverPatch is a patch as computed by crossplane-runtime's
// APISimpleReferenceResolver: only the diff of the current resolution pass.
const resolverPatch = `{"apiVersion":"openidclient.keycloak.crossplane.io/v1alpha2","kind":"Client","spec":{"forProvider":{"authenticationFlowBindingOverrides":[{"browserId":"uuid","browserIdRef":{"name":"my-flow"}}]},"initProvider":null}}`

func TestPatch(t *testing.T) {
	forceTrue := true

	cases := map[string]struct {
		reason string
		patch  client.Patch
		opts   []client.PatchOption
		rv     string
		want   []patchCall
	}{
		"ResolverApplyBecomesMergePatch": {
			reason: "The reference resolver's server-side apply must be sent as a JSON merge patch so that fields resolved in earlier reconciles are not pruned.",
			patch:  client.RawPatch(types.ApplyPatchType, []byte(resolverPatch)),
			opts:   []client.PatchOption{client.FieldOwner(FieldOwnerAPISimpleRefResolver), client.ForceOwnership},
			rv:     "42",
			want: []patchCall{{
				patchType: types.MergePatchType,
				data: map[string]any{
					"apiVersion": "openidclient.keycloak.crossplane.io/v1alpha2",
					"kind":       "Client",
					"metadata":   map[string]any{"resourceVersion": "42"},
					"spec": map[string]any{
						"forProvider": map[string]any{
							"authenticationFlowBindingOverrides": []any{
								map[string]any{"browserId": "uuid", "browserIdRef": map[string]any{"name": "my-flow"}},
							},
						},
						"initProvider": nil,
					},
				},
				opts: client.PatchOptions{FieldManager: FieldOwnerAPISimpleRefResolver},
			}},
		},
		"ResolverApplyKeepsDryRun": {
			reason: "Dry-run requests of the reference resolver must stay dry-run.",
			patch:  client.RawPatch(types.ApplyPatchType, []byte(`{"spec":{}}`)),
			opts:   []client.PatchOption{client.FieldOwner(FieldOwnerAPISimpleRefResolver), client.ForceOwnership, client.DryRunAll},
			rv:     "7",
			want: []patchCall{{
				patchType: types.MergePatchType,
				data:      map[string]any{"metadata": map[string]any{"resourceVersion": "7"}, "spec": map[string]any{}},
				opts:      client.PatchOptions{FieldManager: FieldOwnerAPISimpleRefResolver, DryRun: []string{metav1.DryRunAll}},
			}},
		},
		"ResolverApplyWithoutResourceVersion": {
			reason: "Without a known resourceVersion the patch must be sent without a precondition.",
			patch:  client.RawPatch(types.ApplyPatchType, []byte(`{"spec":{"forProvider":{"realmId":"realm"}}}`)),
			opts:   []client.PatchOption{client.FieldOwner(FieldOwnerAPISimpleRefResolver), client.ForceOwnership},
			want: []patchCall{{
				patchType: types.MergePatchType,
				data:      map[string]any{"spec": map[string]any{"forProvider": map[string]any{"realmId": "realm"}}},
				opts:      client.PatchOptions{FieldManager: FieldOwnerAPISimpleRefResolver},
			}},
		},
		"OtherApplyPassesThrough": {
			reason: "Server-side apply patches of other field managers must not be modified.",
			patch:  client.RawPatch(types.ApplyPatchType, []byte(`{"spec":{}}`)),
			opts:   []client.PatchOption{client.FieldOwner("provider"), client.ForceOwnership},
			rv:     "42",
			want: []patchCall{{
				patchType: types.ApplyPatchType,
				data:      map[string]any{"spec": map[string]any{}},
				opts:      client.PatchOptions{FieldManager: "provider", Force: &forceTrue},
			}},
		},
		"NonApplyPatchPassesThrough": {
			reason: "Non-apply patches, even of the reference resolver's field manager, must not be modified.",
			patch:  client.RawPatch(types.MergePatchType, []byte(`{"spec":{}}`)),
			opts:   []client.PatchOption{client.FieldOwner(FieldOwnerAPISimpleRefResolver)},
			rv:     "42",
			want: []patchCall{{
				patchType: types.MergePatchType,
				data:      map[string]any{"spec": map[string]any{}},
				opts:      client.PatchOptions{FieldManager: FieldOwnerAPISimpleRefResolver},
			}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var calls []patchCall
			c := NewClient(recordingClient(t, &calls))
			if err := c.Patch(context.Background(), object(tc.rv), tc.patch, tc.opts...); err != nil {
				t.Fatalf("\n%s\nPatch(...): unexpected error: %v", tc.reason, err)
			}
			if diff := cmp.Diff(tc.want, calls, cmp.AllowUnexported(patchCall{})); diff != "" {
				t.Errorf("\n%s\nPatch(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestPatchErrors(t *testing.T) {
	cases := map[string]struct {
		reason string
		patch  client.Patch
	}{
		"InvalidJSON": {
			reason: "A patch that is not valid JSON must be rejected.",
			patch:  client.RawPatch(types.ApplyPatchType, []byte(`{`)),
		},
		"MetadataNotAnObject": {
			reason: "A patch whose metadata is not an object must be rejected.",
			patch:  client.RawPatch(types.ApplyPatchType, []byte(`{"metadata":"invalid"}`)),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var calls []patchCall
			c := NewClient(recordingClient(t, &calls))
			err := c.Patch(context.Background(), object("1"), tc.patch, client.FieldOwner(FieldOwnerAPISimpleRefResolver), client.ForceOwnership)
			if err == nil {
				t.Errorf("\n%s\nPatch(...): expected an error", tc.reason)
			}
			if len(calls) != 0 {
				t.Errorf("\n%s\nPatch(...): expected no request, got %d", tc.reason, len(calls))
			}
		})
	}
}

func TestNewClientFunc(t *testing.T) {
	inner := &test.MockClient{}
	got, err := NewClientFunc(func(_ *rest.Config, _ client.Options) (client.Client, error) {
		return inner, nil
	})(&rest.Config{}, client.Options{})
	if err != nil {
		t.Fatalf("NewClientFunc(...): unexpected error: %v", err)
	}
	wrapped, ok := got.(*Client)
	if !ok {
		t.Fatalf("NewClientFunc(...): want *Client, got %T", got)
	}
	if wrapped.Client != inner {
		t.Errorf("NewClientFunc(...): the created client is not wrapped")
	}

	errBoom := errors.New("boom")
	if _, err := NewClientFunc(func(_ *rest.Config, _ client.Options) (client.Client, error) {
		return nil, errBoom
	})(&rest.Config{}, client.Options{}); !errors.Is(err, errBoom) {
		t.Errorf("NewClientFunc(...): want error %v, got %v", errBoom, err)
	}
}
