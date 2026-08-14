package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
)

func TestClientAggregatePolicyReferences(t *testing.T) {
	want := map[string]struct {
		terraformName string
		extractor     string
	}{
		"aggregate_policies": {
			terraformName: "keycloak_openid_client_aggregate_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"client_policies": {
			terraformName: "keycloak_openid_client_client_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"client_scope_policies": {
			terraformName: "keycloak_openid_client_authorization_client_scope_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"generic_policies": {
			terraformName: "keycloak_generic_client_authorization_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"group_policies": {
			terraformName: "keycloak_openid_client_group_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"js_policies": {
			terraformName: "keycloak_openid_client_js_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"regex_policies": {
			terraformName: "keycloak_openid_client_regex_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"role_policies": {
			terraformName: "keycloak_openid_client_role_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"time_policies": {
			terraformName: "keycloak_openid_client_time_policy",
			extractor:     common.PathUUIDExtractor,
		},
		"user_policies": {
			terraformName: "keycloak_openid_client_user_policy",
			extractor:     common.PathUUIDExtractor,
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
			r, ok := p.Resources["keycloak_openid_client_aggregate_policy"]
			if !ok {
				t.Fatalf("keycloak_openid_client_aggregate_policy: resource not registered in provider")
			}

			for field, wantRef := range want {
				ref, ok := r.References[field]
				if !ok {
					t.Fatalf("missing reference configuration for %q", field)
				}
				if ref.TerraformName != wantRef.terraformName {
					t.Errorf("%s: TerraformName = %q, want %q", field, ref.TerraformName, wantRef.terraformName)
				}
				if ref.Extractor != wantRef.extractor {
					t.Errorf("%s: Extractor = %q, want %q", field, ref.Extractor, wantRef.extractor)
				}
			}
		})
	}
}
