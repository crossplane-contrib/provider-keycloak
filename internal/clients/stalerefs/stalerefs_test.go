package stalerefs

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// createdExternally returns a metav1.ObjectMeta marked with the annotation
// crossplane-runtime sets after the first successful external Create. Tests
// that exercise stale-reference recovery must use this — without it the
// MaybeRecover gate (correctly) classifies the resource as never-created and
// skips clearing.
func createdExternally() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Annotations: map[string]string{
			meta.AnnotationKeyExternalCreateSucceeded: time.Now().Format(time.RFC3339),
		},
	}
}

// fakeParams mirrors a real upjet-generated <Resource>Parameters struct: a mix
// of singular-ref pairs, slice-ref pairs, selector-only pairs, and standalone
// fields. The reflection walker must clear only the value fields whose sibling
// ref/selector is set.
type fakeParams struct {
	// Singular ref: value + Ref + Selector
	ClientID         *string
	ClientIDRef      *xpv1.Reference
	ClientIDSelector *xpv1.Selector

	// Slice ref: value + Refs (plural) + Selector
	RoleIds         []*string
	RoleIdsRefs     []xpv1.Reference
	RoleIdsSelector *xpv1.Selector

	// Selector-only (no Ref set)
	GroupID         *string
	GroupIDSelector *xpv1.Selector

	// Standalone (no ref siblings) — must never be cleared
	Name *string

	// Has ref name pattern but neither sibling set — must never be cleared
	RealmID         *string
	RealmIDRef      *xpv1.Reference
	RealmIDSelector *xpv1.Selector
}

type fakeSpec struct {
	ForProvider  fakeParams
	InitProvider fakeParams
}

// fakeObservation mirrors the ID every upjet-generated observation carries.
type fakeObservation struct {
	ID *string
}

type fakeStatus struct {
	xpv1.ConditionedStatus
	AtProvider fakeObservation
}

type fakeManaged struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Spec   fakeSpec
	Status fakeStatus

	// resolve stands in for the generated reference resolver.
	resolve func(*fakeManaged) error
}

// ResolveReferences mirrors zz_generated.resolvers.go: it only fills fields
// that are empty (a set value is a no-op for the default resolve policy).
func (f *fakeManaged) ResolveReferences(_ context.Context, _ client.Reader) error {
	if f.resolve == nil {
		return nil
	}
	return f.resolve(f)
}

// resolveClientID returns a resolver that resolves ClientIDRef to id.
func resolveClientID(id string) func(*fakeManaged) error {
	return func(f *fakeManaged) error {
		if f.Spec.ForProvider.ClientID == nil {
			f.Spec.ForProvider.ClientID = ptr(id)
		}
		return nil
	}
}

func (f *fakeManaged) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return f.Status.GetCondition(ct)
}

func (f *fakeManaged) SetConditions(c ...xpv1.Condition) {
	f.Status.SetConditions(c...)
}

func (f *fakeManaged) GetObjectKind() schema.ObjectKind { return schema.EmptyObjectKind }

func (f *fakeManaged) DeepCopyObject() runtime.Object {
	out := *f
	out.ObjectMeta = *f.DeepCopy()
	return &out
}

func ptr(s string) *string { return &s }

func TestIsStaleRefCondition(t *testing.T) {
	cases := []struct {
		name string
		cond xpv1.Condition
		want bool
	}{
		{
			name: "false status with ApiError message but no 404 → not stale",
			cond: xpv1.Condition{Status: corev1.ConditionFalse, Message: "create failed: keycloak.ApiError: not found"},
			want: false,
		},
		{
			name: "real terraform-provider-keycloak 404 → stale",
			cond: xpv1.Condition{Status: corev1.ConditionFalse, Message: `connect failed: cannot get ID: error sending GET request to /admin/realms/infoma/clients/71d068fb-bc7c-4e48-ad02-bfac716b8076: 404 Not Found. Response body: {"error":"Could not find client"}`},
			want: true,
		},
		{
			name: "401 for a path whose UUID contains 404 → not stale",
			cond: xpv1.Condition{Status: corev1.ConditionFalse, Message: `connect failed: cannot get ID: error sending GET request to /admin/realms/infoma/identity-provider/instances/acme/mappers/0a65e744-e648-4040-b257-ed2492a33d35: 401 Unauthorized. Response body: {"error":"HTTP 401 Unauthorized"}`},
			want: false,
		},
		{
			name: "401 for a path with a segment named 404 → not stale",
			cond: xpv1.Condition{Status: corev1.ConditionFalse, Message: "error sending GET request to /admin/realms/404/clients/a: 401 Unauthorized."},
			want: false,
		},
		{
			name: "lower-case or bare 404 without the status text → not stale",
			cond: xpv1.Condition{Status: corev1.ConditionFalse, Message: "observe failed: keycloak.ApiError: 404 not found"},
			want: false,
		},
		{
			name: "404 inside a hex token only → not stale",
			cond: xpv1.Condition{Status: corev1.ConditionFalse, Message: "error sending GET request to /admin/realms/r/clients/a404b: 500 Internal Server Error."},
			want: false,
		},
		{
			name: "false status with 404 message → stale",
			cond: xpv1.Condition{Status: corev1.ConditionFalse, Message: "observe failed: 404 Not Found from /admin/realms"},
			want: true,
		},
		{
			name: "false status with unrelated message → not stale",
			cond: xpv1.Condition{Status: corev1.ConditionFalse, Message: "connection refused"},
			want: false,
		},
		{
			name: "true status (synced) → not stale",
			cond: xpv1.Condition{Status: corev1.ConditionTrue, Message: ""},
			want: false,
		},
		{
			name: "unknown status with 404 → not stale (we only fire on definite failure)",
			cond: xpv1.Condition{Status: corev1.ConditionUnknown, Message: "404"},
			want: false,
		},
		{
			name: "empty condition → not stale",
			cond: xpv1.Condition{},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isStaleRefCondition(tc.cond); got != tc.want {
				t.Fatalf("isStaleRefCondition() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValueFieldName(t *testing.T) {
	cases := []struct {
		in        string
		wantName  string
		wantMatch bool
	}{
		{"ClientIDRef", "ClientID", true},
		{"ClientIDSelector", "ClientID", true},
		{"RoleIdsRefs", "RoleIds", true},
		{"RoleIdsSelector", "RoleIds", true},
		{"Standalone", "", false},
		{"Selector", "", true}, // edge: "Selector" itself trims to ""
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, ok := valueFieldName(c.in)
			if ok != c.wantMatch || got != c.wantName {
				t.Fatalf("valueFieldName(%q) = (%q,%v), want (%q,%v)", c.in, got, ok, c.wantName, c.wantMatch)
			}
		})
	}
}

func TestClearResolvedRefs(t *testing.T) {
	cases := []struct {
		name        string
		mg          *fakeManaged
		wantCleared bool
		check       func(t *testing.T, mg *fakeManaged)
	}{
		{
			name: "singular value+ref → cleared",
			mg: &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("stale-uuid"),
				ClientIDRef: &xpv1.Reference{Name: "the-client"},
			}}},
			wantCleared: true,
			check: func(t *testing.T, mg *fakeManaged) {
				if mg.Spec.ForProvider.ClientID != nil {
					t.Errorf("ClientID not cleared: %v", *mg.Spec.ForProvider.ClientID)
				}
				if mg.Spec.ForProvider.ClientIDRef == nil {
					t.Errorf("ClientIDRef was cleared but should be preserved")
				}
			},
		},
		{
			name: "singular value+selector → cleared",
			mg: &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
				GroupID:         ptr("stale-uuid"),
				GroupIDSelector: &xpv1.Selector{MatchLabels: map[string]string{"app": "x"}},
			}}},
			wantCleared: true,
			check: func(t *testing.T, mg *fakeManaged) {
				if mg.Spec.ForProvider.GroupID != nil {
					t.Errorf("GroupID not cleared: %v", *mg.Spec.ForProvider.GroupID)
				}
				if mg.Spec.ForProvider.GroupIDSelector == nil {
					t.Errorf("GroupIDSelector was cleared but should be preserved")
				}
			},
		},
		{
			name: "slice value+refs → cleared",
			mg: &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
				RoleIds:     []*string{ptr("uuid-a"), ptr("uuid-b")},
				RoleIdsRefs: []xpv1.Reference{{Name: "role-a"}, {Name: "role-b"}},
			}}},
			wantCleared: true,
			check: func(t *testing.T, mg *fakeManaged) {
				if len(mg.Spec.ForProvider.RoleIds) != 0 {
					t.Errorf("RoleIds not cleared: %v", mg.Spec.ForProvider.RoleIds)
				}
				if len(mg.Spec.ForProvider.RoleIdsRefs) == 0 {
					t.Errorf("RoleIdsRefs was cleared but should be preserved")
				}
			},
		},
		{
			name: "value present but no ref/selector → untouched (manual UUID)",
			mg: &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
				RealmID: ptr("manually-set-uuid"),
			}}},
			wantCleared: false,
			check: func(t *testing.T, mg *fakeManaged) {
				if mg.Spec.ForProvider.RealmID == nil || *mg.Spec.ForProvider.RealmID != "manually-set-uuid" {
					t.Errorf("RealmID was incorrectly cleared")
				}
			},
		},
		{
			name: "standalone field → untouched",
			mg: &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
				Name: ptr("my-rolemapper"),
			}}},
			wantCleared: false,
		},
		{
			name: "value already empty → no clearing needed",
			mg: &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
				ClientIDRef: &xpv1.Reference{Name: "x"},
			}}},
			wantCleared: false,
		},
		{
			name: "InitProvider also cleared",
			mg: &fakeManaged{Spec: fakeSpec{
				ForProvider: fakeParams{
					ClientID:    ptr("stale-1"),
					ClientIDRef: &xpv1.Reference{Name: "c"},
				},
				InitProvider: fakeParams{
					ClientID:    ptr("stale-2"),
					ClientIDRef: &xpv1.Reference{Name: "c"},
				},
			}},
			wantCleared: true,
			check: func(t *testing.T, mg *fakeManaged) {
				if mg.Spec.ForProvider.ClientID != nil {
					t.Errorf("ForProvider.ClientID not cleared")
				}
				if mg.Spec.InitProvider.ClientID != nil {
					t.Errorf("InitProvider.ClientID not cleared")
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := clearResolvedRefs(tc.mg)
			if got != tc.wantCleared {
				t.Fatalf("clearResolvedRefs() = %v, want %v", got, tc.wantCleared)
			}
			if tc.check != nil {
				tc.check(t, tc.mg)
			}
		})
	}
}

func TestMaybeRecover(t *testing.T) {
	staleCond := xpv1.Condition{
		Type:    xpv1.TypeSynced,
		Status:  corev1.ConditionFalse,
		Reason:  xpv1.ReasonReconcileError,
		Message: "connect failed: cannot get ID: error sending GET request to /admin/realms/r/clients/x: 404 Not Found. Response body: {}",
	}
	healthyCond := xpv1.Condition{
		Type:   xpv1.TypeSynced,
		Status: corev1.ConditionTrue,
		Reason: xpv1.ReasonReconcileSuccess,
	}
	// upjet's async callback records a failed create only in the
	// LastAsyncOperation condition; external-create-failed is not set.
	asyncCreateFailed := xpv1.Condition{
		Type:    typeLastAsyncOperation,
		Status:  corev1.ConditionFalse,
		Reason:  reasonAsyncCreateFailure,
		Message: "async create failed: failed to create the resource: [{0 error sending POST request to /admin/realms/r/identity-provider/instances/acme/mappers: 404 Not Found. Response body: {}  []}]",
	}

	// stale returns a created resource holding oldID whose resolver now
	// yields newID.
	stale := func(om metav1.ObjectMeta, oldID, newID string) *fakeManaged {
		mg := &fakeManaged{
			ObjectMeta: om,
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr(oldID),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
			resolve: resolveClientID(newID),
		}
		mg.SetConditions(staleCond)
		return mg
	}

	// recorder captures what MaybeRecover writes.
	type recorder struct {
		updates   int
		persisted *fakeManaged
	}
	recording := func(r *recorder, err error) *test.MockClient {
		return &test.MockClient{
			MockUpdate: func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
				r.updates++
				r.persisted = obj.(*fakeManaged)
				return err
			},
		}
	}

	expect := func(t *testing.T, mg *fakeManaged, kube client.Client, wantRecovered bool) {
		t.Helper()
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered != wantRecovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (%v,nil)", recovered, err, wantRecovered)
		}
	}

	t.Run("healthy: no-op", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		mg.SetConditions(healthyCond)
		var r recorder
		expect(t, mg, recording(&r, nil), false)
		if r.updates != 0 {
			t.Errorf("kube.Update was called on a healthy resource")
		}
	})

	t.Run("stale: persists the re-resolved value, never an empty one", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		var r recorder
		expect(t, mg, recording(&r, nil), true)
		if r.updates != 1 || r.persisted == nil {
			t.Fatalf("kube.Update called %d times, want 1", r.updates)
		}
		if got := r.persisted.Spec.ForProvider.ClientID; got == nil || *got != "new-uuid" {
			t.Errorf("persisted ClientID = %v, want new-uuid", got)
		}
		if mg.Spec.ForProvider.ClientID == nil || *mg.Spec.ForProvider.ClientID != "old-uuid" {
			t.Errorf("the caller's object was modified; only the copy may change")
		}
	})

	// Copilot review on #785: with clearing persisted first, a delete that
	// arrived before the next reconcile left the resource without its
	// required parameters forever. Nothing empty is ever persisted now.
	t.Run("delete right after a recovery finds complete parameters", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		var r recorder
		expect(t, mg, recording(&r, nil), true)

		deleted := r.persisted
		now := metav1.Now()
		deleted.DeletionTimestamp = &now
		if deleted.Spec.ForProvider.ClientID == nil {
			t.Fatalf("the persisted object has no ClientID; its delete would hang")
		}
		expect(t, deleted, recording(&recorder{}, nil), false)
	})

	t.Run("references resolve to the same value → no write", func(t *testing.T) {
		mg := stale(createdExternally(), "same-uuid", "same-uuid")
		var r recorder
		expect(t, mg, recording(&r, nil), false)
		if r.updates != 0 {
			t.Errorf("kube.Update called %d times for unchanged values", r.updates)
		}
	})

	// dev/test/stale-ref-recovery.sh: a Role is recreated out-of-band. Until
	// the Role MR has observed the new UUID, the resolver yields the old one:
	// no write. Once it has, the next reconcile writes the new UUID.
	t.Run("referenced MR catches up later → recovered then, no churn before", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "old-uuid")
		var r recorder
		kube := recording(&r, nil)
		for range 5 {
			expect(t, mg, kube, false)
		}
		mg.resolve = resolveClientID("new-uuid")
		expect(t, mg, kube, true)
		if r.updates != 1 {
			t.Errorf("kube.Update called %d times, want 1", r.updates)
		}
	})

	t.Run("resolver fails → keep current values, no write", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		mg.resolve = func(*fakeManaged) error { return errors.New("referenced resource not ready") }
		var r recorder
		expect(t, mg, recording(&r, nil), false)
		if r.updates != 0 {
			t.Errorf("kube.Update was called although resolution failed")
		}
	})

	t.Run("resolver leaves a field empty → no write", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		mg.resolve = nil
		var r recorder
		expect(t, mg, recording(&r, nil), false)
		if r.updates != 0 {
			t.Errorf("kube.Update was called with an unresolved field")
		}
	})

	t.Run("no reference-resolved fields → no-op", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: createdExternally(),
			Spec:       fakeSpec{ForProvider: fakeParams{Name: ptr("no-refs-here")}},
		}
		mg.SetConditions(staleCond)
		var r recorder
		expect(t, mg, recording(&r, nil), false)
		if r.updates != 0 {
			t.Errorf("kube.Update was called when nothing was resolvable")
		}
	})

	t.Run("update conflict → treated as benign", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		conflict := apierrors.NewConflict(schema.GroupResource{Group: "x", Resource: "y"}, "name", errors.New("stale"))
		expect(t, mg, recording(&recorder{}, conflict), true)
	})

	t.Run("update non-conflict error → returned", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		wantErr := errors.New("nope")
		recovered, err := MaybeRecover(context.Background(), recording(&recorder{}, wantErr), mg)
		if !errors.Is(err, wantErr) || recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (false,%v)", recovered, err, wantErr)
		}
	})

	t.Run("annotation matches current generation → skip", func(t *testing.T) {
		om := createdExternally()
		om.Generation = 5
		om.Annotations[RecoveryAtGenerationAnnotation] = "5"
		var r recorder
		expect(t, stale(om, "old-uuid", "new-uuid"), recording(&r, nil), false)
		if r.updates != 0 {
			t.Errorf("kube.Update was called despite the generation guard")
		}
	})

	t.Run("annotation older than current generation → recover, annotation gen+1", func(t *testing.T) {
		om := createdExternally()
		om.Generation = 7
		om.Annotations[RecoveryAtGenerationAnnotation] = "5"
		var r recorder
		expect(t, stale(om, "old-uuid", "new-uuid"), recording(&r, nil), true)
		if got := r.persisted.GetAnnotations()[RecoveryAtGenerationAnnotation]; got != strconv.FormatInt(8, 10) {
			t.Errorf("annotation = %q, want 8 (current gen + 1)", got)
		}
	})

	t.Run("annotation malformed → recover (defensive)", func(t *testing.T) {
		om := createdExternally()
		om.Generation = 1
		om.Annotations[RecoveryAtGenerationAnnotation] = "not-a-number"
		expect(t, stale(om, "old-uuid", "new-uuid"), recording(&recorder{}, nil), true)
	})

	t.Run("being deleted → skip", func(t *testing.T) {
		om := createdExternally()
		now := metav1.Now()
		om.DeletionTimestamp = &now
		var r recorder
		expect(t, stale(om, "old-uuid", "new-uuid"), recording(&r, nil), false)
		if r.updates != 0 {
			t.Errorf("kube.Update was called on a deleted resource")
		}
	})

	// Regression: a never-created resource that takes a 404 during cold start
	// (e.g. parent realm not yet in Keycloak) must be left alone.
	t.Run("never created externally → skip", func(t *testing.T) {
		var r recorder
		expect(t, stale(metav1.ObjectMeta{}, "old-uuid", "new-uuid"), recording(&r, nil), false)
		if r.updates != 0 {
			t.Errorf("kube.Update was called on a never-created resource")
		}
	})

	t.Run("external-create-failed newer than the start annotation → never created, skip", func(t *testing.T) {
		started := time.Now().Add(-30 * time.Second)
		om := metav1.ObjectMeta{Annotations: map[string]string{
			meta.AnnotationKeyExternalCreateSucceeded: started.Format(time.RFC3339),
			meta.AnnotationKeyExternalCreateFailed:    started.Add(29 * time.Second).Format(time.RFC3339),
		}}
		expect(t, stale(om, "old-uuid", "new-uuid"), recording(&recorder{}, nil), false)
	})

	t.Run("create succeeded after an earlier failure → created, recover", func(t *testing.T) {
		failed := time.Now().Add(-time.Hour)
		om := metav1.ObjectMeta{Annotations: map[string]string{
			meta.AnnotationKeyExternalCreateFailed:    failed.Format(time.RFC3339),
			meta.AnnotationKeyExternalCreateSucceeded: failed.Add(time.Minute).Format(time.RFC3339),
		}}
		expect(t, stale(om, "old-uuid", "new-uuid"), recording(&recorder{}, nil), true)
	})

	t.Run("async create failed, only the start annotation set → never created, skip", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		mg.SetConditions(asyncCreateFailed)
		expect(t, mg, recording(&recorder{}, nil), false)
	})

	// A resource Keycloak returned before (observed ID) that fails to be
	// re-created because a referenced UUID is stale is exactly the case
	// recovery exists for.
	t.Run("async create failed but an ID was observed before → created, recover", func(t *testing.T) {
		mg := stale(createdExternally(), "old-uuid", "new-uuid")
		mg.Status.AtProvider.ID = ptr("3357aaeb-6461-40cc-96e6-7e0b6d3a02e5")
		mg.SetConditions(asyncCreateFailed)
		expect(t, mg, recording(&recorder{}, nil), true)
	})
}

// Compile-time check: ensure fakeManaged satisfies the narrow Managed
// interface used by MaybeRecover.
var _ Managed = (*fakeManaged)(nil)

func TestResolvedValues(t *testing.T) {
	mg := &fakeManaged{Spec: fakeSpec{
		ForProvider: fakeParams{
			ClientID:    ptr("u1"),
			ClientIDRef: &xpv1.Reference{Name: "c"},
			RoleIds:     []*string{ptr("r1"), ptr("r2")},
			RoleIdsRefs: []xpv1.Reference{{Name: "r"}},
			Name:        ptr("not-a-ref"),
			RealmID:     ptr("manual-value-without-ref"),
		},
		InitProvider: fakeParams{
			GroupID:         ptr("g1"),
			GroupIDSelector: &xpv1.Selector{},
		},
	}}
	want := map[string]string{
		"ForProvider.ClientID": "u1",
		"ForProvider.RoleIds":  "[r1,r2]",
		"InitProvider.GroupID": "g1",
	}
	got := resolvedValues(mg)
	if len(got) != len(want) {
		t.Fatalf("resolvedValues() = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("resolvedValues()[%q] = %q, want %q", k, got[k], v)
		}
	}
	if !sameFields(got, want) || sameFields(got, map[string]string{"ForProvider.ClientID": "u1"}) {
		t.Errorf("sameFields() does not compare the field sets")
	}
}
