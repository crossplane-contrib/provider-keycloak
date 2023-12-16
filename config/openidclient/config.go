package openidclient

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_openid_client", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "openidclient"
	})

	p.AddResourceConfigurator("keycloak_openid_client_default_scopes", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "openidclient"
	})

	p.AddResourceConfigurator("keycloak_openid_client_scope", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "openidclient"
	})
}
