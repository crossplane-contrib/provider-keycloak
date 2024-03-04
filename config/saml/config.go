package saml

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_saml_identity_provider", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "saml"
		r.References["realm"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm",
		}
	})

}
