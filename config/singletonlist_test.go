package config

import (
	"strings"
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// TestUserProfileRejectsMultiplePermissions guards against the perpetual
// reconciliation of crossplane-contrib/provider-keycloak#745.
func TestUserProfileRejectsMultiplePermissions(t *testing.T) {
	for name, get := range map[string]func(bool) (*ujconfig.Provider, error){"Cluster": GetProvider, "Namespaced": GetProviderNamespaced} {
		t.Run(name, func(t *testing.T) {
			pc, err := get(false)
			if err != nil {
				t.Fatalf("cannot get the provider configuration: %v", err)
			}
			params := map[string]any{"attribute": []any{map[string]any{
				"permissions": []any{
					map[string]any{"edit": []any{"admin"}},
					map[string]any{"view": []any{"admin", "user"}},
				},
			}}}
			inject := pc.Resources["keycloak_realm_user_profile"].TerraformConfigurationInjector
			if inject == nil {
				t.Fatal("singleton validation injector is not registered")
			}
			err = inject(map[string]any{}, params)
			if err == nil || !strings.Contains(err.Error(), "attribute[0].permissions has 2 entries") {
				t.Fatalf("expected the multi-entry permissions to be rejected, got %v", err)
			}
			params["attribute"] = []any{map[string]any{
				"permissions": []any{map[string]any{
					"edit": []any{"admin"},
					"view": []any{"admin", "user"},
				}},
			}}
			if err := inject(map[string]any{}, params); err != nil {
				t.Fatalf("valid single-entry permissions must still work: %v", err)
			}
		})
	}
}
