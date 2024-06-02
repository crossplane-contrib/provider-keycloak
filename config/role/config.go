package role

import (
	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane/upjet/pkg/config"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_role", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = "role"
		r.References["composite_roles"] = config.Reference{
			Type: "Role",
		}
		// TODO: Add Saml Variant
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})
}
