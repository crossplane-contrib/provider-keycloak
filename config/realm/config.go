package realm

import (
	"github.com/crossplane/upjet/pkg/config"
)

// Group is the short group name for the resources in this package
var Group = "realm"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_realm", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_required_action", func(r *config.Resource) {
		r.ShortGroup = Group
		r.Kind = "RequiredAction"
	})

	p.AddResourceConfigurator("keycloak_realm_keystore_rsa", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = Group
		if s, ok := r.TerraformResource.Schema["private_key"]; ok {
			s.Sensitive = true
		}
		if s, ok := r.TerraformResource.Schema["certificate"]; ok {
			s.Sensitive = true
		}
	})

	p.AddResourceConfigurator("keycloak_realm_user_profile", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_realm_events", func(r *config.Resource) {
		r.ShortGroup = Group
		r.Kind = "RealmEvents"
	})
}
