package user

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane-contrib/provider-keycloak/config/multitypes"
)

const shortGroup = "user"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_user", func(r *config.Resource) {
		r.ShortGroup = shortGroup

		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"required_actions", "initial_password.value", "initial_password.value", "initial_password.temporary"},
		}

	})

	p.AddResourceConfigurator("keycloak_user_groups", func(r *config.Resource) {
		r.ShortGroup = shortGroup

		r.References["user_id"] = config.Reference{
			TerraformName: "keycloak_user",
		}

		r.References["group_ids"] = config.Reference{
			TerraformName: "keycloak_group",
		}
	})

	p.AddResourceConfigurator("keycloak_user_roles", func(r *config.Resource) {
		r.ShortGroup = shortGroup

		r.References["user_id"] = config.Reference{
			TerraformName: "keycloak_user",
		}
	})

	p.AddResourceConfigurator("keycloak_users_permissions", func(r *config.Resource) {
		r.ShortGroup = shortGroup
	})

	p.AddResourceConfigurator("keycloak_users_admin_permissions", func(r *config.Resource) {
		r.ShortGroup = shortGroup

		// policies is a single Terraform field holding the IDs of arbitrary
		// authorization policies living on the realm's admin-permissions
		// client. Expose one strongly-typed list field per referenceable
		// policy type; the values are consolidated back into policies before
		// they are sent to Terraform. The original policies field stays
		// settable for raw IDs of policy types that have no managed resource
		// yet.
		multitypes.ApplyToAsListWithOptions(r, "policies",
			&multitypes.Options{KeepOriginalField: true},
			multitypes.Instance{
				Name: "policies",
			},
			multitypes.Instance{
				Name: "aggregate_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_aggregate_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "client_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_client_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "client_scope_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_authorization_client_scope_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "group_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_group_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "js_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_js_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "regex_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_regex_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "role_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_role_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "time_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_time_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "user_policies",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_user_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			})
	})

	p.AddResourceConfigurator("keycloak_custom_user_federation", func(r *config.Resource) {
		r.ShortGroup = shortGroup
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
