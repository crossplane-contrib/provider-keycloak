/*
Copyright 2022 Upbound Inc.
*/

package connectionsecrettransform

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

	clientv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha2"
	clusterv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/v1beta1"
)

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("cannot add corev1 to scheme: %v", err)
	}
	if err := clientv1alpha2.SchemeBuilder.AddToScheme(s); err != nil {
		t.Fatalf("cannot add openidclient v1alpha2 to scheme: %v", err)
	}
	if err := clusterv1beta1.SchemeBuilder.AddToScheme(s); err != nil {
		t.Fatalf("cannot add cluster v1beta1 to scheme: %v", err)
	}
	return s
}

func TestReconcile(t *testing.T) {
	s := newScheme(t)

	mr := &clientv1alpha2.Client{
		ObjectMeta: metav1.ObjectMeta{Name: "vikunja"},
		Spec: clientv1alpha2.ClientSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{Name: "keycloak-config"},
			},
		},
	}
	gvk := clientv1alpha2.CRDGroupVersion.WithKind(clientKind)

	cases := map[string]struct {
		pc         *clusterv1beta1.ProviderConfig
		secretData map[string][]byte
		wantData   map[string][]byte
		wantExists bool
	}{
		"RenamesConfiguredKeys": {
			pc: &clusterv1beta1.ProviderConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "keycloak-config"},
				Spec: clusterv1beta1.ProviderConfigSpec{
					ConnectionSecretKeys: &clusterv1beta1.ConnectionSecretKeys{
						Rename: map[string]string{
							"clientID":     "client-id",
							"clientSecret": "client-secret",
						},
					},
				},
			},
			secretData: map[string][]byte{
				"clientID":     []byte("vikunja"),
				"clientSecret": []byte("s3cret"),
				"attribute.x":  []byte("unchanged"),
			},
			wantData: map[string][]byte{
				"client-id":     []byte("vikunja"),
				"client-secret": []byte("s3cret"),
				"attribute.x":   []byte("unchanged"),
			},
			wantExists: true,
		},
		"NoRenameConfigured": {
			pc: &clusterv1beta1.ProviderConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "keycloak-config"},
			},
			secretData: map[string][]byte{
				"clientID": []byte("vikunja"),
			},
			wantExists: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vikunja-conn",
					Namespace: "vikunja-ns",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: gvk.GroupVersion().String(),
							Kind:       gvk.Kind,
							Name:       mr.Name,
							UID:        mr.UID,
						},
					},
				},
				Data: tc.secretData,
			}

			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithObjects(mr.DeepCopy(), tc.pc.DeepCopy(), secret).
				Build()

			r := &Reconciler{client: cl}
			if _, err := r.Reconcile(context.Background(), reconcile.Request{
				NamespacedName: types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace},
			}); err != nil {
				t.Fatalf("Reconcile(...): unexpected error: %v", err)
			}

			out := &corev1.Secret{}
			err := cl.Get(context.Background(), types.NamespacedName{
				Name:      secret.Name + transformedSecretSuffix,
				Namespace: secret.Namespace,
			}, out)
			if !tc.wantExists {
				if err == nil {
					t.Fatalf("expected no transformed secret to be written, got one with data %v", out.Data)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected transformed secret to exist: %v", err)
			}
			if len(out.Data) != len(tc.wantData) {
				t.Fatalf("transformed secret data = %v, want %v", out.Data, tc.wantData)
			}
			for k, v := range tc.wantData {
				if string(out.Data[k]) != string(v) {
					t.Errorf("transformed secret[%q] = %q, want %q", k, out.Data[k], v)
				}
			}
		})
	}
}

func TestReconcileIgnoresOwnOutput(t *testing.T) {
	s := newScheme(t)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vikunja-conn-transformed",
			Namespace: "vikunja-ns",
			Labels:    map[string]string{transformedSecretLabel: "true"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret).Build()
	r := &Reconciler{client: cl}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace},
	}); err != nil {
		t.Fatalf("Reconcile(...): unexpected error: %v", err)
	}
}

func TestReconcileMissingSecret(t *testing.T) {
	s := newScheme(t)
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	r := &Reconciler{client: cl}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "does-not-exist", Namespace: "ns"},
	}); err != nil {
		t.Fatalf("Reconcile(...) for a missing secret should be a no-op, got error: %v", err)
	}
}
