package defaults

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {

	p.AddResourceConfigurator("keycloak_default_roles", func(r *config.Resource) {
		r.ShortGroup = "defaults"
		r.References["default_roles"] = config.Reference{
			TerraformName: "keycloak_role",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("name", false)`,
		}

	})

	p.AddResourceConfigurator("keycloak_default_groups", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = "defaults"
		r.Kind = "DefaultGroups"
		r.References["group_ids"] = config.Reference{
			TerraformName: "keycloak_group",
		}

	})
}

// DefaultGroupsIdentifier derives the Terraform ID {realm}/default-groups from
// spec.forProvider.realmId, so it is known before creation and an interrupted
// create can be retried safely. With realmInternalId instead of realmId, the
// ID is taken from the provider like with config.IdentifierFromProvider.
var DefaultGroupsIdentifier = func() config.ExternalName {
	e := config.IdentifierFromProvider
	e.GetIDFn = func(_ context.Context, externalName string, parameters map[string]any, _ map[string]any) (string, error) {
		if realmID, ok := parameters["realm_id"].(string); ok && realmID != "" {
			return realmID + "/default-groups", nil
		}
		return externalName, nil
	}
	return e
}()
