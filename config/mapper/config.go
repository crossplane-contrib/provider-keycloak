package mapper

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_generic_protocol_mapper", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "client"
	})
	p.AddResourceConfigurator("keycloak_generic_role_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"

		r.References["role_id"] = config.Reference{
			Type: "github.com/stakater/provider-keycloak/apis/role/v1alpha1.Role",
		}
	})
}
