package role

import "github.com/upbound/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_role", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = "role"
		r.References["composite_roles"] = config.Reference{
			Type: "Role",
		}
		r.References["client_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/openidclient/v1alpha1.Client",
		}
	})
}
