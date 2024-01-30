package defaults

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {

	p.AddResourceConfigurator("keycloak_default_roles", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = "defaults"
		r.References["default_roles"] = config.Reference{
			Type: "github.com/stakater/provider-keycloak/apis/role/v1alpha1.Role",
		}

	})
}
