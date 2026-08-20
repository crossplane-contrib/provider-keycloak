/*
Copyright 2022 Upbound Inc.
*/

package connectionsecrettransform

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

	clusterv1beta1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/v1beta1"
)

func TestKeycloakBaseURL(t *testing.T) {
	cases := map[string]struct {
		url      string
		basePath string
		want     string
		wantErr  bool
	}{
		"Plain":            {url: "https://keycloak.example.com", want: "https://keycloak.example.com"},
		"TrailingSlash":    {url: "https://keycloak.example.com/", want: "https://keycloak.example.com"},
		"BasePath":         {url: "https://example.com", basePath: "/auth", want: "https://example.com/auth"},
		"BasePathNoSlash":  {url: "https://example.com", basePath: "auth", want: "https://example.com/auth"},
		"BasePathTrailing": {url: "https://example.com", basePath: "/auth/", want: "https://example.com/auth"},
		"Whitespace":       {url: "  https://example.com  ", want: "https://example.com"},
		"Empty":            {url: "", wantErr: true},
		"NoScheme":         {url: "keycloak.example.com", wantErr: true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := keycloakBaseURL(tc.url, tc.basePath)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("keycloakBaseURL(%q, %q) = %q, want error", tc.url, tc.basePath, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("keycloakBaseURL(%q, %q): unexpected error: %v", tc.url, tc.basePath, err)
			}
			if got != tc.want {
				t.Errorf("keycloakBaseURL(%q, %q) = %q, want %q", tc.url, tc.basePath, got, tc.want)
			}
		})
	}
}

func TestRealmOf(t *testing.T) {
	// nested builds {outer: {field: value}}, the shape a managed resource
	// spells its realm in.
	nested := func(outer, field, value string) map[string]interface{} {
		return map[string]interface{}{outer: map[string]interface{}{field: value}}
	}

	cases := map[string]struct {
		obj  map[string]interface{}
		want string
	}{
		"StatusRealmID": {obj: map[string]interface{}{
			specField:   nested(forProviderField, realmIDField, "spec-realm"),
			statusField: nested(atProviderField, realmIDField, "status-realm"),
		}, want: "status-realm"},
		"SpecRealmID": {obj: map[string]interface{}{
			specField: nested(forProviderField, realmIDField, "spec-realm"),
		}, want: "spec-realm"},
		"RealmResource": {obj: map[string]interface{}{
			statusField: nested(atProviderField, realmField, "my-realm"),
		}, want: "my-realm"},
		"Unknown": {obj: map[string]interface{}{specField: map[string]interface{}{forProviderField: map[string]interface{}{}}}, want: ""},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := realmOf(tc.obj); got != tc.want {
				t.Errorf("realmOf(...) = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestKeycloakSourceValueWithoutRealm verifies that a field that needs the
// realm is reported as unresolvable (rather than producing a bogus URL) when
// the owning resource's realm cannot be determined, while "keycloak:url"
// still works.
func TestKeycloakSourceValueWithoutRealm(t *testing.T) {
	kc := &keycloakSource{resolve: func() (string, string, error) {
		return "https://keycloak.example.com", "", nil
	}}

	if got, err := kc.value(keycloakFieldURL); err != nil || got != "https://keycloak.example.com" {
		t.Errorf("value(url) = %q, %v; want the base URL and no error", got, err)
	}
	if got, err := kc.value(keycloakFieldIssuerURL); err == nil {
		t.Errorf("value(issuerUrl) = %q, want an error when the realm is unknown", got)
	}
}

// TestAddFieldMapKeycloakSource exercises the full path: a "keycloak:"
// source resolved from the ProviderConfig's credentials Secret and the
// owning managed resource's realm.
func TestAddFieldMapKeycloakSource(t *testing.T) {
	s := newScheme(t)

	creds := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "keycloak-credentials", Namespace: "crossplane-system"},
		Data: map[string][]byte{
			"credentials": []byte(`{"client_id":"admin-cli","url":"https://keycloak.example.com/","password":"do-not-leak"}`),
		},
	}

	pc := &clusterv1beta1.ProviderConfig{
		ObjectMeta: metav1.ObjectMeta{Name: pcName},
		Spec: clusterv1beta1.ProviderConfigSpec{
			Credentials: clusterv1beta1.ProviderCredentials{
				Source: xpv1.CredentialsSourceSecret,
				CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
					SecretRef: &xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{Name: creds.Name, Namespace: creds.Namespace},
						Key:             "credentials",
					},
				},
			},
		},
	}

	mr := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "openidclient.keycloak.crossplane.io/v1alpha2",
		"kind":       "Client",
		"metadata": map[string]interface{}{
			"name": "vikunja",
			"annotations": map[string]interface{}{
				AnnotationKeyAddFields: "issuerUrl=keycloak:issuerUrl,tokenUrl=keycloak:tokenUrl",
			},
		},
		"spec": map[string]interface{}{
			"providerConfigRef": map[string]interface{}{"name": pcName},
		},
		"status": map[string]interface{}{
			"atProvider": map[string]interface{}{"realmId": "my-realm"},
		},
	}}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(pc, creds).Build()
	r := newReconciler(cl)

	got, invalid, err := r.addFieldMap(context.Background(), mr, nil)
	if err != nil {
		t.Fatalf("addFieldMap(...): unexpected error: %v", err)
	}
	if len(invalid) != 0 {
		t.Fatalf("addFieldMap(...) invalid = %v, want none", invalid)
	}
	if want := "https://keycloak.example.com/realms/my-realm"; string(got["issuerUrl"]) != want {
		t.Errorf("addFieldMap(...)[issuerUrl] = %q, want %q", got["issuerUrl"], want)
	}
	if want := "https://keycloak.example.com/realms/my-realm/protocol/openid-connect/token"; string(got["tokenUrl"]) != want {
		t.Errorf("addFieldMap(...)[tokenUrl] = %q, want %q", got["tokenUrl"], want)
	}
	for k, v := range got {
		if string(v) == "do-not-leak" {
			t.Errorf("addFieldMap(...)[%s] published credential material", k)
		}
	}
}
