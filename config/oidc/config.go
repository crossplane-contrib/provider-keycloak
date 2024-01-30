package oidc

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_oidc_identity_provider", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "oidc"
		r.References["realm"] = config.Reference{
			Type: "github.com/stakater/provider-keycloak/apis/realm/v1alpha1.Realm",
		}
	})

}
