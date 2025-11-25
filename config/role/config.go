package role

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
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
	clientID, _ := parameters["client_id"].(string)
	realmID, _ := parameters["realm_id"].(string)
	name, _ := parameters["name"].(string)
	
	found, err := kcClient.GetRoleByName(ctx, realmID, clientID, name)
	if err != nil {
		// If client_id is empty and we get a 404 error, it likely means we're looking for a client role
		// but the clientIdRef hasn't been resolved yet. Return empty string to signal the resource
		// cannot be identified yet, allowing the controller to retry after references are resolved.
		if clientID == "" {
			var apiErr *keycloak.ApiError
			if errors.As(err, &apiErr) && apiErr.Code == 404 {
				// Return empty to indicate the resource cannot be found/identified yet
				return "", nil
			}
		}
		return "", err
	}
	return found.Id, nil
}
