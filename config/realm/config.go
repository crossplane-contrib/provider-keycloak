package realm

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_realm", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = "realm"
	})

	p.AddResourceConfigurator("keycloak_required_action", func(r *config.Resource) {
		r.ShortGroup = "realm"
		r.Kind = "RequiredAction"
	})
}
