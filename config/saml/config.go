package saml

import (
	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane/upjet/pkg/config"
)

// Group is the short group name for the resources in this package
var Group = "saml"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {

	p.AddResourceConfigurator("keycloak_saml_client_default_scopes", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.SamlClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_saml_client_scope", func(r *config.Resource) {
		r.ShortGroup = Group
	})

}
