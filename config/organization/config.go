package organization

import (
	"context"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane-contrib/provider-keycloak/config/multitypes"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_organization", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "organization"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
	})
	p.AddResourceConfigurator("keycloak_organization_memberships", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "organization"
		r.References["organization_id"] = config.Reference{
			TerraformName: "keycloak_organization",
		}
		// members holds usernames, so resolve them from the referenced User's username.
		r.References["members"] = config.Reference{
			TerraformName: "keycloak_user",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("username", false)`,
		}
	})
	p.AddResourceConfigurator("keycloak_organization_admin_permissions", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "organization"
		r.Kind = "AdminPermissions"

		r.References["organization_ids"] = config.Reference{
			TerraformName: "keycloak_organization",
		}

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
}

var organizationIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm", "name", "domain"},
	GetIDByExternalName:          getOrganizationIDByExternalName,
	GetIDByIdentifyingProperties: getOrganizationIDByIdentifyingProperties,
}

// OrganizationIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var OrganizationIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(organizationIdentifyingPropertiesLookup)

func getOrganizationIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOrganization(ctx, parameters["realm"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getOrganizationIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOrganizationByName(ctx, parameters["realm"].(string), parameters["name"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "organization with name") {
			return "", nil
		}

		return "", err
	}

	return found.Id, nil
}
