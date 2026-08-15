package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

func TestGroupAdminPermissionsReferences(t *testing.T) {
	want := map[string]struct {
		terraformName string
		extractor     string
	}{
		"group_ids": {
			terraformName: "keycloak_group",
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

			for _, field := range []string{
				"policies",
				"aggregate_policies",
				"client_policies",
				"client_scope_policies",
				"group_policies",
				"js_policies",
				"regex_policies",
				"role_policies",
				"time_policies",
				"user_policies",
			} {
				if _, ok := r.References[field]; ok {
					t.Errorf("%s: expected no reference configuration", field)
				}
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
