/*
Copyright 2022 Upbound Inc.
*/

package connectionsecrettransform

import (
	"bytes"
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

	clientv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha2"
	clusterv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/v1beta1"
)

const (
	secretName = "vikunja-conn"
	secretNS   = "vikunja-ns"
	pcName     = "keycloak-config"
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

// newClient returns an openidclient.Client managed resource with the supplied
// annotations, as the owner of the connection secret under test.
func newClient(annotations map[string]string) *clientv1alpha2.Client {
	return &clientv1alpha2.Client{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "vikunja",
			UID:         types.UID("client-uid"),
			Annotations: annotations,
		},
		Spec: clientv1alpha2.ClientSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{Name: pcName},
			},
		},
	}
}

func newConnectionSecret(mr *clientv1alpha2.Client, data map[string][]byte) *corev1.Secret {
	gvk := clientv1alpha2.CRDGroupVersion.WithKind("Client")
	controller := true
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNS,
			UID:       types.UID("source-uid"),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: gvk.GroupVersion().String(),
					Kind:       gvk.Kind,
					Name:       mr.Name,
					UID:        mr.UID,
					Controller: &controller,
				},
			},
		},
		Type: "connection.crossplane.io/v1alpha1",
		Data: data,
	}
}

func providerConfig(rename map[string]string) *clusterv1beta1.ProviderConfig {
	pc := &clusterv1beta1.ProviderConfig{ObjectMeta: metav1.ObjectMeta{Name: pcName}}
	if rename != nil {
		pc.Spec.ConnectionSecretKeys = &clusterv1beta1.ConnectionSecretKeys{Rename: rename}
	}
	return pc
}

func TestReconcile(t *testing.T) {
	s := newScheme(t)

	data := map[string][]byte{
		"clientID":     []byte("vikunja"),
		"clientSecret": []byte("s3cret"),
		"attribute.x":  []byte("unchanged"),
	}

	cases := map[string]struct {
		pcRename    map[string]string
		annotations map[string]string
		wantName    string
		wantData    map[string][]byte
	}{
		"RenamesKeysConfiguredOnProviderConfig": {
			pcRename: map[string]string{
				"clientID":     "client-id",
				"clientSecret": "client-secret",
			},
			wantName: secretName + transformedSecretSuffix,
			wantData: map[string][]byte{
				"client-id":     []byte("vikunja"),
				"client-secret": []byte("s3cret"),
				"attribute.x":   []byte("unchanged"),
			},
		},
		"RenamesKeysConfiguredOnResourceAnnotation": {
			annotations: map[string]string{
				AnnotationKeyRename: "clientID=client-id, clientSecret=client-secret",
			},
			wantName: secretName + transformedSecretSuffix,
			wantData: map[string][]byte{
				"client-id":     []byte("vikunja"),
				"client-secret": []byte("s3cret"),
				"attribute.x":   []byte("unchanged"),
			},
		},
		"AnnotationOverridesProviderConfigPerKey": {
			pcRename: map[string]string{
				"clientID":     "client-id",
				"clientSecret": "client-secret",
			},
			annotations: map[string]string{
				AnnotationKeyRename: "clientSecret=oidc-secret",
			},
			wantName: secretName + transformedSecretSuffix,
			wantData: map[string][]byte{
				"client-id":   []byte("vikunja"),
				"oidc-secret": []byte("s3cret"),
				"attribute.x": []byte("unchanged"),
			},
		},
		"AnnotationOverridesTransformedSecretName": {
			annotations: map[string]string{
				AnnotationKeyRename:          "clientID=client-id",
				AnnotationKeyTransformedName: "envoy-oidc",
			},
			wantName: "envoy-oidc",
			wantData: map[string][]byte{
				"client-id":    []byte("vikunja"),
				"clientSecret": []byte("s3cret"),
				"attribute.x":  []byte("unchanged"),
			},
		},
		"IgnoresMalformedAndInvalidRenames": {
			annotations: map[string]string{
				AnnotationKeyRename: "clientID=client-id,,broken,clientSecret=not valid",
			},
			wantName: secretName + transformedSecretSuffix,
			wantData: map[string][]byte{
				"client-id":    []byte("vikunja"),
				"clientSecret": []byte("s3cret"),
				"attribute.x":  []byte("unchanged"),
			},
		},
		"NoRenameConfigured": {
			wantName: "",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mr := newClient(tc.annotations)
			secret := newConnectionSecret(mr, data)

			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithObjects(mr, providerConfig(tc.pcRename), secret).
				Build()

			r := &Reconciler{client: cl}
			if _, err := r.Reconcile(context.Background(), request(secret)); err != nil {
				t.Fatalf("Reconcile(...): unexpected error: %v", err)
			}

			l := &corev1.SecretList{}
			if err := cl.List(context.Background(), l, client.InNamespace(secretNS),
				client.MatchingLabels{transformedSecretLabel: transformedSecretLabelValue}); err != nil {
				t.Fatalf("cannot list transformed secrets: %v", err)
			}

			if tc.wantName == "" {
				if len(l.Items) != 0 {
					t.Fatalf("expected no transformed secret, got %d", len(l.Items))
				}
				return
			}

			if len(l.Items) != 1 {
				t.Fatalf("expected exactly one transformed secret, got %d", len(l.Items))
			}
			out := l.Items[0]
			if out.Name != tc.wantName {
				t.Errorf("transformed secret name = %q, want %q", out.Name, tc.wantName)
			}
			if out.Type != secret.Type {
				t.Errorf("transformed secret type = %q, want %q", out.Type, secret.Type)
			}
			if len(out.Data) != len(tc.wantData) {
				t.Fatalf("transformed secret data = %v, want %v", out.Data, tc.wantData)
			}
			for k, v := range tc.wantData {
				if !bytes.Equal(out.Data[k], v) {
					t.Errorf("transformed secret[%q] = %q, want %q", k, out.Data[k], v)
				}
			}
			// The transformed secret must be garbage-collected together with
			// the connection secret, which in turn is owned by the managed
			// resource.
			if !metav1.IsControlledBy(&out, secret) {
				t.Errorf("transformed secret is not controlled by the connection secret: %v", out.OwnerReferences)
			}
			if out.Labels[sourceSecretLabel] != secret.Name {
				t.Errorf("transformed secret source label = %q, want %q", out.Labels[sourceSecretLabel], secret.Name)
			}
		})
	}
}

// TestReconcileCollectsStaleOutput verifies that a secret written by an
// earlier reconcile is removed when it is no longer wanted, both when the
// target name changed and when the renaming was dropped entirely.
func TestReconcileCollectsStaleOutput(t *testing.T) {
	s := newScheme(t)
	data := map[string][]byte{"clientID": []byte("vikunja")}

	cases := map[string]struct {
		annotations map[string]string
		wantNames   []string
	}{
		"RenamedTarget": {
			annotations: map[string]string{
				AnnotationKeyRename:          "clientID=client-id",
				AnnotationKeyTransformedName: "envoy-oidc",
			},
			wantNames: []string{"envoy-oidc"},
		},
		"RenamingRemoved": {
			wantNames: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mr := newClient(tc.annotations)
			secret := newConnectionSecret(mr, data)

			stale := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName + transformedSecretSuffix,
					Namespace: secretNS,
					Labels: map[string]string{
						transformedSecretLabel: transformedSecretLabelValue,
						sourceSecretLabel:      secretName,
					},
					OwnerReferences: secret.OwnerReferences,
				},
				Data: map[string][]byte{"client-id": []byte("vikunja")},
			}
			// Owned (and controlled) by the connection secret.
			controller := true
			stale.OwnerReferences = []metav1.OwnerReference{{
				APIVersion: "v1",
				Kind:       "Secret",
				Name:       secret.Name,
				UID:        secret.UID,
				Controller: &controller,
			}}

			cl := fake.NewClientBuilder().
				WithScheme(s).
				WithObjects(mr, providerConfig(nil), secret, stale).
				Build()

			r := &Reconciler{client: cl}
			if _, err := r.Reconcile(context.Background(), request(secret)); err != nil {
				t.Fatalf("Reconcile(...): unexpected error: %v", err)
			}

			l := &corev1.SecretList{}
			if err := cl.List(context.Background(), l, client.InNamespace(secretNS),
				client.MatchingLabels{transformedSecretLabel: transformedSecretLabelValue}); err != nil {
				t.Fatalf("cannot list transformed secrets: %v", err)
			}
			if len(l.Items) != len(tc.wantNames) {
				t.Fatalf("transformed secrets = %d, want %d", len(l.Items), len(tc.wantNames))
			}
			for i, want := range tc.wantNames {
				if l.Items[i].Name != want {
					t.Errorf("transformed secret[%d] = %q, want %q", i, l.Items[i].Name, want)
				}
			}
		})
	}
}

// TestReconcileForeignSecrets verifies the controller keeps its hands off
// secrets it does not own: its own output, secrets without a Keycloak managed
// resource owner, and secrets that no longer exist.
func TestReconcileForeignSecrets(t *testing.T) {
	s := newScheme(t)

	cases := map[string]*corev1.Secret{
		"OwnOutput": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName + transformedSecretSuffix,
				Namespace: secretNS,
				Labels:    map[string]string{transformedSecretLabel: transformedSecretLabelValue},
			},
		},
		"NoManagedOwner": {
			ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: secretNS},
			Data:       map[string][]byte{"clientID": []byte("vikunja")},
		},
		"ProviderConfigCredentials": {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "keycloak-credentials",
				Namespace: secretNS,
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: clusterv1beta1.SchemeGroupVersion.String(),
					Kind:       "ProviderConfig",
					Name:       pcName,
				}},
			},
		},
	}

	for name, secret := range cases {
		t.Run(name, func(t *testing.T) {
			cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret).Build()
			r := &Reconciler{client: cl}
			if _, err := r.Reconcile(context.Background(), request(secret)); err != nil {
				t.Fatalf("Reconcile(...): unexpected error: %v", err)
			}
			l := &corev1.SecretList{}
			if err := cl.List(context.Background(), l, client.InNamespace(secretNS)); err != nil {
				t.Fatalf("cannot list secrets: %v", err)
			}
			// The reconcile must not have written anything: only the input
			// secret itself is left.
			if len(l.Items) != 1 || l.Items[0].Name != secret.Name {
				t.Fatalf("expected only the input secret to exist, got %d secrets", len(l.Items))
			}
		})
	}

	t.Run("MissingSecret", func(t *testing.T) {
		cl := fake.NewClientBuilder().WithScheme(s).Build()
		r := &Reconciler{client: cl}
		if _, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Name: "does-not-exist", Namespace: secretNS},
		}); err != nil {
			t.Fatalf("Reconcile(...) for a missing secret should be a no-op, got error: %v", err)
		}
	})
}

func TestParseRenameAnnotation(t *testing.T) {
	cases := map[string]struct {
		annotation string
		want       map[string]string
	}{
		"Empty":  {annotation: "", want: nil},
		"Single": {annotation: "clientID=client-id", want: map[string]string{"clientID": "client-id"}},
		"MultipleWithWhitespace": {
			annotation: " clientID = client-id , clientSecret=client-secret ",
			want:       map[string]string{"clientID": "client-id", "clientSecret": "client-secret"},
		},
		"SkipsMalformed": {
			annotation: "clientID=client-id,,noseparator,=novalue,nokey=",
			want:       map[string]string{"clientID": "client-id"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := parseRenameAnnotation(tc.annotation)
			if len(got) != len(tc.want) {
				t.Fatalf("parseRenameAnnotation(%q) = %v, want %v", tc.annotation, got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Errorf("parseRenameAnnotation(%q)[%q] = %q, want %q", tc.annotation, k, got[k], v)
				}
			}
		})
	}
}

func request(s *corev1.Secret) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Name: s.Name, Namespace: s.Namespace}}
}
