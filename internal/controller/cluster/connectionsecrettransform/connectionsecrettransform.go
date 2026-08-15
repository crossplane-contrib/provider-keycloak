/*
Copyright 2022 Upbound Inc.
*/

// Package connectionsecrettransform implements a hand-written (i.e. not
// upjet-generated) controller that republishes the connection secret of a
// managed resource with renamed keys. The renaming is configured either
// centrally, on the ProviderConfig the resource uses, or per resource, via
// annotations on the managed resource itself.
//
// Background: connection secret keys are published today via
// config.Resource.Sensitive.AdditionalConnectionDetailsFn (see
// config/openidclient/config.go), which upjet invokes with only the
// Terraform resource's own state - it has no access to the resource's
// resource.Managed object or its ProviderConfig (see
// github.com/crossplane/upjet/v2 v2.3.0, pkg/resource/sensitive.go:
// GetConnectionDetails). That makes it impossible to rename keys, or add
// ProviderConfig-derived values, from inside that hook. This controller
// sidesteps that limitation entirely by operating downstream, on
// already-materialized Kubernetes objects (the Secret written by the managed
// resource's own controller, the owning managed resource, and its
// ProviderConfig), all of which a normal controller-runtime client can read
// directly.
package connectionsecrettransform

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane/upjet/v2/pkg/controller"

	clusterv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/v1beta1"
)

const (
	// AnnotationKeyRename configures the renaming for a single managed
	// resource, as a comma-separated list of "<oldKey>=<newKey>" pairs, e.g.
	//
	//	keycloak.crossplane.io/connection-secret-key-rename: clientID=client-id,clientSecret=client-secret
	//
	// Entries here are merged on top of the ProviderConfig-wide
	// spec.connectionSecretKeys.rename map, so a resource can extend or
	// override the central configuration without replacing it.
	AnnotationKeyRename = "keycloak.crossplane.io/connection-secret-key-rename"

	// AnnotationKeyTransformedName overrides the name of the republished
	// secret. It defaults to the name of the connection secret plus the
	// "-transformed" suffix, and is useful when a consumer (e.g. an Envoy
	// Gateway SecurityPolicy) requires a specific secret name.
	AnnotationKeyTransformedName = "keycloak.crossplane.io/connection-secret-transform-name"

	// transformedSecretSuffix is appended to the name of the source
	// connection secret to build the default name of the republished,
	// transformed one. The two live side by side: the original is left
	// untouched so that upjet's tfstate-rebuild-from-connection-secret
	// machinery keeps working, and consumers that need renamed keys point at
	// the transformed secret instead.
	transformedSecretSuffix = "-transformed"

	// transformedSecretLabel marks secrets written by this controller, so
	// that it can ignore its own output when reconciling Secret events
	// instead of trying to transform an already-transformed secret.
	transformedSecretLabel = "keycloak.crossplane.io/connection-secret-transform"

	// sourceSecretLabel records the name of the connection secret a
	// transformed secret was derived from. It is what makes stale output
	// (e.g. after the transform name annotation changed, or after the
	// renaming was removed altogether) findable and thus collectable.
	sourceSecretLabel = "keycloak.crossplane.io/connection-secret-source"

	// transformedSecretLabelValue is the value of transformedSecretLabel on
	// secrets written by this controller.
	transformedSecretLabelValue = "true"

	clusterGroup    = "keycloak.crossplane.io"
	namespacedGroup = "keycloak.m.crossplane.io"

	// defaultResyncInterval is how often a connection secret owned by a
	// Keycloak managed resource is re-reconciled when the controller options
	// do not specify a poll interval. The rename configuration lives on the
	// managed resource (annotations) and the ProviderConfig, neither of which
	// this controller watches, so a change there does not produce a Secret
	// event; the resync is what makes such edits converge.
	defaultResyncInterval = time.Minute
)

// secretKeyRE matches the key names Kubernetes accepts in a Secret. Renaming
// to anything else would make the write fail, so such entries are skipped.
var secretKeyRE = regexp.MustCompile(`^[-._a-zA-Z0-9]+$`)

// Reconciler republishes a managed resource's connection secret with keys
// renamed according to its ProviderConfig's spec.connectionSecretKeys.rename
// map and/or the resource's own rename annotation.
type Reconciler struct {
	client client.Client

	// resyncInterval is the interval a Keycloak-owned connection secret is
	// requeued with, see defaultResyncInterval.
	resyncInterval time.Duration
}

// Setup adds a controller that watches connection Secrets owned by Keycloak
// managed resources and republishes them with renamed keys.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	interval := o.PollInterval
	if interval <= 0 {
		interval = defaultResyncInterval
	}
	r := &Reconciler{client: mgr.GetClient(), resyncInterval: interval}

	return ctrl.NewControllerManagedBy(mgr).
		Named("connectionsecrettransform").
		WithOptions(o.ForControllerRuntime()).
		For(&corev1.Secret{}).
		Complete(r)
}

// SetupGated adds this controller directly, bypassing the CRD-establishment
// gate other controllers use. Unlike upjet-generated controllers, this one
// does not own a CRD of its own: it only needs the Secret, managed resource
// and ProviderConfig kinds it reads to already be registered with the
// client's scheme, which is guaranteed at manager start-up regardless of CRD
// establishment order.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return Setup(mgr, o)
}

// Reconcile implements reconcile.Reconciler. It is a no-op for any Secret
// that is not the connection secret of a Keycloak managed resource for which
// a key renaming is configured.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	secret := &corev1.Secret{}
	if err := r.client.Get(ctx, req.NamespacedName, secret); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Never try to transform our own output.
	if secret.Labels[transformedSecretLabel] == transformedSecretLabelValue {
		return reconcile.Result{}, nil
	}

	owner := findManagedOwner(secret.OwnerReferences)
	if owner == nil {
		return reconcile.Result{}, nil
	}

	mr, err := r.getOwner(ctx, *owner, secret.Namespace)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	rename, err := r.renameMap(ctx, mr)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(rename) == 0 {
		// Nothing (any more) to transform: collect what we wrote earlier.
		return r.resync(), r.deleteStale(ctx, secret, "")
	}

	name := secret.Name + transformedSecretSuffix
	if n := mr.GetAnnotations()[AnnotationKeyTransformedName]; n != "" {
		name = n
	}

	out := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: secret.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.client, out, func() error {
		if out.Labels == nil {
			out.Labels = map[string]string{}
		}
		out.Labels[transformedSecretLabel] = transformedSecretLabelValue
		out.Labels[sourceSecretLabel] = secret.Name
		out.Type = secret.Type
		out.Data = transform(secret.Data, rename)
		// Tie the transformed secret's lifecycle to the source connection
		// secret, which is itself owned (with a controller reference) by the
		// managed resource. Deleting the managed resource therefore
		// garbage-collects the connection secret and, with it, the
		// transformed one - and so does dropping writeConnectionSecretToRef
		// from a resource that stays around.
		return controllerutil.SetControllerReference(secret, out, r.client.Scheme())
	}); err != nil {
		return reconcile.Result{}, err
	}

	return r.resync(), r.deleteStale(ctx, secret, name)
}

// resync returns the result Keycloak-owned connection secrets are requeued
// with. The rename configuration lives on objects this controller does not
// watch (the managed resource's annotations and its ProviderConfig), so
// editing only those produces no Secret event; requeueing periodically is
// what applies such an edit without touching the connection secret itself.
func (r *Reconciler) resync() reconcile.Result {
	return reconcile.Result{RequeueAfter: r.resyncInterval}
}

// renameMap merges the ProviderConfig-wide renaming with the managed
// resource's own annotation, the latter taking precedence per key.
func (r *Reconciler) renameMap(ctx context.Context, mr *unstructured.Unstructured) (map[string]string, error) {
	rename := map[string]string{}

	// Only cluster-scoped managed resources reference the cluster-scoped
	// ProviderConfig this provider configures renaming on. Namespaced
	// resources are configured through the annotation alone.
	if mr.GetNamespace() == "" {
		pc, err := r.providerConfig(ctx, mr)
		if err != nil {
			return nil, err
		}
		if pc != nil && pc.Spec.ConnectionSecretKeys != nil {
			for k, v := range pc.Spec.ConnectionSecretKeys.Rename {
				rename[k] = v
			}
		}
	}

	for k, v := range parseRenameAnnotation(mr.GetAnnotations()[AnnotationKeyRename]) {
		rename[k] = v
	}

	for k, v := range rename {
		if !secretKeyRE.MatchString(v) {
			delete(rename, k)
		}
	}
	return rename, nil
}

// providerConfig reads the cluster-scoped ProviderConfig the managed resource
// references, if any.
func (r *Reconciler) providerConfig(ctx context.Context, mr *unstructured.Unstructured) (*clusterv1beta1.ProviderConfig, error) {
	name, found, err := unstructured.NestedString(mr.Object, "spec", "providerConfigRef", "name")
	if err != nil || !found || name == "" {
		return nil, err
	}

	pc := &clusterv1beta1.ProviderConfig{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: name}, pc); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return pc, nil
}

// getOwner reads the managed resource owning the connection secret. Only the
// namespaced API group (*.keycloak.m.crossplane.io) holds namespaced managed
// resources; the cluster-scoped ones are read without a namespace.
func (r *Reconciler) getOwner(ctx context.Context, owner metav1.OwnerReference, namespace string) (*unstructured.Unstructured, error) {
	gv, _ := schema.ParseGroupVersion(owner.APIVersion)

	mr := &unstructured.Unstructured{}
	mr.SetGroupVersionKind(gv.WithKind(owner.Kind))

	key := types.NamespacedName{Name: owner.Name}
	if strings.HasSuffix(gv.Group, namespacedGroup) {
		key.Namespace = namespace
	}
	if err := r.client.Get(ctx, key, mr); err != nil {
		return nil, err
	}
	return mr, nil
}

// deleteStale removes transformed secrets this controller previously wrote
// for the given source secret but no longer wants, i.e. all of them when the
// renaming was removed, or the previous one after the name annotation
// changed.
func (r *Reconciler) deleteStale(ctx context.Context, source *corev1.Secret, keep string) error {
	l := &corev1.SecretList{}
	if err := r.client.List(ctx, l, client.InNamespace(source.Namespace), client.MatchingLabels{
		transformedSecretLabel: transformedSecretLabelValue,
		sourceSecretLabel:      source.Name,
	}); err != nil {
		return err
	}

	for i := range l.Items {
		s := &l.Items[i]
		if s.Name == keep || !metav1.IsControlledBy(s, source) {
			continue
		}
		if err := r.client.Delete(ctx, s); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

// transform copies data over, renaming the configured keys. Keys are applied
// in a deterministic order so that a (misconfigured) renaming of two keys
// onto the same target always yields the same result.
func transform(data map[string][]byte, rename map[string]string) map[string][]byte {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(map[string][]byte, len(data))
	for _, k := range keys {
		newKey := k
		if renamed, ok := rename[k]; ok {
			newKey = renamed
		}
		out[newKey] = data[k]
	}
	return out
}

// parseRenameAnnotation parses a comma-separated list of "<oldKey>=<newKey>"
// pairs. Malformed entries are ignored rather than failing the whole
// reconcile, which would otherwise leave a resource stuck on a typo.
func parseRenameAnnotation(v string) map[string]string {
	if v == "" {
		return nil
	}
	rename := map[string]string{}
	for _, pair := range strings.Split(v, ",") {
		oldKey, newKey, found := strings.Cut(strings.TrimSpace(pair), "=")
		oldKey, newKey = strings.TrimSpace(oldKey), strings.TrimSpace(newKey)
		if !found || oldKey == "" || newKey == "" {
			continue
		}
		rename[oldKey] = newKey
	}
	return rename
}

// findManagedOwner returns the owner reference pointing at a Keycloak managed
// resource, i.e. the resource whose controller wrote the connection secret.
func findManagedOwner(refs []metav1.OwnerReference) *metav1.OwnerReference {
	for i := range refs {
		gv, err := schema.ParseGroupVersion(refs[i].APIVersion)
		if err != nil {
			continue
		}
		// Managed resources live in a per-API-group subdomain, e.g.
		// openidclient.keycloak.crossplane.io. The bare provider groups
		// themselves hold only ProviderConfig and friends.
		if strings.HasSuffix(gv.Group, "."+clusterGroup) || strings.HasSuffix(gv.Group, "."+namespacedGroup) {
			return &refs[i]
		}
	}
	return nil
}
