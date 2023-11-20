package openidgroup

import "github.com/upbound/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_openid_group_membership_protocol_mapper", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "openidgroup"
	})
}
