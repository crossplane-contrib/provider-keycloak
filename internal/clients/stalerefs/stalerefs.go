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
// MaybeRecover detects this state from the Synced condition and clears every
// reference-resolved value field on spec.forProvider (and spec.initProvider),
// so that the next reconcile re-resolves them with the new UUIDs.
package stalerefs

import (
	"context"
	"reflect"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// RecoveryAtGenerationAnnotation records the generation at which we last
// cleared stale references on a resource. MaybeRecover refuses to clear
// twice within the same generation; once the user (or another controller)
// bumps the spec, recovery becomes eligible again.
const RecoveryAtGenerationAnnotation = "provider-keycloak.crossplane.io/stale-ref-recovery-at-generation"

// Managed is the minimum subset of resource.Managed that MaybeRecover needs.
// Narrowing the interface keeps the package independently testable.
type Managed interface {
	client.Object
	GetCondition(ct xpv1.ConditionType) xpv1.Condition
}

// MaybeRecover inspects mg's Synced condition. If it signals a stale-reference
// 404, MaybeRecover clears every value field on mg.Spec.ForProvider (and
// mg.Spec.InitProvider) whose sibling XRef/XRefs/XSelector is non-nil, then
// persists the update via kube. It returns true when a clearing was attempted
// (regardless of whether the update raced with another writer).
func MaybeRecover(ctx context.Context, kube client.Client, mg Managed) (bool, error) {
	if !isStaleRefCondition(mg.GetCondition(xpv1.TypeSynced)) {
		return false, nil
	}
	if alreadyRecoveredForCurrentGeneration(mg) {
		return false, nil
	}
	if !clearResolvedRefs(mg) {
		return false, nil
	}
	setRecoveryAnnotation(mg)
	if err := kube.Update(ctx, mg); err != nil {
		if apierrors.IsConflict(err) {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

// alreadyRecoveredForCurrentGeneration returns true when a prior reconcile
// already cleared stale references at the current spec generation. Without
// this guard, a referenced K8s resource that is itself stuck stale would
// cause MaybeRecover to clear, watch the runtime resolver repopulate the
// same stale UUID, clear again, ad infinitum.
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
// generation. Clearing spec.forProvider increments the spec generation by
// one, so we record the predicted post-Update value (mg.GetGeneration()+1)
// to match what a subsequent reconcile will observe.
func setRecoveryAnnotation(mg Managed) {
	a := mg.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}
	a[RecoveryAtGenerationAnnotation] = strconv.FormatInt(mg.GetGeneration()+1, 10)
	mg.SetAnnotations(a)
}

// isStaleRefCondition reports whether c looks like a 404 from
// terraform-provider-keycloak. The terraform-provider-keycloak HTTP client
// returns a *keycloak.ApiError whose Error() includes both the "ApiError"
// type name and the "404" status code, so either substring is a reliable
// hit when wrapped through upjet into the Synced condition message.
func isStaleRefCondition(c xpv1.Condition) bool {
	if c.Status != corev1.ConditionFalse {
		return false
	}
	return strings.Contains(c.Message, "404") || strings.Contains(c.Message, "ApiError")
}

// clearResolvedRefs walks Spec.ForProvider and Spec.InitProvider; for each
// suffix-paired field (X / XRef, X / XRefs, X / XSelector) whose ref or
// selector sibling is non-nil, it zeroes the value field X. Returns true
// if at least one value was cleared.
func clearResolvedRefs(mg Managed) bool {
	v := reflect.ValueOf(mg)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	spec := v.FieldByName("Spec")
	if !spec.IsValid() || spec.Kind() != reflect.Struct {
		return false
	}
	cleared := false
	for _, name := range []string{"ForProvider", "InitProvider"} {
		params := spec.FieldByName(name)
		if !params.IsValid() {
			continue
		}
		if clearRefsInStruct(params) {
			cleared = true
		}
	}
	return cleared
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
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.Len() == 0
	default:
		return false
	}
}
