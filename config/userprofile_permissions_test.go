package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	tfschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestUserProfilePermissionsSchema(t *testing.T) {
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

			r, ok := p.Resources["keycloak_realm_user_profile"]
			if !ok {
				t.Fatal("keycloak_realm_user_profile: resource not registered in provider")
			}

			attribute, ok := r.TerraformResource.Schema["attribute"]
			if !ok {
				t.Fatal("attribute: field missing from schema")
			}
			attributeResource, ok := attribute.Elem.(*tfschema.Resource)
			if !ok {
				t.Fatalf("attribute: expected *schema.Resource element, got %T", attribute.Elem)
			}

			permissions, ok := attributeResource.Schema["permissions"]
			if !ok {
				t.Fatal("permissions: field missing from attribute schema")
			}
			if permissions.MaxItems != 1 {
				t.Fatalf("permissions.MaxItems = %d, want 1", permissions.MaxItems)
			}
		})
	}
}
