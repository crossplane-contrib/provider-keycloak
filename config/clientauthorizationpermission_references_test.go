package config

import (
	"reflect"
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
)

func TestClientAuthorizationPermissionReferences(t *testing.T) {
	want := map[string]struct {
		terraformName string
		extractor     string
	}{
		"resources": {
			terraformName: "keycloak_openid_client_authorization_resource",
			extractor:     common.PathUUIDExtractor,
		},
		"scopes": {
			terraformName: "keycloak_openid_client_authorization_scope",
			extractor:     common.PathUUIDExtractor,
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
			r, ok := p.Resources["keycloak_openid_client_authorization_permission"]
			if !ok {
				t.Fatalf("keycloak_openid_client_authorization_permission: resource not registered in provider")
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

func TestClientAuthorizationPermissionPolicyInjector(t *testing.T) {
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
			r, ok := p.Resources["keycloak_openid_client_authorization_permission"]
			if !ok {
				t.Fatalf("keycloak_openid_client_authorization_permission: resource not registered in provider")
			}
			if r.TerraformConfigurationInjector == nil {
				t.Fatal("TerraformConfigurationInjector is nil")
			}

			jsonMap := map[string]any{
				"clientPolicies": []any{"client-policy-id"},
				"userPolicies":   []any{"user-policy-id"},
			}
			tfMap := map[string]any{
				"policies":        []any{"raw-policy-id"},
				"client_policies": []any{"client-policy-id"},
				"user_policies":   []any{"user-policy-id"},
			}
			if err := r.TerraformConfigurationInjector(jsonMap, tfMap); err != nil {
				t.Fatalf("injecting synthetic policy references: %v", err)
			}
			if diff := reflect.DeepEqual(tfMap["policies"], []any{"client-policy-id", "user-policy-id"}); !diff {
				t.Fatalf("policies not consolidated as expected: got %v", tfMap["policies"])
			}
			if _, ok := tfMap["client_policies"]; ok {
				t.Fatalf("client_policies was not removed from tfMap: %v", tfMap)
			}
			if _, ok := tfMap["user_policies"]; ok {
				t.Fatalf("user_policies was not removed from tfMap: %v", tfMap)
			}

			rawJSONMap := map[string]any{}
			rawTFMap := map[string]any{
				"policies": []any{"raw-policy-id"},
			}
			if err := r.TerraformConfigurationInjector(rawJSONMap, rawTFMap); err != nil {
				t.Fatalf("injecting raw policy IDs: %v", err)
			}
			if !reflect.DeepEqual(rawTFMap["policies"], []any{"raw-policy-id"}) {
				t.Fatalf("raw policies should remain unchanged, got %v", rawTFMap["policies"])
			}
		})
	}
}
