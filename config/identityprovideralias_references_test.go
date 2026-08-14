package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
)

// wantIdentityProviderAliasReferences returns the expected reference target for every
// alias field of the given base field name.
func wantIdentityProviderAliasReferences(field string) map[string]string {
	return map[string]string{
		field:                   "keycloak_oidc_identity_provider",
		"saml_" + field:         "keycloak_saml_identity_provider",
		"google_" + field:       "keycloak_oidc_google_identity_provider",
		"github_" + field:       "keycloak_oidc_github_identity_provider",
		"facebook_" + field:     "keycloak_oidc_facebook_identity_provider",
		"microsoft_" + field:    "keycloak_oidc_microsoft_identity_provider",
		"kubernetes_" + field:   "keycloak_kubernetes_identity_provider",
		"spiffe_" + field:       "keycloak_spiffe_identity_provider",
		"openshift_v4_" + field: "keycloak_oidc_openshift_v4_identity_provider",
	}
}

func TestIdentityProviderAliasReferences(t *testing.T) {
	resources := map[string]string{
		"keycloak_identity_provider_token_exchange_scope_permission": "provider_alias",
		"keycloak_custom_identity_provider_mapper":                   "identity_provider_alias",
		"keycloak_attribute_importer_identity_provider_mapper":       "identity_provider_alias",
		"keycloak_attribute_to_role_identity_provider_mapper":        "identity_provider_alias",
		"keycloak_hardcoded_attribute_identity_provider_mapper":      "identity_provider_alias",
		"keycloak_hardcoded_group_identity_provider_mapper":          "identity_provider_alias",
		"keycloak_hardcoded_role_identity_provider_mapper":           "identity_provider_alias",
		"keycloak_user_template_importer_identity_provider_mapper":   "identity_provider_alias",
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

			for name, field := range resources {
				r, ok := p.Resources[name]
				if !ok {
					t.Fatalf("%s: resource not registered in provider", name)
				}

				for aliasField, wantTerraformName := range wantIdentityProviderAliasReferences(field) {
					ref, ok := r.References[aliasField]
					if !ok {
						t.Errorf("%s: missing reference configuration for %q", name, aliasField)
						continue
					}
					if ref.TerraformName != wantTerraformName {
						t.Errorf("%s.%s: TerraformName = %q, want %q", name, aliasField, ref.TerraformName, wantTerraformName)
					}
					if ref.Extractor != common.PathIdentityProviderAliasExtractor {
						t.Errorf("%s.%s: Extractor = %q, want %q", name, aliasField, ref.Extractor, common.PathIdentityProviderAliasExtractor)
					}
				}

				// The original field must stay settable in spec.forProvider for backward compatibility.
				if s, ok := r.TerraformResource.Schema[field]; !ok || (!s.Optional && !s.Required) {
					t.Errorf("%s: original field %q is no longer settable", name, field)
				}
			}
		})
	}
}
