package client

import (
	"github.com/crossplane/upjet/pkg/config"
)

const (
	// Group is the short group for this provider.
	Group = "client"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_openid_client", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
		r.Kind = "OpenIdClient"
		if s, ok := r.TerraformResource.Schema["client_secret"]; ok {
			s.Sensitive = true
		}
	})

	p.AddResourceConfigurator("keycloak_saml_client", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
		r.Kind = "SamlClient"
		if s, ok := r.TerraformResource.Schema["signing_private_key"]; ok {
			s.Sensitive = true
		}
		if s, ok := r.TerraformResource.Schema["signing_certificate"]; ok {
			s.Sensitive = true
		}
		if s, ok := r.TerraformResource.Schema["encryption_certificate"]; ok {
			s.Sensitive = true
		}
	})
}
