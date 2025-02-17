package group

import (
	"context"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane/upjet/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_group", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "group"

		r.References["parent_id"] = config.Reference{
			TerraformName: "keycloak_group",
		}
	})
	p.AddResourceConfigurator("keycloak_group_memberships", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "group"
		r.References["group_id"] = config.Reference{
			TerraformName: "keycloak_group",
		}

	})
	p.AddResourceConfigurator("keycloak_group_roles", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "group"
		r.References["group_id"] = config.Reference{
			TerraformName: "keycloak_group",
		}
	})
	p.AddResourceConfigurator("keycloak_group_permissions", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "group"
		r.References["group_id"] = config.Reference{
			TerraformName: "keycloak_group",
		}
	})
}

var groupIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
	GetIDByExternalName:          getGroupIDByExternalName,
	GetIDByIdentifyingProperties: getGroupIDByIdentifyingProperties,
}

// GroupIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var GroupIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(groupIdentifyingPropertiesLookup)

func getGroupIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetGroup(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getGroupIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetGroupByName(ctx, parameters["realm_id"].(string), parameters["name"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "no group with name") {
			return "", nil
		}

		return "", err
	}

	return found.Id, nil
}
