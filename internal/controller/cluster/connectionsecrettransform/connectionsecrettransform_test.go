/*
Copyright 2022 Upbound Inc.
*/

package connectionsecrettransform

import (
	"bytes"
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"

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

// newReconciler returns a Reconciler wired up like Setup does, minus the
// manager: events go nowhere and logs are dropped.
func newReconciler(cl client.Client) *Reconciler {
	return &Reconciler{
		client:         cl,
		log:            logging.NewNopLogger(),
		record:         event.NewNopRecorder(),
		resyncInterval: time.Minute,
	}
}

func providerConfig(rename map[string]string) *clusterv1beta1.ProviderConfig {
	pc := &clusterv1beta1.ProviderConfig{ObjectMeta: metav1.ObjectMeta{Name: pcName}}
	if rename != nil {
		pc.Spec.ConnectionSecretKeys = &clusterv1beta1.ConnectionSecretKeys{Rename: rename}
	}
	return pc
}

// providerConfigWithAdd returns a ProviderConfig configuring
// spec.connectionSecretKeys.add.
func providerConfigWithAdd(add map[string]string) *clusterv1beta1.ProviderConfig {
	return &clusterv1beta1.ProviderConfig{
		ObjectMeta: metav1.ObjectMeta{Name: pcName},
		Spec:       clusterv1beta1.ProviderConfigSpec{ConnectionSecretKeys: &clusterv1beta1.ConnectionSecretKeys{Add: add}},
	}
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

			r := newReconciler(cl)
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

			r := newReconciler(cl)
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
			r := newReconciler(cl)
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
		r := newReconciler(cl)
		if _, err := r.Reconcile(context.Background(), reconcile.Request{
			NamespacedName: types.NamespacedName{Name: "does-not-exist", Namespace: secretNS},
		}); err != nil {
			t.Fatalf("Reconcile(...) for a missing secret should be a no-op, got error: %v", err)
		}
	})
}

// TestReconcileResync verifies that a connection secret owned by a Keycloak
// managed resource is requeued periodically. The rename configuration lives
// on the managed resource's annotations and its ProviderConfig, neither of
// which produces a Secret event, so without the requeue an annotation-only
// edit would not be picked up.
func TestReconcileResync(t *testing.T) {
	s := newScheme(t)
	mr := newClient(map[string]string{AnnotationKeyRename: "clientID=client-id"})
	secret := newConnectionSecret(mr, map[string][]byte{"clientID": []byte("vikunja")})

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(mr, providerConfig(nil), secret).Build()
	r := newReconciler(cl)

	res, err := r.Reconcile(context.Background(), request(secret))
	if err != nil {
		t.Fatalf("Reconcile(...): unexpected error: %v", err)
	}
	if res.RequeueAfter != time.Minute {
		t.Errorf("RequeueAfter = %v, want %v", res.RequeueAfter, time.Minute)
	}

	// A secret this controller is not responsible for must not be requeued.
	other := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: secretNS}}
	cl = fake.NewClientBuilder().WithScheme(s).WithObjects(other).Build()
	r = newReconciler(cl)

	res, err = r.Reconcile(context.Background(), request(other))
	if err != nil {
		t.Fatalf("Reconcile(...): unexpected error: %v", err)
	}
	if res.RequeueAfter != 0 {
		t.Errorf("RequeueAfter = %v, want 0 for a foreign secret", res.RequeueAfter)
	}
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
		"Multiline": {
			annotation: "clientID=client-id\nclientSecret=client-secret\n",
			want:       map[string]string{"clientID": "client-id", "clientSecret": "client-secret"},
		},
		"MultilineAndComma": {
			annotation: "clientID=client-id, serviceAccountUserId=service-account-user-id\nclientSecret=client-secret",
			want: map[string]string{
				"clientID":             "client-id",
				"serviceAccountUserId": "service-account-user-id",
				"clientSecret":         "client-secret",
			},
		},
		"YAMLMapping": {
			annotation: "clientID: client-id\nclientSecret: client-secret\n",
			want: map[string]string{
				"clientID":     "client-id",
				"clientSecret": "client-secret",
			},
		},
		"YAMLMappingSingleEntry": {
			annotation: "clientID: client-id",
			want:       map[string]string{"clientID": "client-id"},
		},
		"YAMLMappingIgnoresEmptyValue": {
			annotation: "clientID: client-id\nclientSecret:\n",
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

// TestReconcileRefusesForeignSecret verifies the controller never overwrites a
// secret it did not write itself, which would otherwise let a stray
// annotation destroy unrelated data (e.g. the provider's own credentials).
func TestReconcileRefusesForeignSecret(t *testing.T) {
	s := newScheme(t)
	mr := newClient(map[string]string{
		AnnotationKeyRename:          "clientID=client-id",
		AnnotationKeyTransformedName: "keycloak-credentials",
	})
	secret := newConnectionSecret(mr, map[string][]byte{"clientID": []byte("vikunja")})

	foreign := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "keycloak-credentials", Namespace: secretNS},
		Data:       map[string][]byte{"credentials": []byte("do-not-touch")},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(mr, providerConfig(nil), secret, foreign).Build()
	r := newReconciler(cl)

	if _, err := r.Reconcile(context.Background(), request(secret)); err != nil {
		t.Fatalf("Reconcile(...): unexpected error: %v", err)
	}

	got := &corev1.Secret{}
	if err := cl.Get(context.Background(), types.NamespacedName{Name: foreign.Name, Namespace: secretNS}, got); err != nil {
		t.Fatalf("cannot get the foreign secret: %v", err)
	}
	if !bytes.Equal(got.Data["credentials"], []byte("do-not-touch")) {
		t.Errorf("foreign secret was modified: %v", got.Data)
	}
	if _, renamed := got.Data["client-id"]; renamed {
		t.Errorf("foreign secret was overwritten with transformed data: %v", got.Data)
	}
	if got.Labels[transformedSecretLabel] != "" {
		t.Errorf("foreign secret was adopted: %v", got.Labels)
	}
}

// TestReconcileInvalidTransformName verifies that a transform name that
// Kubernetes would reject, or that points at the connection secret itself, is
// skipped instead of retried forever or applied destructively.
func TestReconcileInvalidTransformName(t *testing.T) {
	s := newScheme(t)
	data := map[string][]byte{"clientID": []byte("vikunja")}

	for name, transformName := range map[string]string{
		"NotADNSName":       "Not A Valid Name",
		"TheSourceItself":   secretName,
		"TrailingSeparator": "invalid-",
	} {
		t.Run(name, func(t *testing.T) {
			mr := newClient(map[string]string{
				AnnotationKeyRename:          "clientID=client-id",
				AnnotationKeyTransformedName: transformName,
			})
			secret := newConnectionSecret(mr, data)

			cl := fake.NewClientBuilder().WithScheme(s).WithObjects(mr, providerConfig(nil), secret).Build()
			r := newReconciler(cl)

			if _, err := r.Reconcile(context.Background(), request(secret)); err != nil {
				t.Fatalf("Reconcile(...): unexpected error: %v", err)
			}

			l := &corev1.SecretList{}
			if err := cl.List(context.Background(), l, client.InNamespace(secretNS)); err != nil {
				t.Fatalf("cannot list secrets: %v", err)
			}
			if len(l.Items) != 1 || l.Items[0].Name != secretName {
				t.Fatalf("expected only the untouched connection secret, got %d secrets", len(l.Items))
			}
			if _, renamed := l.Items[0].Data["client-id"]; renamed {
				t.Errorf("the connection secret itself was transformed: %v", l.Items[0].Data)
			}
		})
	}
}

// TestReconcileSkipsDeletedSecret verifies the controller does not recreate
// output for a connection secret that is being garbage-collected.
func TestReconcileSkipsDeletedSecret(t *testing.T) {
	s := newScheme(t)
	mr := newClient(map[string]string{AnnotationKeyRename: "clientID=client-id"})
	secret := newConnectionSecret(mr, map[string][]byte{"clientID": []byte("vikunja")})
	now := metav1.Now()
	secret.DeletionTimestamp = &now
	secret.Finalizers = []string{"keycloak.crossplane.io/test"}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(mr, providerConfig(nil), secret).Build()
	r := newReconciler(cl)

	if _, err := r.Reconcile(context.Background(), request(secret)); err != nil {
		t.Fatalf("Reconcile(...): unexpected error: %v", err)
	}

	l := &corev1.SecretList{}
	if err := cl.List(context.Background(), l, client.InNamespace(secretNS),
		client.MatchingLabels{transformedSecretLabel: transformedSecretLabelValue}); err != nil {
		t.Fatalf("cannot list transformed secrets: %v", err)
	}
	if len(l.Items) != 0 {
		t.Fatalf("expected no transformed secret for a deleted source, got %d", len(l.Items))
	}
}

func TestTransform(t *testing.T) {
	data := map[string][]byte{
		"clientID":     []byte("id"),
		"clientSecret": []byte("secret"),
		"extra":        []byte("extra"),
	}

	cases := map[string]struct {
		rename        map[string]string
		want          map[string][]byte
		wantConflicts []string
	}{
		"RenamesAndCopies": {
			rename: map[string]string{"clientID": "client-id"},
			want: map[string][]byte{
				"client-id":    []byte("id"),
				"clientSecret": []byte("secret"),
				"extra":        []byte("extra"),
			},
		},
		"IgnoresRenamesForAbsentKeys": {
			rename: map[string]string{"nonexistent": "whatever"},
			want:   data,
		},
		"DropsRenameOntoAnUntouchedKey": {
			// clientID -> extra would overwrite the "extra" key, so the
			// rename is skipped and no value is lost.
			rename:        map[string]string{"clientID": "extra"},
			want:          data,
			wantConflicts: []string{`"clientID"="extra"`},
		},
		"DropsRenamesOntoTheSameTarget": {
			rename: map[string]string{"clientID": "same", "clientSecret": "same"},
			want: map[string][]byte{
				"same":         []byte("id"),
				"clientSecret": []byte("secret"),
				"extra":        []byte("extra"),
			},
			wantConflicts: []string{`"clientSecret"="same"`},
		},
		"SwapIsNotAConflict": {
			rename: map[string]string{"clientID": "clientSecret", "clientSecret": "clientID"},
			want: map[string][]byte{
				"clientSecret": []byte("id"),
				"clientID":     []byte("secret"),
				"extra":        []byte("extra"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, conflicts := transform(data, tc.rename)
			if len(got) != len(tc.want) {
				t.Fatalf("transform(...) = %v, want %v", got, tc.want)
			}
			for k, v := range tc.want {
				if !bytes.Equal(got[k], v) {
					t.Errorf("transform(...)[%q] = %q, want %q", k, got[k], v)
				}
			}
			if len(conflicts) != len(tc.wantConflicts) {
				t.Fatalf("conflicts = %v, want %v", conflicts, tc.wantConflicts)
			}
			for i, want := range tc.wantConflicts {
				if conflicts[i] != want {
					t.Errorf("conflicts[%d] = %q, want %q", i, conflicts[i], want)
				}
			}
		})
	}
}

func TestIsRelevantSecret(t *testing.T) {
	cases := map[string]struct {
		object client.Object
		want   bool
	}{
		"ConnectionSecret": {
			object: &corev1.Secret{Type: xpresource.SecretTypeConnection},
			want:   true,
		},
		"OwnOutput": {
			object: &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{transformedSecretLabel: transformedSecretLabelValue},
			}},
			want: true,
		},
		"CredentialsSecret": {
			object: &corev1.Secret{Type: corev1.SecretTypeOpaque},
			want:   false,
		},
		"NotASecret": {
			object: &corev1.ConfigMap{},
			want:   false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := isRelevantSecret(tc.object); got != tc.want {
				t.Errorf("isRelevantSecret(...) = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestResolveAddSource(t *testing.T) {
	mrObj := map[string]interface{}{
		"status": map[string]interface{}{
			"atProvider": map[string]interface{}{
				"clientId": "vikunja",
				"enabled":  true,
				"count":    int64(3),
				"nested":   map[string]interface{}{"foo": "bar"},
			},
		},
	}
	pcObj := map[string]interface{}{
		"metadata": map[string]interface{}{"name": pcName},
	}

	cases := map[string]struct {
		expr    string
		mrObj   map[string]interface{}
		pcObj   map[string]interface{}
		want    string
		wantErr bool
	}{
		"StatusString":      {expr: "status:atProvider.clientId", mrObj: mrObj, want: "vikunja"},
		"StatusBool":        {expr: "status:atProvider.enabled", mrObj: mrObj, want: "true"},
		"StatusNumber":      {expr: "status:atProvider.count", mrObj: mrObj, want: "3"},
		"StatusNotFound":    {expr: "status:atProvider.missing", mrObj: mrObj, wantErr: true},
		"StatusNonScalar":   {expr: "status:atProvider.nested", mrObj: mrObj, wantErr: true},
		"ProviderConfig":    {expr: "providerConfig:metadata.name", pcObj: pcObj, want: pcName},
		"ProviderConfigNil": {expr: "providerConfig:metadata.name", pcObj: nil, wantErr: true},
		"UnsupportedPrefix": {expr: "spec:forProvider.clientId", mrObj: mrObj, wantErr: true},
		"EmptyPath":         {expr: "status:", mrObj: mrObj, wantErr: true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := resolveAddSource(tc.expr, tc.mrObj, tc.pcObj)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveAddSource(%q) = %q, want error", tc.expr, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveAddSource(%q): unexpected error: %v", tc.expr, err)
			}
			if got != tc.want {
				t.Errorf("resolveAddSource(%q) = %q, want %q", tc.expr, got, tc.want)
			}
		})
	}
}

func TestMergeAdded(t *testing.T) {
	data := map[string][]byte{"clientID": []byte("vikunja")}
	added := map[string][]byte{
		"issuerUrl": []byte("https://keycloak.example.com"),
		"clientID":  []byte("would-overwrite"),
	}

	conflicts := mergeAdded(data, added)
	if len(conflicts) != 1 || conflicts[0] != "clientID" {
		t.Fatalf("mergeAdded(...) conflicts = %v, want [clientID]", conflicts)
	}
	if string(data["clientID"]) != "vikunja" {
		t.Errorf("mergeAdded(...) overwrote data[clientID] = %q, want unchanged", data["clientID"])
	}
	if string(data["issuerUrl"]) != "https://keycloak.example.com" {
		t.Errorf("mergeAdded(...) data[issuerUrl] = %q, want the added value", data["issuerUrl"])
	}
}

func TestAddFieldMap(t *testing.T) {
	s := newScheme(t)

	mr := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "openidclient.keycloak.crossplane.io/v1alpha2",
		"kind":       "Client",
		"metadata": map[string]interface{}{
			"name": "vikunja",
			"annotations": map[string]interface{}{
				AnnotationKeyAddFields: "internalId=status:atProvider.id",
			},
		},
		"spec": map[string]interface{}{
			"providerConfigRef": map[string]interface{}{"name": pcName},
		},
		"status": map[string]interface{}{
			"atProvider": map[string]interface{}{"id": "abc-123"},
		},
	}}

	pc := providerConfigWithAdd(map[string]string{"issuerUrl": "providerConfig:metadata.name"})
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(pc).Build()
	r := newReconciler(cl)

	got, invalid, err := r.addFieldMap(context.Background(), mr)
	if err != nil {
		t.Fatalf("addFieldMap(...): unexpected error: %v", err)
	}
	if len(invalid) != 0 {
		t.Errorf("addFieldMap(...) invalid = %v, want none", invalid)
	}
	if string(got["issuerUrl"]) != pcName {
		t.Errorf("addFieldMap(...)[issuerUrl] = %q, want %q", got["issuerUrl"], pcName)
	}
	if string(got["internalId"]) != "abc-123" {
		t.Errorf("addFieldMap(...)[internalId] = %q, want %q", got["internalId"], "abc-123")
	}

	// An unresolved source must not fail the whole reconcile: it is
	// reported and the resource keeps only the still-resolvable fields.
	mr.Object["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
		AnnotationKeyAddFields: "internalId=status:atProvider.missing",
	}
	got, invalid, err = r.addFieldMap(context.Background(), mr)
	if err != nil {
		t.Fatalf("addFieldMap(...): unexpected error: %v", err)
	}
	if len(invalid) != 1 {
		t.Fatalf("addFieldMap(...) invalid = %v, want one entry", invalid)
	}
	if _, ok := got["internalId"]; ok {
		t.Errorf("addFieldMap(...) resolved %q despite an unresolvable source", "internalId")
	}
	if string(got["issuerUrl"]) != pcName {
		t.Errorf("addFieldMap(...)[issuerUrl] = %q, want %q (still applied from ProviderConfig)", got["issuerUrl"], pcName)
	}
}
