package config

import (
	"encoding/json"
	"testing"

	"github.com/crossplane/upjet/v2/pkg/controller/conversion"
	"k8s.io/apimachinery/pkg/runtime"

	oidcv1alpha1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/oidc/v1alpha1"
	oidcv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/oidc/v1alpha2"
	clientv1alpha1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha1"
	clientv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha2"
)

// storedV1Alpha1Client is a Client as it is persisted in etcd by
// provider-keycloak <= v2.x, i.e. with a numeric clientSecretWoVersion.
const storedV1Alpha1Client = `{
  "apiVersion": "openidclient.keycloak.crossplane.io/v1alpha1",
  "kind": "Client",
  "metadata": {"name": "client"},
  "spec": {"forProvider": {"clientId": "client", "realmId": "realm", "accessType": "CONFIDENTIAL", "clientSecretWoVersion": 202605192}},
  "status": {"atProvider": {"clientSecretWoVersion": 202605192}}
}`

// storedStringV1Alpha1Client is a Client as it was persisted at v1alpha1 by
// the pre-v3.0.0 development builds published between the terraform provider
// v5.9.0 bump and the v1alpha2 split, i.e. with a string clientSecretWoVersion
// (crossplane-contrib/provider-keycloak#669).
const storedStringV1Alpha1Client = `{
  "apiVersion": "openidclient.keycloak.crossplane.io/v1alpha1",
  "kind": "Client",
  "metadata": {"name": "client"},
  "spec": {"forProvider": {"clientId": "client", "realmId": "realm", "accessType": "CONFIDENTIAL", "clientSecretWoVersion": "202605192"}},
  "status": {"atProvider": {"clientSecretWoVersion": "202605192"}}
}`

// TestStoredV1Alpha1ObjectsRequireAConversionWebhook is the proof that the
// number -> string type change of client_secret_wo_version (introduced by
// terraform-provider-keycloak v5.9.0) cannot be shipped in-place on v1alpha1.
//
// Objects created with the released provider are persisted with a numeric
// clientSecretWoVersion. If the same API version simply changed its type to
// string, those objects could no longer be decoded into the Go type, and every
// write would be rejected by the API server's schema validation. Serving a
// second API version and converting between them is the only lossless option.
func TestStoredV1Alpha1ObjectsRequireAConversionWebhook(t *testing.T) {
	if err := json.Unmarshal([]byte(storedV1Alpha1Client), &clientv1alpha2.Client{}); err == nil {
		t.Fatal("expected the stored numeric clientSecretWoVersion to be undecodable by the string-typed API version")
	}
	if err := json.Unmarshal([]byte(storedV1Alpha1Client), &clientv1alpha1.Client{}); err != nil {
		t.Fatalf("the frozen v1alpha1 API version must keep decoding stored objects: %v", err)
	}
}

// TestClientSecretWoVersionConversion exercises the very same code path the
// conversion webhook uses at runtime (upjet's conversion registry + RoundTrip)
// for both the upgrade (v1alpha1 -> v1alpha2) and the downgrade
// (v1alpha2 -> v1alpha1) direction.
func TestClientSecretWoVersionConversion(t *testing.T) {
	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		clientv1alpha1.SchemeBuilder.AddToScheme,
		clientv1alpha2.SchemeBuilder.AddToScheme,
		oidcv1alpha1.SchemeBuilder.AddToScheme,
		oidcv1alpha2.SchemeBuilder.AddToScheme,
	} {
		if err := add(scheme); err != nil {
			t.Fatalf("cannot build the test scheme: %v", err)
		}
	}

	pcCluster, err := GetProvider(true)
	if err != nil {
		t.Fatalf("cannot get the cluster-scoped provider configuration: %v", err)
	}
	pcNamespaced, err := GetProviderNamespaced(true)
	if err != nil {
		t.Fatalf("cannot get the namespaced provider configuration: %v", err)
	}
	if err := conversion.RegisterConversions(pcCluster, pcNamespaced, scheme); err != nil {
		t.Fatalf("cannot register the conversions: %v", err)
	}

	t.Run("Upgrade", func(t *testing.T) {
		src := &clientv1alpha1.Client{}
		if err := json.Unmarshal([]byte(storedV1Alpha1Client), src); err != nil {
			t.Fatalf("cannot decode the stored object: %v", err)
		}
		dst := &clientv1alpha2.Client{}
		if err := src.ConvertTo(dst); err != nil {
			t.Fatalf("cannot convert v1alpha1 to v1alpha2: %v", err)
		}
		if got := dst.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != "202605192" {
			t.Errorf("spec.forProvider.clientSecretWoVersion: want %q, got %v", "202605192", got)
		}
		if got := dst.Status.AtProvider.ClientSecretWoVersion; got == nil || *got != "202605192" {
			t.Errorf("status.atProvider.clientSecretWoVersion: want %q, got %v", "202605192", got)
		}
		if got := dst.Spec.ForProvider.ClientID; got == nil || *got != "client" {
			t.Errorf("spec.forProvider.clientId: want %q, got %v", "client", got)
		}
	})

	t.Run("Downgrade", func(t *testing.T) {
		version := "202605192"
		src := &clientv1alpha2.Client{}
		src.Name = "client"
		src.Spec.ForProvider.ClientSecretWoVersion = &version
		src.Status.AtProvider.ClientSecretWoVersion = &version

		dst := &clientv1alpha1.Client{}
		if err := dst.ConvertFrom(src); err != nil {
			t.Fatalf("cannot convert v1alpha2 to v1alpha1: %v", err)
		}
		if got := dst.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != 202605192 {
			t.Errorf("spec.forProvider.clientSecretWoVersion: want %v, got %v", 202605192, got)
		}
		if got := dst.Status.AtProvider.ClientSecretWoVersion; got == nil || *got != 202605192 {
			t.Errorf("status.atProvider.clientSecretWoVersion: want %v, got %v", 202605192, got)
		}
	})

	t.Run("IdentityProviderUpgrade", func(t *testing.T) {
		version := float64(202605192)
		src := &oidcv1alpha1.IdentityProvider{}
		src.Name = "idp"
		src.Spec.ForProvider.ClientSecretWoVersion = &version

		dst := &oidcv1alpha2.IdentityProvider{}
		if err := src.ConvertTo(dst); err != nil {
			t.Fatalf("cannot convert v1alpha1 to v1alpha2: %v", err)
		}
		if got := dst.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != "202605192" {
			t.Errorf("spec.forProvider.clientSecretWoVersion: want %q, got %v", "202605192", got)
		}
	})

	// Pre-v3.0.0 development builds persisted a string clientSecretWoVersion at
	// v1alpha1 (crossplane-contrib/provider-keycloak#669). The conversion
	// webhook's typed decode must tolerate that encoding and the upgrade must
	// still yield the v1alpha2 string.
	t.Run("StringEncodedV1Alpha1Upgrade", func(t *testing.T) {
		src := &clientv1alpha1.Client{}
		if err := json.Unmarshal([]byte(storedStringV1Alpha1Client), src); err != nil {
			t.Fatalf("cannot decode the string-encoded stored object: %v", err)
		}
		if got := src.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != 202605192 {
			t.Fatalf("spec.forProvider.clientSecretWoVersion: want %v, got %v", 202605192, got)
		}
		dst := &clientv1alpha2.Client{}
		if err := src.ConvertTo(dst); err != nil {
			t.Fatalf("cannot convert v1alpha1 to v1alpha2: %v", err)
		}
		if got := dst.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != "202605192" {
			t.Errorf("spec.forProvider.clientSecretWoVersion: want %q, got %v", "202605192", got)
		}
		if got := dst.Status.AtProvider.ClientSecretWoVersion; got == nil || *got != "202605192" {
			t.Errorf("status.atProvider.clientSecretWoVersion: want %q, got %v", "202605192", got)
		}
	})

	t.Run("StringEncodedV1Alpha1IdentityProvider", func(t *testing.T) {
		const stored = `{
  "apiVersion": "oidc.keycloak.m.crossplane.io/v1alpha1",
  "kind": "IdentityProvider",
  "metadata": {"name": "idp", "namespace": "ns"},
  "spec": {"forProvider": {"alias": "idp", "realm": "realm", "clientSecretWoVersion": "1277175040"}}
}`
		src := &oidcv1alpha1.IdentityProvider{}
		if err := json.Unmarshal([]byte(stored), src); err != nil {
			t.Fatalf("cannot decode the string-encoded stored object: %v", err)
		}
		if got := src.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != 1277175040 {
			t.Fatalf("spec.forProvider.clientSecretWoVersion: want %v, got %v", 1277175040, got)
		}
	})

	t.Run("NonNumericStringIsRejected", func(t *testing.T) {
		src := &clientv1alpha1.Client{}
		if err := json.Unmarshal([]byte(`{"spec": {"forProvider": {"clientSecretWoVersion": "version1"}}}`), src); err == nil {
			t.Fatal("expected a descriptive error for a non-numeric string clientSecretWoVersion")
		}
	})
}
