package identityprovider

import (
	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane/upjet/pkg/config"
)

const (
	// Group is the short group for this provider.
	Group = "idp"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_oidc_identity_provider", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
		r.Kind = "OpenIdIdentityProvider"
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_saml_identity_provider", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	// TODO: Add SAML variant
	p.AddResourceConfigurator("keycloak_custom_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		// The alias of the associated identity provider.
		r.References["identity_provider_alias"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/idp/v1alpha1.OpenIdIdentityProvider",
			Extractor: `github.com/crossplane/upjet/pkg/resource.ExtractParamPath("alias", false)`,
		}
	})
}
