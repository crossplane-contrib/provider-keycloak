package mapper

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_generic_protocol_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"
		r.References["client_scope_id"] = config.Reference{
			TerraformName: "keycloak_openid_client_scope",
		}
	})

	p.AddResourceConfigurator("keycloak_generic_role_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"
		r.References["role_id"] = config.Reference{
			TerraformName: "keycloak_role",
		}

	})
}
