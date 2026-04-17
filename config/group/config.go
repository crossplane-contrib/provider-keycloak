package group

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
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
	OptionalParameters:           []string{"parent_id"},
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
	realmID := parameters["realm_id"].(string)
	name := parameters["name"].(string)
	parentID, _ := parameters["parent_id"].(string)

	groups, err := kcClient.ListGroupsWithName(ctx, realmID, name)
	if err != nil {
		return "", err
	}

	group := findGroupByNameAndParent(name, parentID, groups, "")
	if group == nil {
		return "", nil
	}

	return group.Id, nil
}

// findGroupByNameAndParent walks the group tree returned by the Keycloak search API
// and finds the group that matches both the given name and parent ID.
// For top-level groups, parentID and currentParentID are both empty strings.
// For child groups, parentID is the expected parent's UUID.
func findGroupByNameAndParent(name, parentID string, groups []*keycloak.Group, currentParentID string) *keycloak.Group {
	for _, group := range groups {
		if group.Name == name && currentParentID == parentID {
			return group
		}
		found := findGroupByNameAndParent(name, parentID, group.SubGroups, group.Id)
		if found != nil {
			return found
		}
	}
	return nil
}
