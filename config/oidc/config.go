package oidc

import (
	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane/upjet/pkg/config"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_oidc_identity_provider", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "oidc"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		r.References["first_broker_login_flow_alias"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/authenticationflow/v1alpha1.Flow",
			Extractor: common.PathAuthenticationFlowAliasExtractor,
		}
	})
}
