/*
Copyright 2022 Upbound Inc.
*/

// Package connectionsecrettransform implements a hand-written (i.e. not
// upjet-generated) controller that republishes the connection secret of a
// managed resource with renamed keys. The renaming is configured centrally
// on the ProviderConfig the resource uses, per resource via annotations on
// the managed resource itself, and/or via a standalone
// ConnectionSecretTransform custom resource (clusterv1beta1) that names the
// source secret directly - useful when the managed resource is owned by a
// Composition or otherwise inconvenient to annotate. All three sources may
// be combined; a ConnectionSecretTransform takes precedence over the
// annotation, which in turn takes precedence over the ProviderConfig, per
// key.
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
//
// The controller is deliberately conservative: it only ever writes secrets it
// created itself (marked with the transform label and controller-owned by the
// source connection secret), never modifies the source secret, and treats
// every misconfiguration (an unusable target key, a target name that is not a
// valid secret name or that belongs to somebody else, two keys renamed onto
// one another) as a skipped operation reported via a Kubernetes event rather
// than as a reconcile error. A misconfiguration can therefore neither destroy
// data nor spin the work queue.
package connectionsecrettransform

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/controller"

	clusterv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/v1beta1"
)

const (
	// AnnotationKeyRename configures the renaming for a single managed
	// resource. Two syntaxes are supported, tried in this order:
	//
	//  1. A YAML block mapping of "<oldKey>: <newKey>" pairs:
	//
	//     keycloak.crossplane.io/connection-secret-key-rename: |
	//       clientID: client-id
	//       clientSecret: client-secret
	//
	//  2. A comma- and/or newline-separated list of "<oldKey>=<newKey>"
	//     pairs, e.g.
	//
	//     keycloak.crossplane.io/connection-secret-key-rename: clientID=client-id,clientSecret=client-secret
	//
	//     Kubernetes annotations may be multi-line (a YAML block scalar),
	//     which reads better for longer lists:
	//
	//     keycloak.crossplane.io/connection-secret-key-rename: |
	//       clientID=client-id
	//       clientSecret=client-secret
	//
	// The value is first tried as a YAML mapping; if it does not parse as
	// one (e.g. it uses "=" instead of ": ", which YAML treats as a plain
	// scalar, not a mapping) it falls back to the "key=value" syntax. Either
	// way, entries here are merged on top of the ProviderConfig-wide
	// spec.connectionSecretKeys.rename map, so a resource can extend or
	// override the central configuration without replacing it.
	AnnotationKeyRename = "keycloak.crossplane.io/connection-secret-key-rename"

	// AnnotationKeyTransformedName overrides the name of the republished
	// secret. It defaults to the name of the connection secret plus the
	// "-transformed" suffix, and is useful when a consumer (e.g. an Envoy
	// Gateway SecurityPolicy) requires a specific secret name.
	AnnotationKeyTransformedName = "keycloak.crossplane.io/connection-secret-transform-name"

	// AnnotationKeyAddFields adds extra keys to the transformed connection
	// secret, sourced from elsewhere rather than renamed from the secret
	// itself. Same YAML-mapping-or-"key=value" syntax as AnnotationKeyRename
	// (see its doc comment for both forms), e.g.
	//
	//	keycloak.crossplane.io/connection-secret-add-fields: |
	//	  issuerUrl: providerConfig:metadata.name
	//	  internalClientId: status:atProvider.id
	//
	// or, using the "key=value" fallback syntax:
	//
	//	keycloak.crossplane.io/connection-secret-add-fields: |
	//	  issuerUrl=providerConfig:metadata.name
	//	  internalClientId=status:atProvider.id
	//
	// where <source> is one of:
	//   - "status:<dot.path>": a field under the owning managed resource's
	//     status (e.g. "status:atProvider.clientId").
	//   - "providerConfig:<dot.path>": a field of the ProviderConfig the
	//     resource references (e.g. "providerConfig:metadata.name"). Only
	//     available for cluster-scoped resources, same restriction as
	//     ProviderConfig-wide renaming.
	//
	// Only scalar (string, number, boolean) fields are supported. Entries
	// here are merged on top of the ProviderConfig-wide
	// spec.connectionSecretKeys.add map, same precedence as renaming.
	AnnotationKeyAddFields = "keycloak.crossplane.io/connection-secret-add-fields"

	// transformedSecretSuffix is appended to the name of the source
	// connection secret to build the default name of the republished,
	// transformed one. The two live side by side: the original is left
	// untouched so that upjet's tfstate-rebuild-from-connection-secret
	// machinery keeps working, and consumers that need renamed keys point at
	// the transformed secret instead.
	transformedSecretSuffix = "-transformed"

	// transformedSecretLabel marks secrets written by this controller, so
	// that it can ignore its own output when reconciling Secret events
	// instead of trying to transform an already-transformed secret, and so
	// that it can tell its own output from a pre-existing, foreign secret it
	// must not touch.
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

	// controllerName names the controller, its event source and its log
	// entries.
	controllerName = "connectionsecrettransform"

	// defaultResyncInterval is how often a connection secret owned by a
	// Keycloak managed resource is re-reconciled when the controller options
	// do not specify a poll interval. The rename configuration lives on the
	// managed resource (annotations) and the ProviderConfig, neither of which
	// this controller watches, so a change there does not produce a Secret
	// event; the resync is what makes such edits converge.
	defaultResyncInterval = time.Minute

	// addSourceStatus and addSourceProviderConfig are the recognized source
	// prefixes for AnnotationKeyAddFields / spec.connectionSecretKeys.add
	// entries. See AnnotationKeyAddFields for the full syntax.
	addSourceStatus         = "status:"
	addSourceProviderConfig = "providerConfig:"
)

// Event reasons reported on the source connection secret.
const (
	reasonInvalidRename      event.Reason = "InvalidConnectionSecretKeyRename"
	reasonRenameConflict     event.Reason = "ConnectionSecretKeyRenameConflict"
	reasonInvalidName        event.Reason = "InvalidTransformedSecretName"
	reasonNotOwned           event.Reason = "TransformedSecretNotOwned"
	reasonTransformFailure   event.Reason = "CannotTransformConnectionSecret"
	reasonInvalidAddField    event.Reason = "InvalidConnectionSecretFieldAdd"
	reasonAddFieldConflict   event.Reason = "ConnectionSecretFieldAddConflict"
	reasonAmbiguousTransform event.Reason = "ConnectionSecretTransformAmbiguous"
)

// secretKeyRE matches the key names Kubernetes accepts in a Secret. Renaming
// to anything else would make the write fail, so such entries are skipped.
var secretKeyRE = regexp.MustCompile(`^[-._a-zA-Z0-9]+$`)

// Reconciler republishes a managed resource's connection secret with keys
// renamed according to its ProviderConfig's spec.connectionSecretKeys.rename
// map and/or the resource's own rename annotation.
type Reconciler struct {
	client client.Client
	log    logging.Logger
	record event.Recorder

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

	r := &Reconciler{
		client: mgr.GetClient(),
		log:    logger(o).WithValues("controller", controllerName),
		// crossplane-runtime's event.APIRecorder takes the (deprecated)
		// client-go record.EventRecorder, which is what every generated
		// controller in this provider uses too.
		record:         event.NewAPIRecorder(mgr.GetEventRecorderFor(controllerName)), //nolint:staticcheck // Required by crossplane-runtime's event.NewAPIRecorder.
		resyncInterval: interval,
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(o.ForControllerRuntime()).
		// Only Crossplane connection secrets and this controller's own output
		// are of interest. Everything else - most notably the provider's own
		// credentials secrets and every unrelated secret in the cluster - is
		// filtered out before it reaches the work queue.
		For(&corev1.Secret{}, builder.WithPredicates(predicate.NewPredicateFuncs(isRelevantSecret))).
		// A ConnectionSecretTransform names its source secret directly, so an
		// edit to it (or its creation/deletion) must reconcile that secret
		// even though the secret itself did not change.
		Watches(&clusterv1beta1.ConnectionSecretTransform{}, handler.EnqueueRequestsFromMapFunc(requestFromTransform)).
		Complete(r)
}

// requestFromTransform maps a ConnectionSecretTransform to a reconcile
// request for the connection secret it names, so that creating, editing or
// deleting one reconciles that secret immediately instead of waiting for the
// resync interval.
func requestFromTransform(_ context.Context, o client.Object) []reconcile.Request {
	t, ok := o.(*clusterv1beta1.ConnectionSecretTransform)
	if !ok || t.Spec.SourceSecretRef.Name == "" {
		return nil
	}
	return []reconcile.Request{{NamespacedName: types.NamespacedName{
		Namespace: t.Namespace,
		Name:      t.Spec.SourceSecretRef.Name,
	}}}
}

// SetupGated adds this controller directly, bypassing the CRD-establishment
// gate other controllers use. It now does own one CRD, ConnectionSecretTransform,
// but the controller-manager only calls SetupGated once every provider CRD -
// including that one - is established (see cmd/provider/main.go), so there is
// no ordering hazard in starting it the same way as before.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return Setup(mgr, o)
}

// logger returns the configured logger, or a no-op one when the options carry
// none.
func logger(o controller.Options) logging.Logger {
	if o.Logger == nil {
		return logging.NewNopLogger()
	}
	return o.Logger
}

// isRelevantSecret reports whether a Secret event is worth queueing: either it
// is a Crossplane connection secret (a potential transform source) or it is a
// secret this controller wrote (which it must observe to correct drift).
func isRelevantSecret(o client.Object) bool {
	s, ok := o.(*corev1.Secret)
	if !ok {
		return false
	}
	return s.Type == xpresource.SecretTypeConnection || s.Labels[transformedSecretLabel] == transformedSecretLabelValue
}

// Reconcile implements reconcile.Reconciler. It is a no-op for any Secret
// that is not the connection secret of a Keycloak managed resource for which
// a key renaming is configured.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) { //nolint:gocyclo // A linear sequence of cheap guards; splitting it would obscure the flow.
	log := r.log.WithValues("secret", req.String())

	secret := &corev1.Secret{}
	if err := r.client.Get(ctx, req.NamespacedName, secret); err != nil {
		return reconcile.Result{}, errors.Wrap(client.IgnoreNotFound(err), "cannot get connection secret")
	}

	// Never try to transform our own output, and never fight the garbage
	// collector over a secret that is on its way out.
	if secret.Labels[transformedSecretLabel] == transformedSecretLabelValue || secret.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	crd, err := r.crdTransform(ctx, secret)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "cannot resolve ConnectionSecretTransform")
	}

	owner := findManagedOwner(secret.OwnerReferences)
	if owner == nil {
		return reconcile.Result{}, nil
	}

	mr, err := r.getOwner(ctx, *owner, secret.Namespace)
	if err != nil {
		if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			// The managed resource is gone (its connection secret is about to
			// be garbage-collected) or its CRD is not established yet.
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, errors.Wrap(err, "cannot get the managed resource owning the connection secret")
	}

	rename, invalid, err := r.renameMap(ctx, mr, crd)
	if err != nil {
		r.record.Event(secret, event.Warning(reasonTransformFailure, err))
		return reconcile.Result{}, errors.Wrap(err, "cannot determine the connection secret key renaming")
	}
	if len(invalid) > 0 {
		// A typo must not wedge the resource: the remaining, valid entries
		// are still applied and the bad ones are reported.
		log.Debug("Ignoring unusable connection secret key renames", "entries", strings.Join(invalid, ", "))
		r.record.Event(secret, event.Warning(reasonInvalidRename,
			errors.Errorf("ignoring rename(s) with an unusable target key: %s", strings.Join(invalid, ", "))))
	}

	added, invalidAdd, err := r.addFieldMap(ctx, mr, crd)
	if err != nil {
		r.record.Event(secret, event.Warning(reasonTransformFailure, err))
		return reconcile.Result{}, errors.Wrap(err, "cannot determine the connection secret fields to add")
	}
	if len(invalidAdd) > 0 {
		// Same tolerance as renaming: an unresolved source must not wedge
		// the resource, the remaining entries are still applied.
		log.Debug("Ignoring unresolved connection secret field(s) to add", "entries", strings.Join(invalidAdd, ", "))
		r.record.Event(secret, event.Warning(reasonInvalidAddField,
			errors.Errorf("ignoring field(s) to add that could not be resolved: %s", strings.Join(invalidAdd, ", "))))
	}

	if len(rename) == 0 && len(added) == 0 {
		// Nothing (any more) to transform: collect what we wrote earlier.
		r.setTransformStatus(ctx, crd, "", xpv1.Unavailable().WithMessage("nothing to rename or add"))
		return r.resync(), errors.Wrap(r.deleteStale(ctx, secret, ""), errDeleteStale)
	}

	name, err := transformedName(secret, mr, crd)
	if err != nil {
		log.Debug("Refusing to write a transformed connection secret", "error", err)
		r.record.Event(secret, event.Warning(reasonInvalidName, err))
		r.setTransformStatus(ctx, crd, "", xpv1.Unavailable().WithMessage(err.Error()))
		// Terminal until the annotation is fixed. Returning no error keeps
		// the work queue calm; the resync picks the correction up.
		return r.resync(), errors.Wrap(r.deleteStale(ctx, secret, ""), errDeleteStale)
	}

	owned, err := r.isOursOrAbsent(ctx, secret, name)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "cannot inspect the transformed secret")
	}
	if !owned {
		// Somebody else's secret occupies the target name. Overwriting it
		// could destroy unrelated data - the provider's own credentials, say
		// - so we refuse, and say why.
		err := errors.Errorf("secret %q already exists and was not written by this controller", name)
		log.Debug("Refusing to overwrite a foreign secret", "target", name)
		r.record.Event(secret, event.Warning(reasonNotOwned, err))
		r.setTransformStatus(ctx, crd, "", xpv1.Unavailable().WithMessage(err.Error()))
		return r.resync(), errors.Wrap(r.deleteStale(ctx, secret, ""), errDeleteStale)
	}

	data, conflicts := transform(secret.Data, rename)
	if len(conflicts) > 0 {
		// Applying these would silently drop a key, so the affected renames
		// are skipped (those keys keep their original name) and reported.
		log.Debug("Skipping conflicting connection secret key renames", "entries", strings.Join(conflicts, ", "))
		r.record.Event(secret, event.Warning(reasonRenameConflict,
			errors.Errorf("skipping rename(s) that would overwrite another key: %s", strings.Join(conflicts, ", "))))
	}

	addConflicts := mergeAdded(data, added)
	if len(addConflicts) > 0 {
		// Same rule as renaming: never silently drop a value that is
		// already there.
		log.Debug("Skipping added connection secret field(s) that would overwrite an existing key", "entries", strings.Join(addConflicts, ", "))
		r.record.Event(secret, event.Warning(reasonAddFieldConflict,
			errors.Errorf("skipping added field(s) that would overwrite an existing key: %s", strings.Join(addConflicts, ", "))))
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
		out.Data = data
		// Tie the transformed secret's lifecycle to the source connection
		// secret, which is itself owned (with a controller reference) by the
		// managed resource. Deleting the managed resource therefore
		// garbage-collects the connection secret and, with it, the
		// transformed one - and so does dropping writeConnectionSecretToRef
		// from a resource that stays around.
		return controllerutil.SetControllerReference(secret, out, r.client.Scheme())
	}); err != nil {
		r.record.Event(secret, event.Warning(reasonTransformFailure, err))
		r.setTransformStatus(ctx, crd, "", xpv1.Unavailable().WithMessage(err.Error()))
		return reconcile.Result{}, errors.Wrap(err, "cannot write the transformed connection secret")
	}

	r.setTransformStatus(ctx, crd, name, xpv1.Available().WithMessage(fmt.Sprintf("published as secret %q", name)))
	return r.resync(), errors.Wrap(r.deleteStale(ctx, secret, name), errDeleteStale)
}

const errDeleteStale = "cannot delete stale transformed secrets"

// resync returns the result Keycloak-owned connection secrets are requeued
// with. The rename configuration lives on objects this controller does not
// watch (the managed resource's annotations and its ProviderConfig), so
// editing only those produces no Secret event; requeueing periodically is
// what applies such an edit without touching the connection secret itself.
func (r *Reconciler) resync() reconcile.Result {
	return reconcile.Result{RequeueAfter: r.resyncInterval}
}

// isOursOrAbsent reports whether the target name is available, i.e. either no
// secret exists under it or the one that does was written by this controller
// for this very source secret.
func (r *Reconciler) isOursOrAbsent(ctx context.Context, source *corev1.Secret, name string) (bool, error) {
	existing := &corev1.Secret{}
	err := r.client.Get(ctx, types.NamespacedName{Name: name, Namespace: source.Namespace}, existing)
	if apierrors.IsNotFound(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return existing.Labels[transformedSecretLabel] == transformedSecretLabelValue && metav1.IsControlledBy(existing, source), nil
}

// renameMap merges the ProviderConfig-wide renaming, the managed resource's
// own annotation, and a matching ConnectionSecretTransform's spec.rename, in
// that order of precedence per key. Entries whose target Kubernetes would
// reject as a secret key are dropped and returned separately so that the
// caller can report them.
func (r *Reconciler) renameMap(ctx context.Context, mr *unstructured.Unstructured, crd *clusterv1beta1.ConnectionSecretTransform) (map[string]string, []string, error) {
	rename := map[string]string{}

	// Only cluster-scoped managed resources reference the cluster-scoped
	// ProviderConfig this provider configures renaming on. Namespaced
	// resources are configured through the annotation or a
	// ConnectionSecretTransform instead.
	if mr.GetNamespace() == "" {
		pc, err := r.providerConfig(ctx, mr)
		if err != nil {
			return nil, nil, err
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

	if crd != nil {
		for k, v := range crd.Spec.Rename {
			rename[k] = v
		}
	}

	var invalid []string
	for k, v := range rename {
		if !secretKeyRE.MatchString(v) {
			invalid = append(invalid, fmt.Sprintf("%q=%q", k, v))
			delete(rename, k)
		}
	}
	sort.Strings(invalid)
	return rename, invalid, nil
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

// crdTransform returns the single ConnectionSecretTransform in the secret's
// namespace that names it via spec.sourceSecretRef, if exactly one exists. It
// returns nil, nil when none does. More than one would make the merge order
// among them ambiguous, so that case is treated as a misconfiguration: it is
// reported as an event on the source secret and reflected in each matching
// object's own status, and none of them contribute to the transform.
func (r *Reconciler) crdTransform(ctx context.Context, secret *corev1.Secret) (*clusterv1beta1.ConnectionSecretTransform, error) {
	l := &clusterv1beta1.ConnectionSecretTransformList{}
	if err := r.client.List(ctx, l, client.InNamespace(secret.Namespace)); err != nil {
		return nil, errors.Wrap(err, "cannot list ConnectionSecretTransforms")
	}

	var matches []*clusterv1beta1.ConnectionSecretTransform
	for i := range l.Items {
		if l.Items[i].Spec.SourceSecretRef.Name == secret.Name {
			matches = append(matches, &l.Items[i])
		}
	}

	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		return matches[0], nil
	default:
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			names = append(names, m.Name)
		}
		sort.Strings(names)
		err := errors.Errorf("secret %q is named by multiple ConnectionSecretTransforms: %s", secret.Name, strings.Join(names, ", "))
		r.record.Event(secret, event.Warning(reasonAmbiguousTransform, err))
		for _, m := range matches {
			r.setTransformStatus(ctx, m, "", xpv1.Unavailable().WithMessage(err.Error()))
		}
		return nil, nil
	}
}

// setTransformStatus records the outcome of a reconcile on a
// ConnectionSecretTransform's own status. It is a no-op when crd is nil (no
// ConnectionSecretTransform is involved) and best-effort otherwise: a failure
// to persist the status is logged but never fails the reconcile, since the
// status is purely observational.
func (r *Reconciler) setTransformStatus(ctx context.Context, crd *clusterv1beta1.ConnectionSecretTransform, secretName string, cond xpv1.Condition) {
	if crd == nil {
		return
	}
	orig := crd.DeepCopy()
	crd.Status.SetConditions(cond)
	crd.Status.TransformedSecretName = secretName
	if err := r.client.Status().Patch(ctx, crd, client.MergeFrom(orig)); err != nil {
		r.log.Debug("cannot update ConnectionSecretTransform status", "name", crd.Name, "namespace", crd.Namespace, "error", err)
	}
}

// mergedAddSources merges the ProviderConfig-wide spec.connectionSecretKeys.add
// map, the managed resource's own AnnotationKeyAddFields annotation, and a
// matching ConnectionSecretTransform's spec.add, in that order of
// precedence per key. It also returns the ProviderConfig (if any), since
// resolveAddSource needs it as a possible value source.
func (r *Reconciler) mergedAddSources(ctx context.Context, mr *unstructured.Unstructured, crd *clusterv1beta1.ConnectionSecretTransform) (*clusterv1beta1.ProviderConfig, map[string]string, error) {
	add := map[string]string{}

	// Same restriction as renameMap: the cluster-scoped ProviderConfig is
	// only available to cluster-scoped resources.
	var pc *clusterv1beta1.ProviderConfig
	if mr.GetNamespace() == "" {
		var err error
		pc, err = r.providerConfig(ctx, mr)
		if err != nil {
			return nil, nil, err
		}
		if pc != nil && pc.Spec.ConnectionSecretKeys != nil {
			for k, v := range pc.Spec.ConnectionSecretKeys.Add {
				add[k] = v
			}
		}
	}

	for k, v := range parseRenameAnnotation(mr.GetAnnotations()[AnnotationKeyAddFields]) {
		add[k] = v
	}

	if crd != nil {
		for k, v := range crd.Spec.Add {
			add[k] = v
		}
	}
	return pc, add, nil
}

// addFieldMap merges the ProviderConfig-wide spec.connectionSecretKeys.add
// map, the managed resource's own AnnotationKeyAddFields annotation, and a
// matching ConnectionSecretTransform's spec.add, in that order of precedence
// per key, and resolves every source expression into a value. Entries with
// an unusable target key or a source that cannot be resolved are dropped and
// returned separately so the caller can report them, mirroring renameMap's
// tolerance for a single bad entry.
func (r *Reconciler) addFieldMap(ctx context.Context, mr *unstructured.Unstructured, crd *clusterv1beta1.ConnectionSecretTransform) (map[string][]byte, []string, error) {
	pc, add, err := r.mergedAddSources(ctx, mr, crd)
	if err != nil {
		return nil, nil, err
	}
	if len(add) == 0 {
		return nil, nil, nil
	}

	var pcObj map[string]interface{}
	if pc != nil {
		conv, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pc)
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot convert ProviderConfig to unstructured")
		}
		pcObj = conv
	}

	keys := make([]string, 0, len(add))
	for k := range add {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(map[string][]byte, len(add))
	var invalid []string
	for _, k := range keys {
		expr := add[k]
		if !secretKeyRE.MatchString(k) {
			invalid = append(invalid, fmt.Sprintf("%q: not a valid secret key", k))
			continue
		}
		val, err := resolveAddSource(expr, mr.Object, pcObj)
		if err != nil {
			invalid = append(invalid, fmt.Sprintf("%s=%s: %s", k, expr, err))
			continue
		}
		out[k] = []byte(val)
	}
	sort.Strings(invalid)
	return out, invalid, nil
}

// resolveAddSource resolves a single AnnotationKeyAddFields source
// expression ("status:<dot.path>" or "providerConfig:<dot.path>") against
// the managed resource's own object or the (possibly nil) ProviderConfig's,
// converted to unstructured content. Only scalar values are supported: a
// map or a list would silently coerce to an unhelpful string via fmt.Sprint,
// which is worse than refusing it outright.
func resolveAddSource(expr string, mrObj, pcObj map[string]interface{}) (string, error) {
	var root map[string]interface{}
	var path string
	switch {
	case strings.HasPrefix(expr, addSourceStatus):
		status, _, err := unstructured.NestedMap(mrObj, "status")
		if err != nil {
			return "", err
		}
		root = status
		path = strings.TrimPrefix(expr, addSourceStatus)
	case strings.HasPrefix(expr, addSourceProviderConfig):
		if pcObj == nil {
			return "", errors.New("no ProviderConfig available for this resource")
		}
		root = pcObj
		path = strings.TrimPrefix(expr, addSourceProviderConfig)
	default:
		return "", errors.Errorf("unsupported source (expected a %q or %q prefix)", addSourceStatus, addSourceProviderConfig)
	}
	if path == "" {
		return "", errors.New("empty field path")
	}

	val, found, err := unstructured.NestedFieldNoCopy(root, strings.Split(path, ".")...)
	if err != nil {
		return "", err
	}
	if !found {
		return "", errors.Errorf("field %q not found", path)
	}
	switch v := val.(type) {
	case string:
		return v, nil
	case bool, int64, float64:
		return fmt.Sprint(v), nil
	default:
		return "", errors.Errorf("field %q is not a scalar value", path)
	}
}

// mergeAdded adds the resolved fields into data in place, skipping (and
// returning, for reporting) any key that data already holds. A field to add
// therefore never silently overwrites a value copied or renamed from the
// source secret.
func mergeAdded(data map[string][]byte, added map[string][]byte) []string {
	keys := make([]string, 0, len(added))
	for k := range added {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var conflicts []string
	for _, k := range keys {
		if _, exists := data[k]; exists {
			conflicts = append(conflicts, k)
			continue
		}
		data[k] = added[k]
	}
	return conflicts
}

// getOwner reads the managed resource owning the connection secret. Only the
// namespaced API group (*.keycloak.m.crossplane.io) holds namespaced managed
// resources; the cluster-scoped ones are read without a namespace.
func (r *Reconciler) getOwner(ctx context.Context, owner metav1.OwnerReference, namespace string) (*unstructured.Unstructured, error) {
	gv, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse the owner's apiVersion")
	}

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

// transformedName returns the name of the secret the transformed connection
// details are published under: the source secret's name plus a suffix, the
// value of the transform name annotation, or a matching
// ConnectionSecretTransform's spec.transformedSecretName - checked in that
// order, the CRD taking precedence. A name that is not a valid secret name,
// or that points at the source secret itself, is rejected - writing it would
// either fail forever or destroy the connection secret.
func transformedName(source *corev1.Secret, mr *unstructured.Unstructured, crd *clusterv1beta1.ConnectionSecretTransform) (string, error) {
	sourceField := AnnotationKeyTransformedName
	name := strings.TrimSpace(mr.GetAnnotations()[AnnotationKeyTransformedName])
	if crd != nil && strings.TrimSpace(crd.Spec.TransformedSecretName) != "" {
		name = strings.TrimSpace(crd.Spec.TransformedSecretName)
		sourceField = "spec.transformedSecretName"
	}
	if name == "" {
		return source.Name + transformedSecretSuffix, nil
	}
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		return "", errors.Errorf("%s: %q is not a valid secret name: %s", sourceField, name, strings.Join(errs, "; "))
	}
	if name == source.Name {
		return "", errors.Errorf("%s: %q is the connection secret itself, which must not be overwritten", sourceField, name)
	}
	return name, nil
}

// transform copies data over, renaming the configured keys. Renames that
// would collide - two keys renamed onto the same name, or a key renamed onto
// one that is copied over unchanged - are dropped instead of silently losing
// a value; the affected entries are returned for reporting. Keys are
// processed in a deterministic order, so the outcome never depends on map
// iteration order.
func transform(data map[string][]byte, rename map[string]string) (map[string][]byte, []string) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// effective holds the renames that are still applied; conflicting ones
	// are removed from it until no target name is claimed twice. Every pass
	// removes at least one entry, so this terminates.
	effective := make(map[string]string, len(rename))
	for _, k := range keys {
		if v, ok := rename[k]; ok && v != k {
			effective[k] = v
		}
	}

	var conflicts []string
	for {
		claimed := make(map[string]string, len(keys))
		var dropped []string
		for _, k := range keys {
			target := k
			if v, ok := effective[k]; ok {
				target = v
			}
			prev, taken := claimed[target]
			if !taken {
				claimed[target] = k
				continue
			}
			// The renamed key loses. If both are renamed, the one sorting
			// later loses, which keeps the outcome deterministic.
			loser := k
			if _, renamed := effective[k]; !renamed {
				loser = prev
			}
			dropped = append(dropped, loser)
		}
		if len(dropped) == 0 {
			break
		}
		sort.Strings(dropped)
		for _, k := range dropped {
			conflicts = append(conflicts, fmt.Sprintf("%q=%q", k, effective[k]))
			delete(effective, k)
		}
	}
	sort.Strings(conflicts)

	out := make(map[string][]byte, len(data))
	for _, k := range keys {
		target := k
		if v, ok := effective[k]; ok {
			target = v
		}
		out[target] = data[k]
	}
	return out, conflicts
}

// parseRenameAnnotation parses an annotation value holding a set of
// "<oldKey>=<newKey>" pairs. Two syntaxes are accepted:
//
//   - A YAML block mapping, e.g. "clientID: client-id\nclientSecret:
//     client-secret". This is tried first: if v unmarshals into a
//     map[string]string it is used as-is.
//   - A flat list of "<oldKey>=<newKey>" pairs separated by commas and/or
//     newlines (Kubernetes annotations may be multi-line YAML block
//     scalars, which is easier to read for longer lists), e.g.
//     "clientID=client-id,clientSecret=client-secret" or one pair per line.
//     This is the fallback used whenever v is not a YAML mapping.
//
// Malformed entries are ignored rather than failing the whole reconcile,
// which would otherwise leave a resource stuck on a typo. The YAML mapping
// form is all-or-nothing (a malformed YAML document simply fails to parse
// as a mapping and falls back to the flat-list parser), so per-entry
// tolerance is preserved either way.
func parseRenameAnnotation(v string) map[string]string {
	if v == "" {
		return nil
	}
	if m, ok := parseRenameAnnotationYAML(v); ok {
		return m
	}
	rename := map[string]string{}
	for _, pair := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == '\n' }) {
		oldKey, newKey, found := strings.Cut(strings.TrimSpace(pair), "=")
		oldKey, newKey = strings.TrimSpace(oldKey), strings.TrimSpace(newKey)
		if !found || oldKey == "" || newKey == "" {
			continue
		}
		rename[oldKey] = newKey
	}
	return rename
}

// parseRenameAnnotationYAML tries to interpret v as a YAML block mapping of
// "<oldKey>: <newKey>" pairs. It returns ok == false whenever v does not
// unmarshal cleanly into a map[string]string (in particular, a "key=value"
// flat list parses as a YAML scalar rather than a mapping, so it never
// matches here and safely falls back to the flat-list parser), or whenever
// it parses to an empty map, so that callers can fall back to the
// comma/newline "key=value" syntax without ambiguity.
func parseRenameAnnotationYAML(v string) (map[string]string, bool) {
	m := map[string]string{}
	if err := yaml.Unmarshal([]byte(v), &m); err != nil || len(m) == 0 {
		return nil, false
	}
	rename := make(map[string]string, len(m))
	for oldKey, newKey := range m {
		oldKey, newKey = strings.TrimSpace(oldKey), strings.TrimSpace(newKey)
		if oldKey == "" || newKey == "" {
			continue
		}
		rename[oldKey] = newKey
	}
	if len(rename) == 0 {
		return nil, false
	}
	return rename, true
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
