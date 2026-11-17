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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// RecoveryAtGenerationAnnotation records the generation at which we last
// cleared stale references on a resource. MaybeRecover refuses to clear
// twice within the same generation; once the user (or another controller)
// bumps the spec, recovery becomes eligible again.
const RecoveryAtGenerationAnnotation = "provider-keycloak.crossplane.io/stale-ref-recovery-at-generation"

// RecoveredValuesAnnotation records a fingerprint of the reference-resolved
// values the last recovery cleared. MaybeRecover does not clear the same
// values twice: if the resolver puts back exactly what was cleared, the
// references were not stale.
const RecoveredValuesAnnotation = "provider-keycloak.crossplane.io/stale-ref-recovery-values"

// Managed is the minimum subset of resource.Managed that MaybeRecover needs.
// Narrowing the interface keeps the package independently testable.
type Managed interface {
	client.Object
	GetCondition(ct xpv1.ConditionType) xpv1.Condition
}

// MaybeRecover inspects mg's Synced condition. If it signals a stale-reference
// 404 AND the resource was previously created in Keycloak, MaybeRecover clears
// every value field on mg.Spec.ForProvider (and mg.Spec.InitProvider) whose
// sibling XRef/XRefs/XSelector is non-nil, then persists the update via kube.
// It returns true when a clearing was attempted (regardless of whether the
// update raced with another writer).
func MaybeRecover(ctx context.Context, kube client.Client, mg Managed) (bool, error) {
	// The reference resolver does not run for deleted resources, so values
	// cleared now would never come back and the delete could never finish.
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
	fingerprint := resolvedValuesFingerprint(mg)
	if alreadyRecoveredForValues(mg, fingerprint) {
		return false, nil
	}
	if !clearResolvedRefs(mg) {
		return false, nil
	}
	setRecoveryAnnotations(mg, fingerprint)
	if err := kube.Update(ctx, mg); err != nil {
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
// operation starts, not when Keycloak accepted the object. A failed async
// create is recorded later as external-create-failed. A failure newer than
// the last "success" therefore means the resource was never created.
func wasCreatedExternally(mg Managed) bool {
	succeeded := meta.GetExternalCreateSucceeded(mg)
	if succeeded.IsZero() {
		return false
	}
	failed := meta.GetExternalCreateFailed(mg)
	return failed.IsZero() || succeeded.After(failed)
}

// alreadyRecoveredForCurrentGeneration returns true when a prior reconcile
// already cleared stale references at the current spec generation, i.e. the
// resolver has not written the values back yet. It does not stop a loop on
// its own: once the resolver repopulates the values, the generation moves
// past the recorded one. alreadyRecoveredForValues covers that case.
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

// alreadyRecoveredForValues returns true when the last recovery cleared
// exactly the values the resolver has put back since. The generation guard
// alone cannot stop a loop: the resolver writes the re-resolved values into
// spec.forProvider, which bumps the generation past the recorded one. If the
// resolver yields the same values again, the references were not stale and
// clearing them again cannot help.
func alreadyRecoveredForValues(mg Managed, fingerprint string) bool {
	if fingerprint == "" {
		return false
	}
	return mg.GetAnnotations()[RecoveredValuesAnnotation] == fingerprint
}

// setRecoveryAnnotations marks the resource as recovered at its post-Update
// generation and records a fingerprint of the values that were cleared.
// Clearing spec.forProvider increments the spec generation by one, so we
// record the predicted post-Update value (mg.GetGeneration()+1) to match what
// a subsequent reconcile will observe.
func setRecoveryAnnotations(mg Managed, fingerprint string) {
	a := mg.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}
	a[RecoveryAtGenerationAnnotation] = strconv.FormatInt(mg.GetGeneration()+1, 10)
	if fingerprint != "" {
		a[RecoveredValuesAnnotation] = fingerprint
	}
	mg.SetAnnotations(a)
}

// notFoundStatus matches a standalone 404 status code, e.g. the "404 Not Found"
// that terraform-provider-keycloak puts into its ApiError message ("error
// sending GET request to <path>: 404 Not Found. Response body: ..."). The
// digits must not be part of a longer token: the message also carries the
// request path, whose UUIDs regularly contain "404" (e.g. ".../mappers/
// 0a65e744-e648-4040-b257-ed2492a33d35"), and a 401 or 500 for such a path
// must not be taken for a stale reference.
var notFoundStatus = regexp.MustCompile(`(^|[^0-9A-Za-z-])404([^0-9A-Za-z-]|$)`)

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

// resolvedValuesFingerprint returns a short hash over every value field that
// clearResolvedRefs would clear (same pairing rules), or "" if there is none.
func resolvedValuesFingerprint(mg Managed) string {
	var entries []string
	for _, params := range parameterStructs(mg) {
		t := params.value.Type()
		seen := map[string]bool{}
		for i := 0; i < params.value.NumField(); i++ {
			valueName, ok := valueFieldName(t.Field(i).Name)
			if !ok || seen[valueName] || isNilOrEmpty(params.value.Field(i)) {
				continue
			}
			value := params.value.FieldByName(valueName)
			if !value.IsValid() || isNilOrEmpty(value) {
				continue
			}
			seen[valueName] = true
			entries = append(entries, params.name+"."+valueName+"="+describeValue(value))
		}
	}
	if len(entries) == 0 {
		return ""
	}
	sort.Strings(entries)
	sum := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(sum[:8])
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
