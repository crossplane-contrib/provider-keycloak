package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
)

func TestGroupAdminPermissionsReferences(t *testing.T) {
	want := map[string]struct {
		terraformName string
		extractor     string
	}{
		"group_ids": {
			terraformName: "keycloak_group",
		},
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
			r, ok := p.Resources["keycloak_group_admin_permissions"]
			if !ok {
				t.Fatalf("keycloak_group_admin_permissions: resource not registered in provider")
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

			// The original policies field stays settable for raw IDs of
			// policy types that are not (yet) exposed as managed resources.
			if _, ok := r.References["policies"]; ok {
				t.Errorf("policies: expected no reference configuration on the original field")
			}
			s, ok := r.TerraformResource.Schema["policies"]
			if !ok {
				t.Fatalf("policies: field missing from schema")
			}
			if s.Computed {
				t.Errorf("policies: field is computed, expected it to stay settable")
			}
		})
	}
}
