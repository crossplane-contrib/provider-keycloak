package user

import (
	"context"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_user", func(r *config.Resource) {
		r.ShortGroup = "user"

		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"required_actions", "initial_password.value", "initial_password.value", "initial_password.temporary"},
		}

	})

	p.AddResourceConfigurator("keycloak_user_groups", func(r *config.Resource) {
		r.ShortGroup = "user"

		r.References["user_id"] = config.Reference{
			TerraformName: "keycloak_user",
		}

		r.References["group_ids"] = config.Reference{
			TerraformName: "keycloak_group",
		}
	})

	p.AddResourceConfigurator("keycloak_user_roles", func(r *config.Resource) {
		r.ShortGroup = "user"

		r.References["user_id"] = config.Reference{
			TerraformName: "keycloak_user",
		}
	})

	p.AddResourceConfigurator("keycloak_users_permissions", func(r *config.Resource) {
		r.ShortGroup = "user"
	})
}

var userIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "username"},
	GetIDByExternalName:          getUserIDByExternalName,
	GetIDByIdentifyingProperties: getUserIDByIdentifyingProperties,
}

// UserIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var UserIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(userIdentifyingPropertiesLookup)

func getUserIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetUser(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getUserIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetUserByUsername(ctx, parameters["realm_id"].(string), parameters["username"].(string))
	if err != nil {
		return "", err
	}
	if found == nil {
		return "", nil
	}
	return found.Id, nil
}
