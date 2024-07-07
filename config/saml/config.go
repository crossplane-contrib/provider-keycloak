package saml

import "github.com/crossplane/upjet/pkg/config"

// Group is the short group name for the resources in this package
var Group = "saml"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_saml_identity_provider", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "saml"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
	})
	p.AddResourceConfigurator("keycloak_saml_client", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_saml_client_default_scopes", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_saml_client_scope", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
	})

}
