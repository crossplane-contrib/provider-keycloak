package config

import (
	"slices"
	"strings"
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

func TestSingletonListEmbeddingInventory(t *testing.T) {
	expected := map[string][]string{
		"keycloak_group_permissions":         {"v1alpha1"},
		"keycloak_ldap_user_federation":      {"v1alpha1"},
		"keycloak_openid_client":             {"v1alpha1", "v1alpha2"},
		"keycloak_openid_client_permissions": {"v1alpha1"},
		"keycloak_realm":                     {"v1alpha1"},
		"keycloak_realm_user_profile":        {"v1alpha1"},
		"keycloak_saml_client":               {"v1alpha1"},
		"keycloak_user":                      {"v1alpha1"},
		"keycloak_users_permissions":         {"v1alpha1"},
		"keycloak_workflow":                  {"v1alpha1"},
	}

	for name, getProvider := range map[string]func(bool) (*ujconfig.Provider, error){
		"cluster":    GetProvider,
		"namespaced": GetProviderNamespaced,
	} {
		t.Run(name, func(t *testing.T) {
			pc, err := getProvider(true)
			if err != nil {
				t.Fatalf("GetProvider: %v", err)
			}
			totalPaths := 0
			found := map[string][]string{}
			for resourceName, r := range pc.Resources {
				paths := r.CRDListConversionPaths()
				if len(paths) == 0 {
					continue
				}
				totalPaths += len(paths)
				for _, tfPath := range r.TFListConversionPaths() {
					schemaPath := strings.ReplaceAll(strings.ReplaceAll(tfPath, "[*]", ""), "[0]", "")
					option := r.SchemaElementOptions[schemaPath]
					if option == nil || !option.EmbeddedObject {
						t.Errorf("%s.%s: expected embedded-object schema option", resourceName, tfPath)
					}
				}
				found[resourceName] = r.PreviousVersions
				expectedCurrentVersion := "v1alpha2"
				if resourceName == "keycloak_openid_client" {
					expectedCurrentVersion = "v1alpha3"
				}
				if r.Version != expectedCurrentVersion {
					t.Errorf("%s: version: want %q, got %q", resourceName, expectedCurrentVersion, r.Version)
				}
				if r.CRDStorageVersion() != expectedCurrentVersion {
					t.Errorf("%s: storage version: want %q, got %q", resourceName, expectedCurrentVersion, r.CRDStorageVersion())
				}
				if r.CRDHubVersion() != expectedCurrentVersion {
					t.Errorf("%s: hub version: want %q, got %q", resourceName, expectedCurrentVersion, r.CRDHubVersion())
				}
			}
			if totalPaths != 36 {
				t.Errorf("singleton conversion path count: want 36, got %d", totalPaths)
			}
			if len(found) != len(expected) {
				t.Errorf("affected resource count: want %d, got %d", len(expected), len(found))
			}
			for resourceName, versions := range expected {
				if got, ok := found[resourceName]; !ok {
					t.Errorf("missing affected resource %q", resourceName)
				} else if !slices.Equal(got, versions) {
					t.Errorf("%s: previous versions: want %v, got %v", resourceName, versions, got)
				}
			}
		})
	}
}
