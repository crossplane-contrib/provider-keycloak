package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/crossplane/upjet/v2/pkg/controller/conversion"
	"k8s.io/apimachinery/pkg/runtime"

	oidcv1alpha1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/oidc/v1alpha1"
	oidcv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/oidc/v1alpha2"
	clientv1alpha1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha1"
	clientv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha2"
	clientv1alpha3 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/openidclient/v1alpha3"
	realmv1alpha1 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/realm/v1alpha1"
	realmv1alpha2 "github.com/crossplane-contrib/provider-keycloak/apis/cluster/realm/v1alpha2"
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
// for both the upgrade (v1alpha1 -> v1alpha3) and the downgrade
// (v1alpha3 -> v1alpha1) direction.
func TestClientSecretWoVersionConversion(t *testing.T) {
	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		clientv1alpha1.SchemeBuilder.AddToScheme,
		clientv1alpha2.SchemeBuilder.AddToScheme,
		clientv1alpha3.SchemeBuilder.AddToScheme,
		realmv1alpha1.SchemeBuilder.AddToScheme,
		realmv1alpha2.SchemeBuilder.AddToScheme,
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
		dst := &clientv1alpha3.Client{}
		if err := src.ConvertTo(dst); err != nil {
			t.Fatalf("cannot convert v1alpha1 to v1alpha3: %v", err)
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
		src := &clientv1alpha3.Client{}
		src.Name = "client"
		src.Spec.ForProvider.ClientSecretWoVersion = &version
		src.Status.AtProvider.ClientSecretWoVersion = &version

		dst := &clientv1alpha1.Client{}
		if err := dst.ConvertFrom(src); err != nil {
			t.Fatalf("cannot convert v1alpha3 to v1alpha1: %v", err)
		}
		if got := dst.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != 202605192 {
			t.Errorf("spec.forProvider.clientSecretWoVersion: want %v, got %v", 202605192, got)
		}
		if got := dst.Status.AtProvider.ClientSecretWoVersion; got == nil || *got != 202605192 {
			t.Errorf("status.atProvider.clientSecretWoVersion: want %v, got %v", 202605192, got)
		}
	})

	t.Run("StringVersionIdentityConversion", func(t *testing.T) {
		version := "202605192"
		src := &clientv1alpha2.Client{}
		src.Name = "client"
		src.Spec.ForProvider.ClientSecretWoVersion = &version
		src.Status.AtProvider.ClientSecretWoVersion = &version

		hub := &clientv1alpha3.Client{}
		if err := src.ConvertTo(hub); err != nil {
			t.Fatalf("cannot convert v1alpha2 to v1alpha3: %v", err)
		}
		if got := hub.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != version {
			t.Errorf("spec.forProvider.clientSecretWoVersion: want %q, got %v", version, got)
		}
		if got := hub.Status.AtProvider.ClientSecretWoVersion; got == nil || *got != version {
			t.Errorf("status.atProvider.clientSecretWoVersion: want %q, got %v", version, got)
		}

		dst := &clientv1alpha2.Client{}
		if err := dst.ConvertFrom(hub); err != nil {
			t.Fatalf("cannot convert v1alpha3 to v1alpha2: %v", err)
		}
		if got := dst.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != version {
			t.Errorf("spec.forProvider.clientSecretWoVersion: want %q, got %v", version, got)
		}
		if got := dst.Status.AtProvider.ClientSecretWoVersion; got == nil || *got != version {
			t.Errorf("status.atProvider.clientSecretWoVersion: want %q, got %v", version, got)
		}
	})

	t.Run("SingletonListsRoundTrip", func(t *testing.T) {
		const stored = `{
  "apiVersion": "openidclient.keycloak.crossplane.io/v1alpha2",
  "kind": "Client",
  "metadata": {"name": "client"},
  "spec": {
    "forProvider": {
      "authenticationFlowBindingOverrides": [{"browserId": "browser-flow", "directGrantId": "direct-flow"}],
      "authorization": [{"policyEnforcementMode": "PERMISSIVE"}]
    },
    "initProvider": {
      "authorization": [{"policyEnforcementMode": "DISABLED"}]
    }
  },
  "status": {
    "atProvider": {
      "authenticationFlowBindingOverrides": [{"browserId": "observed-browser-flow", "directGrantId": "observed-direct-flow"}],
      "authorization": [{"policyEnforcementMode": "ENFORCING"}]
    }
  }
}`
		src := &clientv1alpha2.Client{}
		if err := json.Unmarshal([]byte(stored), src); err != nil {
			t.Fatalf("cannot decode the stored object: %v", err)
		}

		hub := &clientv1alpha3.Client{}
		if err := src.ConvertTo(hub); err != nil {
			t.Fatalf("cannot convert v1alpha2 to v1alpha3: %v", err)
		}
		if got := hub.Spec.ForProvider.Authorization; got == nil || got.PolicyEnforcementMode == nil || *got.PolicyEnforcementMode != "PERMISSIVE" {
			t.Errorf("spec.forProvider.authorization: want embedded PERMISSIVE block, got %#v", got)
		}
		if got := hub.Spec.ForProvider.AuthenticationFlowBindingOverrides; got == nil || got.BrowserID == nil || *got.BrowserID != "browser-flow" {
			t.Errorf("spec.forProvider.authenticationFlowBindingOverrides: want embedded browser flow, got %#v", got)
		}
		if got := hub.Spec.InitProvider.Authorization; got == nil || got.PolicyEnforcementMode == nil || *got.PolicyEnforcementMode != "DISABLED" {
			t.Errorf("spec.initProvider.authorization: want embedded DISABLED block, got %#v", got)
		}
		if got := hub.Status.AtProvider.Authorization; got == nil || got.PolicyEnforcementMode == nil || *got.PolicyEnforcementMode != "ENFORCING" {
			t.Errorf("status.atProvider.authorization: want embedded ENFORCING block, got %#v", got)
		}

		dst := &clientv1alpha2.Client{}
		if err := dst.ConvertFrom(hub); err != nil {
			t.Fatalf("cannot convert v1alpha3 to v1alpha2: %v", err)
		}
		if got := dst.Spec.ForProvider.Authorization; len(got) != 1 || got[0].PolicyEnforcementMode == nil || *got[0].PolicyEnforcementMode != "PERMISSIVE" {
			t.Errorf("spec.forProvider.authorization: want one PERMISSIVE list item, got %#v", got)
		}
		if got := dst.Spec.InitProvider.Authorization; len(got) != 1 || got[0].PolicyEnforcementMode == nil || *got[0].PolicyEnforcementMode != "DISABLED" {
			t.Errorf("spec.initProvider.authorization: want one DISABLED list item, got %#v", got)
		}
		if got := dst.Status.AtProvider.AuthenticationFlowBindingOverrides; len(got) != 1 || got[0].BrowserID == nil || *got[0].BrowserID != "observed-browser-flow" {
			t.Errorf("status.atProvider.authenticationFlowBindingOverrides: want one observed list item, got %#v", got)
		}
	})

	t.Run("OldestClientVersionPreservesBothConversions", func(t *testing.T) {
		const stored = `{
  "apiVersion": "openidclient.keycloak.crossplane.io/v1alpha1",
  "kind": "Client",
  "metadata": {"name": "client"},
  "spec": {
    "forProvider": {
      "clientSecretWoVersion": 202605192,
      "authorization": [{"policyEnforcementMode": "PERMISSIVE"}]
    },
    "initProvider": {}
  },
  "status": {"atProvider": {}}
}`
		src := &clientv1alpha1.Client{}
		if err := json.Unmarshal([]byte(stored), src); err != nil {
			t.Fatalf("cannot decode the stored v1alpha1 object: %v", err)
		}

		hub := &clientv1alpha3.Client{}
		if err := src.ConvertTo(hub); err != nil {
			t.Fatalf("cannot convert v1alpha1 to v1alpha3: %v", err)
		}
		if got := hub.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != "202605192" {
			t.Errorf("clientSecretWoVersion: want string \"202605192\", got %#v", got)
		}
		if got := hub.Spec.ForProvider.Authorization; got == nil || got.PolicyEnforcementMode == nil || *got.PolicyEnforcementMode != "PERMISSIVE" {
			t.Errorf("authorization: want embedded PERMISSIVE block, got %#v", got)
		}

		dst := &clientv1alpha1.Client{}
		if err := dst.ConvertFrom(hub); err != nil {
			t.Fatalf("cannot convert v1alpha3 to v1alpha1: %v", err)
		}
		if got := dst.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != 202605192 {
			t.Errorf("clientSecretWoVersion: want integer 202605192, got %#v", got)
		}
		if got := dst.Spec.ForProvider.Authorization; len(got) != 1 || got[0].PolicyEnforcementMode == nil || *got[0].PolicyEnforcementMode != "PERMISSIVE" {
			t.Errorf("authorization: want one PERMISSIVE list item, got %#v", got)
		}
	})

	t.Run("NestedSingletonListsRoundTrip", func(t *testing.T) {
		const stored = `{
  "apiVersion": "realm.keycloak.crossplane.io/v1alpha1",
  "kind": "Realm",
  "metadata": {"name": "realm"},
  "spec": {
    "forProvider": {
      "realm": "realm",
      "smtpServer": [{"host": "smtp.example.test", "auth": [{"username": "sender"}]}]
    },
    "initProvider": {}
  },
  "status": {
    "atProvider": {
      "smtpServer": [{"host": "observed.example.test", "auth": [{"username": "observed-sender"}]}]
    }
  }
}`
		src := &realmv1alpha1.Realm{}
		if err := json.Unmarshal([]byte(stored), src); err != nil {
			t.Fatalf("cannot decode the stored Realm: %v", err)
		}

		hub := &realmv1alpha2.Realm{}
		if err := src.ConvertTo(hub); err != nil {
			t.Fatalf("cannot convert Realm v1alpha1 to v1alpha2: %v", err)
		}
		if got := hub.Spec.ForProvider.SMTPServer; got == nil || got.Auth == nil || got.Auth.Username == nil || *got.Auth.Username != "sender" {
			t.Errorf("spec.forProvider.smtpServer.auth: want embedded sender, got %#v", got)
		}

		dst := &realmv1alpha1.Realm{}
		if err := dst.ConvertFrom(hub); err != nil {
			t.Fatalf("cannot convert Realm v1alpha2 to v1alpha1: %v", err)
		}
		if got := dst.Spec.ForProvider.SMTPServer; len(got) != 1 || len(got[0].Auth) != 1 || got[0].Auth[0].Username == nil || *got[0].Auth[0].Username != "sender" {
			t.Errorf("spec.forProvider.smtpServer.auth: want nested singleton lists, got %#v", got)
		}
		if got := dst.Status.AtProvider.SMTPServer; len(got) != 1 || len(got[0].Auth) != 1 || got[0].Auth[0].Username == nil || *got[0].Auth[0].Username != "observed-sender" {
			t.Errorf("status.atProvider.smtpServer.auth: want nested singleton lists, got %#v", got)
		}
	})

	t.Run("MultiItemStoredListRequiresRepair", func(t *testing.T) {
		const stored = `{
  "apiVersion": "openidclient.keycloak.crossplane.io/v1alpha2",
  "kind": "Client",
  "metadata": {"name": "client"},
  "spec": {
    "forProvider": {
      "authorization": [
        {"policyEnforcementMode": "PERMISSIVE"},
        {"policyEnforcementMode": "ENFORCING"}
      ]
    },
    "initProvider": {}
  },
  "status": {"atProvider": {}}
}`
		src := &clientv1alpha2.Client{}
		if err := json.Unmarshal([]byte(stored), src); err != nil {
			t.Fatalf("cannot decode the stored object: %v", err)
		}
		err := src.ConvertTo(&clientv1alpha3.Client{})
		if err == nil {
			t.Fatal("expected a descriptive error for a stored singleton list with multiple items")
		}
		if !strings.Contains(err.Error(), "authorization") || !strings.Contains(err.Error(), "length of at most 1") {
			t.Errorf("expected a field-specific repair error, got %v", err)
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
	// still yield the v1alpha3 string.
	t.Run("StringEncodedV1Alpha1Upgrade", func(t *testing.T) {
		src := &clientv1alpha1.Client{}
		if err := json.Unmarshal([]byte(storedStringV1Alpha1Client), src); err != nil {
			t.Fatalf("cannot decode the string-encoded stored object: %v", err)
		}
		if got := src.Spec.ForProvider.ClientSecretWoVersion; got == nil || *got != 202605192 {
			t.Fatalf("spec.forProvider.clientSecretWoVersion: want %v, got %v", 202605192, got)
		}
		dst := &clientv1alpha3.Client{}
		if err := src.ConvertTo(dst); err != nil {
			t.Fatalf("cannot convert v1alpha1 to v1alpha3: %v", err)
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
