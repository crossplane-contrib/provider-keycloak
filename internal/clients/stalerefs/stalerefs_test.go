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

	t.Run("healthy: no-op", func(t *testing.T) {
		mg := &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
			ClientID:    ptr("uuid"),
			ClientIDRef: &xpv1.Reference{Name: "c"},
		}}}
		mg.SetConditions(healthyCond)

		updateCalled := false
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				updateCalled = true
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (false,nil)", recovered, err)
		}
		if updateCalled {
			t.Errorf("kube.Update was called on a healthy resource")
		}
		if mg.Spec.ForProvider.ClientID == nil {
			t.Errorf("ClientID cleared on a healthy resource")
		}
	})

	t.Run("stale: clears and persists", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: createdExternally(),
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		var persisted *fakeManaged
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
				persisted = obj.(*fakeManaged)
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || !recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}
		if persisted == nil {
			t.Fatalf("kube.Update was not called")
		}
		if persisted.Spec.ForProvider.ClientID != nil {
			t.Errorf("persisted ClientID was not cleared: %v", *persisted.Spec.ForProvider.ClientID)
		}
	})

	t.Run("stale but nothing to clear → no-op", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: createdExternally(),
			Spec: fakeSpec{ForProvider: fakeParams{
				Name: ptr("no-refs-here"),
			}},
		}
		mg.SetConditions(staleCond)

		updateCalled := false
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				updateCalled = true
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (false,nil)", recovered, err)
		}
		if updateCalled {
			t.Errorf("kube.Update was called when nothing was cleared")
		}
	})

	t.Run("update conflict → treated as benign", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: createdExternally(),
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				return apierrors.NewConflict(schema.GroupResource{Group: "x", Resource: "y"}, "name", errors.New("stale"))
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil {
			t.Fatalf("MaybeRecover() returned unexpected error: %v", err)
		}
		if !recovered {
			t.Fatalf("expected recovered=true even on conflict, got false")
		}
	})

	t.Run("annotation matches current generation → skip", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 5,
				Annotations: map[string]string{
					RecoveryAtGenerationAnnotation:            "5",
					meta.AnnotationKeyExternalCreateSucceeded: time.Now().Format(time.RFC3339),
				},
			},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		updateCalled := false
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				updateCalled = true
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (false,nil) — already recovered this gen", recovered, err)
		}
		if updateCalled {
			t.Errorf("kube.Update was called despite annotation guard")
		}
		if mg.Spec.ForProvider.ClientID == nil {
			t.Errorf("ClientID was cleared despite annotation guard")
		}
	})

	t.Run("annotation older than current generation → recover again", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 7,
				Annotations: map[string]string{
					RecoveryAtGenerationAnnotation:            "5",
					meta.AnnotationKeyExternalCreateSucceeded: time.Now().Format(time.RFC3339),
				},
			},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		var persistedAnno string
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
				persistedAnno = obj.GetAnnotations()[RecoveryAtGenerationAnnotation]
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || !recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}
		if persistedAnno != "8" {
			t.Errorf("annotation = %q, want %q (current gen + 1)", persistedAnno, "8")
		}
	})

	t.Run("annotation malformed → recover (defensive)", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 1,
				Annotations: map[string]string{
					RecoveryAtGenerationAnnotation:            "not-a-number",
					meta.AnnotationKeyExternalCreateSucceeded: time.Now().Format(time.RFC3339),
				},
			},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error { return nil },
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || !recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (true,nil) — malformed annotation should not block recovery", recovered, err)
		}
	})

	t.Run("first recovery sets annotation to gen+1", func(t *testing.T) {
		const gen int64 = 3
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{
				Generation: gen,
				Annotations: map[string]string{
					meta.AnnotationKeyExternalCreateSucceeded: time.Now().Format(time.RFC3339),
				},
			},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		var persisted *fakeManaged
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
				persisted = obj.(*fakeManaged)
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || !recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}
		got := persisted.GetAnnotations()[RecoveryAtGenerationAnnotation]
		want := strconv.FormatInt(gen+1, 10)
		if got != want {
			t.Errorf("annotation = %q, want %q", got, want)
		}
	})

	// Regression: a never-created resource that takes a 404 during cold start
	// (e.g. parent realm not yet in Keycloak) must NOT have its sibling-XRef
	// value fields cleared. Clearing them strands fields like serviceAccountUserId
	// — whose resolver pulls from a sibling resource's status — empty, and the
	// per-generation guard then locks recovery off, leaving the resource stuck
	// for the full reconcile budget.
	t.Run("never created externally → skip even with stale-looking condition", func(t *testing.T) {
		mg := &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
			ClientID:    ptr("uuid"),
			ClientIDRef: &xpv1.Reference{Name: "c"},
		}}}
		mg.SetConditions(staleCond)

		updateCalled := false
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				updateCalled = true
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (false,nil) — cold-start 404 must not trigger clearing", recovered, err)
		}
		if updateCalled {
			t.Errorf("kube.Update was called on a never-created resource")
		}
		if mg.Spec.ForProvider.ClientID == nil {
			t.Errorf("ClientID cleared on a never-created resource")
		}
		if _, ok := mg.GetAnnotations()[RecoveryAtGenerationAnnotation]; ok {
			t.Errorf("recovery annotation was set despite cold-start gate")
		}
	})

	// The reference resolver skips deleted resources, so values cleared during
	// a delete never come back and the delete hangs on "required param ...
	// not set".
	t.Run("being deleted → skip", func(t *testing.T) {
		now := metav1.Now()
		om := createdExternally()
		om.DeletionTimestamp = &now
		mg := &fakeManaged{
			ObjectMeta: om,
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		updateCalled := false
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				updateCalled = true
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (false,nil) for a deleted resource", recovered, err)
		}
		if updateCalled || mg.Spec.ForProvider.ClientID == nil {
			t.Errorf("references were cleared on a deleted resource")
		}
	})

	// upjet's async client: external-create-succeeded is set when the async
	// create starts; the 404 of a parent that does not exist yet arrives later
	// as external-create-failed.
	t.Run("async create failed after it started → never created, skip", func(t *testing.T) {
		started := time.Now().Add(-30 * time.Second)
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				meta.AnnotationKeyExternalCreateSucceeded: started.Format(time.RFC3339),
				meta.AnnotationKeyExternalCreateFailed:    started.Add(29 * time.Second).Format(time.RFC3339),
			}},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				t.Errorf("kube.Update was called for a failed async create")
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (false,nil)", recovered, err)
		}
	})

	t.Run("create succeeded after an earlier failure → created, recover", func(t *testing.T) {
		failed := time.Now().Add(-time.Hour)
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				meta.AnnotationKeyExternalCreateFailed:    failed.Format(time.RFC3339),
				meta.AnnotationKeyExternalCreateSucceeded: failed.Add(time.Minute).Format(time.RFC3339),
			}},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error { return nil },
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || !recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}
	})

	// Regression for the loop the generation guard cannot stop: recovery
	// clears (gen N -> N+1, annotation N+1), the resolver writes the same
	// value back (gen N+2), the error persists, recovery clears again, ...
	// Observed as generation 79 on a mapper during a 401 outage.
	t.Run("resolver restored the same values → do not clear again", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 4,
				Annotations: map[string]string{
					meta.AnnotationKeyExternalCreateSucceeded: time.Now().Format(time.RFC3339),
				},
			},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("same-uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		updates := 0
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				updates++
				return nil
			},
		}
		if recovered, err := MaybeRecover(context.Background(), kube, mg); err != nil || !recovered {
			t.Fatalf("first MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}
		if mg.GetAnnotations()[RecoveredValuesAnnotation] == "" {
			t.Fatalf("fingerprint annotation not set")
		}

		// The API server bumps the generation for the clearing update, the
		// resolver writes the same value back and bumps it again.
		mg.SetGeneration(6)
		mg.Spec.ForProvider.ClientID = ptr("same-uuid")

		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered {
			t.Fatalf("second MaybeRecover() = (%v,%v), want (false,nil)", recovered, err)
		}
		if updates != 1 {
			t.Errorf("kube.Update called %d times, want 1", updates)
		}
		if mg.Spec.ForProvider.ClientID == nil {
			t.Errorf("ClientID was cleared again although the resolver restored the same value")
		}
	})

	t.Run("resolver produced new values → recover again", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 4,
				Annotations: map[string]string{
					meta.AnnotationKeyExternalCreateSucceeded: time.Now().Format(time.RFC3339),
				},
			},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("old-uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error { return nil },
		}
		if recovered, err := MaybeRecover(context.Background(), kube, mg); err != nil || !recovered {
			t.Fatalf("first MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}

		mg.SetGeneration(6)
		mg.Spec.ForProvider.ClientID = ptr("new-uuid")

		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || !recovered {
			t.Fatalf("second MaybeRecover() = (%v,%v), want (true,nil) for changed values", recovered, err)
		}
	})

	// upjet's async callback records a failed create only in the
	// LastAsyncOperation condition; external-create-failed is not set.
	asyncCreateFailed := xpv1.Condition{
		Type:    typeLastAsyncOperation,
		Status:  corev1.ConditionFalse,
		Reason:  reasonAsyncCreateFailure,
		Message: "async create failed: failed to create the resource: [{0 error sending POST request to /admin/realms/r/identity-provider/instances/acme/mappers: 404 Not Found. Response body: {}  []}]",
	}

	t.Run("async create failed, only the start annotation set → never created, skip", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: createdExternally(),
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond, asyncCreateFailed)

		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				t.Errorf("kube.Update was called for a failed async create")
				return nil
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (false,nil)", recovered, err)
		}
	})

	// A resource Keycloak returned before (observed ID) that fails to be
	// re-created because a referenced UUID is stale is exactly the case
	// recovery exists for.
	t.Run("async create failed but an ID was observed before → created, recover", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: createdExternally(),
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.Status.AtProvider.ID = ptr("3357aaeb-6461-40cc-96e6-7e0b6d3a02e5")
		mg.SetConditions(staleCond, asyncCreateFailed)

		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error { return nil },
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || !recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}
	})

	// The resolver may restore the old UUID because the referenced resource
	// has not observed its recreated Keycloak object yet. The values guard
	// must not block that for good.
	t.Run("same values restored, backoff elapsed → recover again", func(t *testing.T) {
		t0 := time.Date(2026, 11, 16, 15, 0, 0, 0, time.UTC)
		clock := t0
		now = func() time.Time { return clock }
		t.Cleanup(func() { now = time.Now })

		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{
				Generation:  4,
				Annotations: map[string]string{meta.AnnotationKeyExternalCreateSucceeded: t0.Format(time.RFC3339)},
			},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("same-uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		updates := 0
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				updates++
				return nil
			},
		}
		if recovered, err := MaybeRecover(context.Background(), kube, mg); err != nil || !recovered {
			t.Fatalf("first MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}

		restore := func(gen int64) {
			mg.SetGeneration(gen)
			mg.Spec.ForProvider.ClientID = ptr("same-uuid")
		}

		restore(6)
		clock = t0.Add(ValuesRecoveryBackoff - time.Second)
		if recovered, err := MaybeRecover(context.Background(), kube, mg); err != nil || recovered {
			t.Fatalf("within backoff: MaybeRecover() = (%v,%v), want (false,nil)", recovered, err)
		}

		clock = t0.Add(ValuesRecoveryBackoff + time.Second)
		if recovered, err := MaybeRecover(context.Background(), kube, mg); err != nil || !recovered {
			t.Fatalf("after backoff: MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}
		if updates != 2 {
			t.Errorf("kube.Update called %d times, want 2", updates)
		}
	})

	t.Run("values annotation malformed → recover (defensive)", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 4,
				Annotations: map[string]string{
					meta.AnnotationKeyExternalCreateSucceeded: time.Now().Format(time.RFC3339),
					RecoveredValuesAnnotation:                 "no-timestamp",
				},
			},
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error { return nil },
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if err != nil || !recovered {
			t.Fatalf("MaybeRecover() = (%v,%v), want (true,nil)", recovered, err)
		}
	})

	t.Run("update non-conflict error → returned", func(t *testing.T) {
		mg := &fakeManaged{
			ObjectMeta: createdExternally(),
			Spec: fakeSpec{ForProvider: fakeParams{
				ClientID:    ptr("uuid"),
				ClientIDRef: &xpv1.Reference{Name: "c"},
			}},
		}
		mg.SetConditions(staleCond)

		wantErr := errors.New("nope")
		kube := &test.MockClient{
			MockUpdate: func(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
				return wantErr
			},
		}
		recovered, err := MaybeRecover(context.Background(), kube, mg)
		if !errors.Is(err, wantErr) {
			t.Fatalf("MaybeRecover() err = %v, want %v", err, wantErr)
		}
		if recovered {
			t.Errorf("expected recovered=false on non-conflict error")
		}
	})
}

// Compile-time check: ensure fakeManaged satisfies the narrow Managed
// interface used by MaybeRecover.
var _ Managed = (*fakeManaged)(nil)

func TestResolvedValuesFingerprint(t *testing.T) {
	withRefs := func(clientID string, roles ...string) *fakeManaged {
		mg := &fakeManaged{Spec: fakeSpec{ForProvider: fakeParams{
			ClientID:    ptr(clientID),
			ClientIDRef: &xpv1.Reference{Name: "c"},
			Name:        ptr("not-a-ref"),
		}}}
		for _, r := range roles {
			mg.Spec.ForProvider.RoleIds = append(mg.Spec.ForProvider.RoleIds, ptr(r))
		}
		if len(roles) > 0 {
			mg.Spec.ForProvider.RoleIdsRefs = []xpv1.Reference{{Name: "r"}}
		}
		return mg
	}

	if got := resolvedValuesFingerprint(&fakeManaged{}); got != "" {
		t.Errorf("no resolved values: fingerprint = %q, want empty", got)
	}
	a := resolvedValuesFingerprint(withRefs("u1", "r1", "r2"))
	if a == "" {
		t.Fatalf("fingerprint is empty")
	}
	if b := resolvedValuesFingerprint(withRefs("u1", "r1", "r2")); a != b {
		t.Errorf("same values: %q != %q", a, b)
	}
	if b := resolvedValuesFingerprint(withRefs("u2", "r1", "r2")); a == b {
		t.Errorf("changed singular value did not change the fingerprint")
	}
	if b := resolvedValuesFingerprint(withRefs("u1", "r1", "r3")); a == b {
		t.Errorf("changed slice value did not change the fingerprint")
	}
	plain := withRefs("u1", "r1", "r2")
	plain.Spec.ForProvider.Name = ptr("changed")
	if b := resolvedValuesFingerprint(plain); a != b {
		t.Errorf("field without ref sibling changed the fingerprint")
	}
}
