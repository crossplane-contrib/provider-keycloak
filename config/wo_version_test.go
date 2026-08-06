package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestClientSecretWoVersionIsString locks in the client_secret_wo_version type
// override. The embedded generation schema (config/schema.json, dumped from an
// older published provider release) types this field as a number, while the
// runtime terraform-provider-keycloak SDK declares it schema.TypeString and
// stores it in Terraform state as a quoted string. Without the override, upjet
// generates a *float64 field and late-initialization fails to unmarshal the
// observed state ("readNumberAsString: invalid number"), stalling every
// reconcile of these resources at observe.
func TestClientSecretWoVersionIsString(t *testing.T) {
	resources := []string{
		"keycloak_oidc_identity_provider",
		"keycloak_openid_client",
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
			for _, tfName := range resources {
				r, ok := p.Resources[tfName]
				if !ok {
					t.Errorf("%s: resource not registered in provider", tfName)
					continue
				}
				s, ok := r.TerraformResource.Schema["client_secret_wo_version"]
				if !ok {
					t.Errorf("%s: client_secret_wo_version not present in schema", tfName)
					continue
				}
				if s.Type != schema.TypeString {
					t.Errorf("%s: client_secret_wo_version type = %v, want %v — a number type makes late-init fail with readNumberAsString on the string state value",
						tfName, s.Type, schema.TypeString)
				}
			}
		})
	}
}
