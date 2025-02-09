package identityprovider

import (
	"context"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane/upjet/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const (
	// Group is the short group for this provider.
	Group = "identityprovider"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_custom_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
	})
}

var identifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm", "identity_provider_alias", "name"},
	GetIDByExternalName:          getIDByExternalName,
	GetIDByIdentifyingProperties: getIDByIdentifyingProperties,
}

// IdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var IdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(identifyingPropertiesLookup)

func getIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetIdentityProviderMapper(ctx, parameters["realm"].(string), parameters["identity_provider_alias"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetIdentityProviderMappers(ctx, parameters["realm"].(string), parameters["identity_provider_alias"].(string))
	if err != nil {
		return "", err
	}

	filtered := lookup.Filter(found, func(mapper *keycloak.IdentityProviderMapper) bool {
		return mapper.Name == parameters["name"].(string)
	})

	return lookup.SingleOrEmpty(filtered, func(mapper *keycloak.IdentityProviderMapper) string {
		return mapper.Id
	})
}
