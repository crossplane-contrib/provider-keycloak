package role

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_role", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = "role"
		r.References["composite_roles"] = config.Reference{
			TerraformName: "keycloak_role",
		}
	})
}

var identifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
	OptionalParameters:           []string{"client_id"},
	GetIDByExternalName:          getIDByExternalName,
	GetIDByIdentifyingProperties: getIDByIdentifyingProperties,
}

// IdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var IdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(identifyingPropertiesLookup)

func getIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetRole(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetRoleByName(ctx, parameters["realm_id"].(string), parameters["client_id"].(string), parameters["name"].(string))
	if err != nil {
		return "", err
	}
	return found.Id, nil
}
