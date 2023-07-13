package mapper

import "github.com/upbound/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_generic_protocol_mapper", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "client"
	})
	p.AddResourceConfigurator("keycloak_generic_role_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"

		r.References["client_id"] = config.Reference{
			Type: "github.com/corewire/apis/openidclient/v1alpha1.Client",
		}

		r.References["role_id"] = config.Reference{
			Type: "github.com/corewire/apis/role/v1alpha1.Role",
		}
	})
}
