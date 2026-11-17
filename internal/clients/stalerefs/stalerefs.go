// Package stalerefs recovers from stale cross-resource references.
//
// When a Keycloak resource managed by this provider references another by
// UUID (e.g. RoleMapper.spec.forProvider.roleId resolved from roleIdRef),
// and the referenced Keycloak object is deleted and recreated out-of-band,
// the stored UUID becomes stale. Every subsequent reconcile fails with a
// keycloak.ApiError{Code:404}, and the crossplane-runtime reference resolver
// will not retry because the value field is non-empty (default resolve
// policy: IsNoOp() returns true when the value is set).
//
// MaybeRecover detects this state from the Synced condition and re-resolves
// every reference-resolved value field on spec.forProvider (and
// spec.initProvider) in memory: it clears them on a copy of the resource,
// runs the resource's generated reference resolver on that copy and persists
// the result only if every field got a value again and at least one changed.
// Cleared values are never persisted, so a delete arriving at any point
// cannot find a resource without its required parameters, and a reference
// that resolves to the same value causes no write at all.
package stalerefs

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// RecoveryAtGenerationAnnotation records the generation at which we last
// replaced stale references on a resource. MaybeRecover refuses to recover
// twice within the same generation; once the user (or another controller)
// bumps the spec, recovery becomes eligible again.
const RecoveryAtGenerationAnnotation = "provider-keycloak.crossplane.io/stale-ref-recovery-at-generation"

// Upjet's async client records a failed create only in this condition (see
// upjet pkg/resource.LastAsyncOperationCondition); it does not set
// crossplane.io/external-create-failed.
const (
	typeLastAsyncOperation   xpv1.ConditionType   = "LastAsyncOperation"
	reasonAsyncCreateFailure xpv1.ConditionReason = "AsyncCreateFailure"
)

// Managed is the minimum subset of resource.Managed that MaybeRecover needs.
// Narrowing the interface keeps the package independently testable.
type Managed interface {
	client.Object
	GetCondition(ct xpv1.ConditionType) xpv1.Condition
}

// referenceResolver is implemented by every generated managed resource that
// has references (zz_generated.resolvers.go).
type referenceResolver interface {
	ResolveReferences(ctx context.Context, c client.Reader) error
}

// MaybeRecover inspects mg's Synced condition. If it signals a stale-reference
// 404 AND the resource was previously created in Keycloak, MaybeRecover
// re-resolves every value field on mg.Spec.ForProvider (and
// mg.Spec.InitProvider) whose sibling XRef/XRefs/XSelector is non-nil,
// ignoring the current values, on a copy of mg. It persists that copy via
// kube only when every such field resolved to a value again and at least one
// value changed. It returns true when an update was attempted (regardless of
// whether the update raced with another writer).
func MaybeRecover(ctx context.Context, kube client.Client, mg Managed) (bool, error) {
	// Never touch a resource that is being deleted: the reference resolver
	// does not run for it, and Observe/Delete need its parameters as they are.
	if meta.WasDeleted(mg) {
		return false, nil
	}
	if !isStaleRefCondition(mg.GetCondition(xpv1.TypeSynced)) {
		return false, nil
	}
	if !wasCreatedExternally(mg) {
		return false, nil
	}
	if alreadyRecoveredForCurrentGeneration(mg) {
		return false, nil
	}
	before := resolvedValues(mg)
	if len(before) == 0 {
		return false, nil
	}
	candidate, ok := mg.DeepCopyObject().(Managed)
	if !ok {
		return false, nil
	}
	resolver, ok := candidate.(referenceResolver)
	if !ok {
		return false, nil
	}
	clearResolvedRefs(candidate)
	if err := resolver.ResolveReferences(ctx, kube); err != nil {
		// A reference cannot be resolved right now (e.g. its target is not
		// ready). Keep the current values; the next reconcile tries again.
		return false, nil //nolint:nilerr // not resolvable yet is not an error of the recovery
	}
	after := resolvedValues(candidate)
	if !sameFields(before, after) || maps.Equal(before, after) {
		// A field stayed empty (never persist that), or the references
		// resolve to what the resource already has: they are not stale.
		return false, nil
	}
	setRecoveryAnnotation(candidate)
	if err := kube.Update(ctx, candidate); err != nil {
		if apierrors.IsConflict(err) {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

// wasCreatedExternally reports whether the managed resource has ever
// successfully been created in Keycloak, as recorded by crossplane-runtime
// via the external-create-succeeded annotation. A resource that has never
// been created cannot hold a stale UUID; any 404 during cold start is a
// transient parent-not-ready race and must not trigger destructive clearing.
//
// With upjet's async external client, crossplane-runtime sets
// external-create-succeeded as soon as Create returns, i.e. when the async
// operation starts, not when Keycloak accepted the object. So the annotation
// alone does not confirm creation:
//   - an observed ID (status.atProvider.id) confirms it: Keycloak returned
//     the object at least once;
//   - otherwise a failed async create (LastAsyncOperation=False,
//     AsyncCreateFailure) or an external-create-failed newer than the last
//     "success" means the resource was never created.
func wasCreatedExternally(mg Managed) bool {
	succeeded := meta.GetExternalCreateSucceeded(mg)
	if succeeded.IsZero() {
		return false
	}
	if hasObservedID(mg) {
		return true
	}
	if c := mg.GetCondition(typeLastAsyncOperation); c.Status == corev1.ConditionFalse && c.Reason == reasonAsyncCreateFailure {
		return false
	}
	failed := meta.GetExternalCreateFailed(mg)
	return failed.IsZero() || succeeded.After(failed)
}

// hasObservedID reports whether status.atProvider.id is set. Upjet fills it
// from the Terraform state once Keycloak has returned the object, and every
// generated resource observation carries it.
func hasObservedID(mg Managed) bool {
	v := reflect.ValueOf(mg)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	for _, name := range []string{"Status", "AtProvider", "ID"} {
		v = v.FieldByName(name)
		if !v.IsValid() {
			return false
		}
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return false
			}
			v = v.Elem()
		}
	}
	return v.Kind() == reflect.String && v.Len() > 0
}

// alreadyRecoveredForCurrentGeneration returns true when a prior reconcile
// already replaced stale references at the current spec generation.
func alreadyRecoveredForCurrentGeneration(mg Managed) bool {
	v, ok := mg.GetAnnotations()[RecoveryAtGenerationAnnotation]
	if !ok {
		return false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return false
	}
	return n >= mg.GetGeneration()
}

// setRecoveryAnnotation marks the resource as recovered at its post-Update
// generation. Replacing values in spec.forProvider increments the spec
// generation by one, so we record the predicted post-Update value
// (mg.GetGeneration()+1) to match what a subsequent reconcile will observe.
func setRecoveryAnnotation(mg Managed) {
	a := mg.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}
	a[RecoveryAtGenerationAnnotation] = strconv.FormatInt(mg.GetGeneration()+1, 10)
	mg.SetAnnotations(a)
}

// notFoundStatus matches the HTTP status that terraform-provider-keycloak
// puts into its ApiError message: "error sending GET request to <path>: 404
// Not Found. Response body: ...". Only that signature counts. The message
// also carries the request path, which may contain "404" anywhere: in UUIDs
// (".../mappers/0a65e744-e648-4040-b257-ed2492a33d35") or as a whole segment
// (a realm named "404"). A 401 or 500 for such a path is not a stale reference.
var notFoundStatus = regexp.MustCompile(`(^|: )404 Not Found\b`)

// isStaleRefCondition reports whether c looks like a 404 from
// terraform-provider-keycloak. keycloak.ApiError.Error() returns only its
// message (no type name), so the status code in that message is the only
// reliable signal.
func isStaleRefCondition(c xpv1.Condition) bool {
	if c.Status != corev1.ConditionFalse {
		return false
	}
	return notFoundStatus.MatchString(c.Message)
}

// clearResolvedRefs walks Spec.ForProvider and Spec.InitProvider; for each
// suffix-paired field (X / XRef, X / XRefs, X / XSelector) whose ref or
// selector sibling is non-nil, it zeroes the value field X. Returns true
// if at least one value was cleared.
func clearResolvedRefs(mg Managed) bool {
	cleared := false
	for _, params := range parameterStructs(mg) {
		if clearRefsInStruct(params.value) {
			cleared = true
		}
	}
	return cleared
}

type namedParams struct {
	name  string
	value reflect.Value
}

// parameterStructs returns Spec.ForProvider and Spec.InitProvider of mg, if
// present.
func parameterStructs(mg Managed) []namedParams {
	v := reflect.ValueOf(mg)
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	spec := v.FieldByName("Spec")
	if !spec.IsValid() || spec.Kind() != reflect.Struct {
		return nil
	}
	var out []namedParams
	for _, name := range []string{"ForProvider", "InitProvider"} {
		params := spec.FieldByName(name)
		if params.IsValid() && params.Kind() == reflect.Struct {
			out = append(out, namedParams{name: name, value: params})
		}
	}
	return out
}

// resolvedValues returns every value field that clearResolvedRefs would
// clear (same pairing rules), keyed by "<ForProvider|InitProvider>.<Field>".
func resolvedValues(mg Managed) map[string]string {
	out := map[string]string{}
	for _, params := range parameterStructs(mg) {
		t := params.value.Type()
		for i := 0; i < params.value.NumField(); i++ {
			valueName, ok := valueFieldName(t.Field(i).Name)
			if !ok || isNilOrEmpty(params.value.Field(i)) {
				continue
			}
			key := params.name + "." + valueName
			if _, seen := out[key]; seen {
				continue
			}
			value := params.value.FieldByName(valueName)
			if !value.IsValid() || isNilOrEmpty(value) {
				continue
			}
			out[key] = describeValue(value)
		}
	}
	return out
}

// sameFields reports whether a and b hold values for exactly the same fields.
func sameFields(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

// describeValue renders the reference-resolved value types upjet generates
// (*string, []*string, string, ...) as a stable string.
func describeValue(v reflect.Value) string {
	switch v.Kind() { //nolint:exhaustive // pointers and slices need unwrapping, everything else prints as is
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return "<nil>"
		}
		return describeValue(v.Elem())
	case reflect.Slice, reflect.Array:
		parts := make([]string, v.Len())
		for i := range parts {
			parts[i] = describeValue(v.Index(i))
		}
		return "[" + strings.Join(parts, ",") + "]"
	default:
		return fmt.Sprint(v.Interface())
	}
}

// refSuffixes are the field-name suffixes that mark a field as a reference
// or selector sibling. Order matters: "Refs" must precede "Ref" so the
// suffix match doesn't truncate "Refs" to "Ref" and look up the wrong value
// field name.
var refSuffixes = []string{"Selector", "Refs", "Ref"}

func clearRefsInStruct(params reflect.Value) bool {
	if params.Kind() != reflect.Struct {
		return false
	}
	t := params.Type()
	clearedValues := map[string]bool{}
	cleared := false
	for i := 0; i < params.NumField(); i++ {
		fname := t.Field(i).Name
		valueName, ok := valueFieldName(fname)
		if !ok {
			continue
		}
		if clearedValues[valueName] {
			continue
		}
		if isNilOrEmpty(params.Field(i)) {
			continue
		}
		value := params.FieldByName(valueName)
		if !value.IsValid() || !value.CanSet() || isNilOrEmpty(value) {
			continue
		}
		value.Set(reflect.Zero(value.Type()))
		clearedValues[valueName] = true
		cleared = true
	}
	return cleared
}

// valueFieldName returns the value-field name a given ref/selector field
// resolves into, e.g. "ClientIDRef" -> "ClientID", "RoleIdsRefs" -> "RoleIds",
// "ClientIDSelector" -> "ClientID". Returns ("", false) if the name isn't a
// ref/selector field.
func valueFieldName(name string) (string, bool) {
	for _, s := range refSuffixes {
		if strings.HasSuffix(name, s) {
			return strings.TrimSuffix(name, s), true
		}
	}
	return "", false
}

func isNilOrEmpty(v reflect.Value) bool {
	switch v.Kind() { //nolint:exhaustive // only the kinds upjet uses for ref-paired fields matter
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.Len() == 0
	default:
		return false
	}
}
