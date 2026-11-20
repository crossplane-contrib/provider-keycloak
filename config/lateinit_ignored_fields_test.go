package config

import (
	"slices"
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
)

// TestLateInitIgnoredFields locks in the LateInitializer.IgnoredFields
// configuration that prevents spec.forProvider drift on Optional+Computed
// reference targets (the bug fixed in this branch — see
// docs/assessments/2026-04-client-forprovider-spec-drift.md).
//
// This is a unit-level regression test for the drift fix: if a future
// refactor drops one of these IgnoredFields entries, the generated
// LateInitialize method would silently start copying server-observed
// values back into spec.forProvider again and ArgoCD would loop.
//
// The existing e2e suite (uptest) cannot catch this — it only checks
// Ready/Synced and clean deletion; it does not diff spec.forProvider
// before vs. after reconcile, and the affected leaf fields are not even
// exercised by the demos under dev/demos/{basic,namespaced}/. So we lock
// the configuration in here instead.
func TestLateInitIgnoredFields(t *testing.T) {
	want := map[string][]string{
		"keycloak_openid_client": {
			"authentication_flow_binding_overrides.browser_id",
			"authentication_flow_binding_overrides.direct_grant_id",
			"valid_redirect_uris",
			"web_origins",
		},
		"keycloak_saml_client": {
			"authentication_flow_binding_overrides.browser_id",
			"authentication_flow_binding_overrides.direct_grant_id",
		},
		"keycloak_role": {
			"composite_roles",
		},
		"keycloak_authentication_bindings": {
			"browser_flow",
			"registration_flow",
			"direct_grant_flow",
			"reset_credentials_flow",
			"client_authentication_flow",
			"docker_authentication_flow",
			"first_broker_login_flow",
		},
		"keycloak_realm": {
			"browser_flow",
			"registration_flow",
			"direct_grant_flow",
			"reset_credentials_flow",
			"client_authentication_flow",
			"docker_authentication_flow",
			"first_broker_login_flow",
			"web_authn_policy.signature_algorithms",
			"web_authn_policy.signature_algorithms_ordered",
			"web_authn_passwordless_policy.signature_algorithms",
			"web_authn_passwordless_policy.signature_algorithms_ordered",
		},
		"keycloak_generic_protocol_mapper": {
			"config",
		},
	}

	flavours := map[string]func() (*ujconfig.Provider, error){
		"cluster":    func() (*ujconfig.Provider, error) { return GetProvider(true) },
		"namespaced": func() (*ujconfig.Provider, error) { return GetProviderNamespaced(true) },
	}

	for flavourName, get := range flavours {
		t.Run(flavourName, func(t *testing.T) {
			p, err := get()
			if err != nil {
				t.Fatalf("loading provider: %v", err)
			}
			for tfName, fields := range want {
				r, ok := p.Resources[tfName]
				if !ok {
					t.Errorf("%s: resource not registered in provider", tfName)
					continue
				}
				got := r.LateInitializer.IgnoredFields
				for _, f := range fields {
					if !slices.Contains(got, f) {
						t.Errorf("%s: LateInitializer.IgnoredFields missing %q (got %v) — would re-introduce ArgoCD drift on this field",
							tfName, f, got)
					}
				}
			}
		})
	}
}

// TestRealmReferenceLateInitIgnored ensures that no realm-scoped resource
// late-initializes realm_id/realm or realm_internal_id. They are mutually
// exclusive (ExactlyOneOf) and both Optional+Computed, so late-initializing
// the unconfigured one would make every subsequent apply fail.
func TestRealmReferenceLateInitIgnored(t *testing.T) {
	flavours := map[string]func() (*ujconfig.Provider, error){
		"cluster":    func() (*ujconfig.Provider, error) { return GetProvider(true) },
		"namespaced": func() (*ujconfig.Provider, error) { return GetProviderNamespaced(true) },
	}

	for flavourName, get := range flavours {
		t.Run(flavourName, func(t *testing.T) {
			p, err := get()
			if err != nil {
				t.Fatalf("loading provider: %v", err)
			}
			checked := 0
			for tfName, r := range p.Resources {
				if _, ok := r.TerraformResource.Schema["realm_internal_id"]; !ok {
					continue
				}
				checked++
				realmField := "realm_id"
				if _, ok := r.TerraformResource.Schema[realmField]; !ok {
					realmField = "realm"
				}
				for _, f := range []string{"realm_internal_id", realmField} {
					if !slices.Contains(r.LateInitializer.IgnoredFields, f) {
						t.Errorf("%s: LateInitializer.IgnoredFields missing %q (got %v)", tfName, f, r.LateInitializer.IgnoredFields)
					}
				}
			}
			if checked == 0 {
				t.Fatal("no resource with realm_internal_id found")
			}
		})
	}
}

func TestOpenIDClientAuthenticationFlowBindingOverrideExtractors(t *testing.T) {
	flavours := map[string]func() (*ujconfig.Provider, error){
		"cluster":    func() (*ujconfig.Provider, error) { return GetProvider(true) },
		"namespaced": func() (*ujconfig.Provider, error) { return GetProviderNamespaced(true) },
	}

	for flavourName, get := range flavours {
		t.Run(flavourName, func(t *testing.T) {
			p, err := get()
			if err != nil {
				t.Fatalf("loading provider: %v", err)
			}
			r, ok := p.Resources["keycloak_openid_client"]
			if !ok {
				t.Fatalf("keycloak_openid_client: resource not registered in provider")
			}

			for _, field := range []string{
				"authentication_flow_binding_overrides.browser_id",
				"authentication_flow_binding_overrides.direct_grant_id",
			} {
				ref, ok := r.References[field]
				if !ok {
					t.Fatalf("keycloak_openid_client: reference %q not configured", field)
				}
				if ref.Extractor != common.PathUUIDExtractor {
					t.Errorf("keycloak_openid_client: reference %q extractor = %q, want %q",
						field, ref.Extractor, common.PathUUIDExtractor)
				}
			}
		})
	}
}

func TestExecutionConfigExecutionIDExtractor(t *testing.T) {
	flavours := map[string]func() (*ujconfig.Provider, error){
		"cluster":    func() (*ujconfig.Provider, error) { return GetProvider(true) },
		"namespaced": func() (*ujconfig.Provider, error) { return GetProviderNamespaced(true) },
	}

	for flavourName, get := range flavours {
		t.Run(flavourName, func(t *testing.T) {
			p, err := get()
			if err != nil {
				t.Fatalf("loading provider: %v", err)
			}
			r, ok := p.Resources["keycloak_authentication_execution_config"]
			if !ok {
				t.Fatalf("keycloak_authentication_execution_config: resource not registered in provider")
			}

			ref, ok := r.References["execution_id"]
			if !ok {
				t.Fatalf("keycloak_authentication_execution_config: reference %q not configured", "execution_id")
			}
			if ref.Extractor != common.PathUUIDExtractor {
				t.Errorf("keycloak_authentication_execution_config: reference %q extractor = %q, want %q",
					"execution_id", ref.Extractor, common.PathUUIDExtractor)
			}
		})
	}
}
