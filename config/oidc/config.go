package oidc

import (
	"context"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_oidc_identity_provider", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "oidc"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		r.References["organization_id"] = config.Reference{
			TerraformName: "keycloak_organization",
		}
		r.References["first_broker_login_flow_alias"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor: common.PathAuthenticationFlowAliasExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_oidc_google_identity_provider", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "oidc"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		r.References["first_broker_login_flow_alias"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor: common.PathAuthenticationFlowAliasExtractor,
		}

		if s, ok := r.TerraformResource.Schema["client_id"]; ok {
			s.Sensitive = true
		}
		if s, ok := r.TerraformResource.Schema["client_secret"]; ok {
			s.Sensitive = true
		}

	})
}

var identifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm", "alias"},
	GetIDByExternalName:          getIDByExternalName,
	GetIDByIdentifyingProperties: getIDByIdentifyingProperties,
}

// IdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var IdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(identifyingPropertiesLookup)

func getIDByExternalName(ctx context.Context, _ string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	return getIDByIdentifyingProperties(ctx, parameters, kcClient)
}

func getIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetIdentityProvider(ctx, parameters["realm"].(string), parameters["alias"].(string))
	if err != nil {
		return "", err
	}
	return found.Alias, nil
}
