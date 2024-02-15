package openidgroup

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_openid_group_membership_protocol_mapper", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "openidgroup"

		r.References["client_scope_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/openidclient/v1alpha1.ClientScope",
		}
	})
}
