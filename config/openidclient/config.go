package openidclient

import (
	"context"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	n "github.com/crossplane/upjet/v2/pkg/types/name"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane-contrib/provider-keycloak/config/multitypes"
)

const (
	// Group is the short group for this provider.
	Group = "openidclient"
)

// clientConnectionDetails publishes an OpenID client's credentials as connection
// details under simplified, camelCase keys so consumers (e.g. Argo CD) can mount them
// directly from the connection secret. client_secret is a computed attribute for
// CONFIDENTIAL clients, so it is present in the Terraform state. Upjet already
// publishes the raw attribute.<name> variants; empty values are omitted here.
func clientConnectionDetails(attr map[string]any) (map[string][]byte, error) {
	conn := map[string][]byte{}
	if v, ok := attr["client_secret"].(string); ok && v != "" {
		conn["clientSecret"] = []byte(v)
	}
	if v, ok := attr["client_id"].(string); ok && v != "" {
		conn["clientID"] = []byte(v)
	}
	if v, ok := attr["service_account_user_id"].(string); ok && v != "" {
		conn["serviceAccountUserId"] = []byte(v)
	}
	return conn, nil
}

type syntheticListReference struct {
	name      string
	reference config.Reference
}

func addSyntheticListReferences(r *config.Resource, field string, refs ...syntheticListReference) {
	for _, ref := range refs {
		cp := *r.TerraformResource.Schema[field]
		r.TerraformResource.Schema[ref.name] = &cp
		r.References[ref.name] = ref.reference
	}

	ci := r.TerraformConfigurationInjector
	r.TerraformConfigurationInjector = func(jsonMap, tfMap map[string]any) error {
		if ci != nil {
			if err := ci(jsonMap, tfMap); err != nil {
				return err
			}
		}

		var union []any
		for _, ref := range refs {
			value := jsonMap[n.NewFromSnake(ref.name).LowerCamelComputed]
			if value == nil {
				continue
			}
			list, ok := value.([]any)
			if !ok {
				continue
			}
			union = append(union, list...)
			delete(tfMap, ref.name)
		}
		if union != nil {
			tfMap[field] = union
		}
		return nil
	}
}

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_openid_client", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group

		r.References["authentication_flow_binding_overrides.browser_id"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor:     common.PathUUIDExtractor,
		}
		r.References["authentication_flow_binding_overrides.direct_grant_id"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor:     common.PathUUIDExtractor,
		}

		// Skip late-initialization for the binding-override IDs so the
		// observed Terraform state never gets copied back into
		// spec.forProvider. This silences one of the two sources of
		// ArgoCD reconciliation drift on this field
		// (see docs/assessments/2026-04-client-forprovider-spec-drift.md).
		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{
				"authentication_flow_binding_overrides.browser_id",
				"authentication_flow_binding_overrides.direct_grant_id",
				// Prevent late-init of valid_redirect_uris and web_origins:
				// these are Optional+Computed fields that the Keycloak server
				// returns as empty lists. If late-init copies them back into
				// spec.forProvider, the Terraform provider's CustomizeDiff
				// rejects them with "valid_redirect_uris cannot be set when
				// standard or implicit flow is not enabled" (and similarly
				// for web_origins). See #416.
				"valid_redirect_uris",
				"web_origins",
			},
		}

		// schema.json types this as a number, but the runtime provider stores it as
		// a string; force string so late-init can unmarshal state (readNumberAsString).
		if s, ok := r.TerraformResource.Schema["client_secret_wo_version"]; ok {
			s.Type = schema.TypeString
		}

		// Publish the client's credentials as connection details.
		r.Sensitive.AdditionalConnectionDetailsFn = clientConnectionDetails
	})

	p.AddResourceConfigurator("keycloak_openid_client_default_scopes", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
		// Allow empty default_scopes to remove all default scopes from a client.
		// The Terraform provider handles empty sets correctly, but upjet generates
		// a required validation rule that rejects empty arrays.
		if s, ok := r.TerraformResource.Schema["default_scopes"]; ok {
			s.Required = false
			s.Optional = true
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_optional_scopes", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
		// Allow empty optional_scopes to remove all optional scopes from a client.
		// The Terraform provider handles empty sets correctly, but upjet generates
		// a required validation rule that rejects empty arrays.
		if s, ok := r.TerraformResource.Schema["optional_scopes"]; ok {
			s.Required = false
			s.Optional = true
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_scope", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_openid_client_service_account_role", func(r *config.Resource) {
		r.ShortGroup = Group
		//  The id of the client that provides the role.
		r.References["client_id"] = config.Reference{

			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
		// The id of the service account that is assigned the role (the service account of the client that "consumes" the role).
		r.References["service_account_user_id"] = config.Reference{
			TerraformName:     "keycloak_openid_client",
			Extractor:         common.PathServiceAccountRoleIDExtractor,
			RefFieldName:      "ServiceAccountUserClientIDRef",
			SelectorFieldName: "ServiceAccountUserClientIDSelector",
		}
		// The name of the role that is assigned.
		r.References["role"] = config.Reference{
			TerraformName: "keycloak_role",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("name", false)`,
		}
		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"service_account_user_id"},
		}

	})

	p.AddResourceConfigurator("keycloak_openid_client_service_account_realm_role", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["service_account_user_id"] = config.Reference{
			TerraformName:     "keycloak_openid_client",
			Extractor:         common.PathServiceAccountRoleIDExtractor,
			RefFieldName:      "ServiceAccountUserClientIDRef",
			SelectorFieldName: "ServiceAccountUserClientIDSelector",
		}

		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"service_account_user_id"},
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_client_policy", func(r *config.Resource) {
		r.ShortGroup = Group

		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}

		multitypes.ApplyToAsListWithOptions(r, "clients",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_clients",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "clients",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client",
					Extractor:     common.PathUUIDExtractor,
				},
			})

		if s, ok := r.TerraformResource.Schema["decisionStrategy"]; ok {
			s.Optional = false
			s.Computed = false
		}

		if s, ok := r.TerraformResource.Schema["logic"]; ok {
			s.Optional = false
			s.Computed = false
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_group_policy", func(r *config.Resource) {
		r.ShortGroup = Group

		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}

		r.References["groups.id"] = config.Reference{
			TerraformName: "keycloak_group",
			Extractor:     common.PathUUIDExtractor,
		}

		if s, ok := r.TerraformResource.Schema["decisionStrategy"]; ok {
			s.Optional = false
			s.Computed = false
		}

		if s, ok := r.TerraformResource.Schema["logic"]; ok {
			s.Optional = false
			s.Computed = false
		}

	})

	p.AddResourceConfigurator("keycloak_openid_client_role_policy", func(r *config.Resource) {
		r.ShortGroup = Group

		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}

		r.References["role.id"] = config.Reference{
			TerraformName: "keycloak_role",
			Extractor:     common.PathUUIDExtractor,
		}

		if s, ok := r.TerraformResource.Schema["decisionStrategy"]; ok {
			s.Optional = false
			s.Computed = false
		}

		if s, ok := r.TerraformResource.Schema["logic"]; ok {
			s.Optional = false
			s.Computed = false
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_user_policy", func(r *config.Resource) {
		r.ShortGroup = Group

		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}

		r.References["users"] = config.Reference{
			TerraformName: "keycloak_user",
			Extractor:     common.PathUUIDExtractor,
		}

		if s, ok := r.TerraformResource.Schema["decisionStrategy"]; ok {
			s.Optional = false
			s.Computed = false
		}

		if s, ok := r.TerraformResource.Schema["logic"]; ok {
			s.Optional = false
			s.Computed = false
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_permissions", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_authorization_resource", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_authorization_permission", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}

		r.References["resources"] = config.Reference{
			TerraformName: "keycloak_openid_client_authorization_resource",
			Extractor:     common.PathUUIDExtractor,
		}

		addSyntheticListReferences(r, "policies",
			syntheticListReference{
				name: "client_policies",
				reference: config.Reference{
					TerraformName: "keycloak_openid_client_client_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			syntheticListReference{
				name: "group_policies",
				reference: config.Reference{
					TerraformName: "keycloak_openid_client_group_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			syntheticListReference{
				name: "regex_policies",
				reference: config.Reference{
					TerraformName: "keycloak_openid_client_regex_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			syntheticListReference{
				name: "role_policies",
				reference: config.Reference{
					TerraformName: "keycloak_openid_client_role_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			syntheticListReference{
				name: "user_policies",
				reference: config.Reference{
					TerraformName: "keycloak_openid_client_user_policy",
					Extractor:     common.PathUUIDExtractor,
				},
			})
	})

	p.AddResourceConfigurator("keycloak_openid_client_authorization_scope", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_aggregate_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_js_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_time_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_authorization_client_scope_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_regex_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}

		if s, ok := r.TerraformResource.Schema["decisionStrategy"]; ok {
			s.Optional = false
			s.Computed = false
		}

		if s, ok := r.TerraformResource.Schema["logic"]; ok {
			s.Optional = false
			s.Computed = false
		}
	})
}

var clientIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "client_id"},
	GetIDByExternalName:          getClientIDByExternalName,
	GetIDByIdentifyingProperties: getClientIDByIdentifyingProperties,
}

// ClientIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var ClientIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(clientIdentifyingPropertiesLookup)

func getClientIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetGenericClient(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getClientIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetGenericClientByClientId(ctx, parameters["realm_id"].(string), parameters["client_id"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return "", nil
		}

		return "", err
	}
	return found.Id, nil
}

var clientScopeIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
	GetIDByExternalName:          getClientScopeIDByExternalName,
	GetIDByIdentifyingProperties: getClientScopeIDByIdentifyingProperties,
}

// ClientScopeIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var ClientScopeIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(clientScopeIdentifyingPropertiesLookup)

func getClientScopeIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientScope(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getClientScopeIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.ListOpenidClientScopesWithFilter(ctx, parameters["realm_id"].(string), func(scope *keycloak.OpenidClientScope) bool {
		return scope.Name == parameters["name"].(string)
	})

	if err != nil {
		return "", err
	}

	return lookup.SingleOrEmpty(found, func(scope *keycloak.OpenidClientScope) string {
		return scope.Id
	})
}

func getAuthzPolicyIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetClientAuthorizationPolicyByName(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), parameters["name"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "unable to find client authorization policy with name") {
			return "", nil
		}

		return "", err
	}
	return found.Id, nil
}

var authzClientPoliciesIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "resource_server_id", "name"},
	GetIDByExternalName:          getAuthzClientPoliciesIDByExternalName,
	GetIDByIdentifyingProperties: getAuthzClientPoliciesIDByIdentifyingProperties,
}

// AuthzClientPoliciesIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var AuthzClientPoliciesIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(authzClientPoliciesIdentifyingPropertiesLookup)

func getAuthzClientPoliciesIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientAuthorizationClientPolicy(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getAuthzClientPoliciesIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	return getAuthzPolicyIDByIdentifyingProperties(ctx, parameters, kcClient)
}

var authzGroupPoliciesIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "resource_server_id", "name"},
	GetIDByExternalName:          getAuthzGroupPoliciesIDByExternalName,
	GetIDByIdentifyingProperties: getAuthzGroupPoliciesIDByIdentifyingProperties,
}

// AuthzGroupPoliciesIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var AuthzGroupPoliciesIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(authzGroupPoliciesIdentifyingPropertiesLookup)

func getAuthzGroupPoliciesIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientAuthorizationGroupPolicy(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getAuthzGroupPoliciesIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	return getAuthzPolicyIDByIdentifyingProperties(ctx, parameters, kcClient)
}

var authzRolePoliciesIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "resource_server_id", "name"},
	GetIDByExternalName:          getAuthzRolePoliciesIDByExternalName,
	GetIDByIdentifyingProperties: getAuthzRolePoliciesIDByIdentifyingProperties,
}

// AuthzRolePoliciesIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var AuthzRolePoliciesIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(authzRolePoliciesIdentifyingPropertiesLookup)

func getAuthzRolePoliciesIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientAuthorizationRolePolicy(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getAuthzRolePoliciesIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	return getAuthzPolicyIDByIdentifyingProperties(ctx, parameters, kcClient)
}

var authzUserPoliciesIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "resource_server_id", "name"},
	GetIDByExternalName:          getAuthzUserPoliciesIDByExternalName,
	GetIDByIdentifyingProperties: getAuthzUserPoliciesIDByIdentifyingProperties,
}

// AuthzUserPoliciesIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var AuthzUserPoliciesIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(authzUserPoliciesIdentifyingPropertiesLookup)

func getAuthzUserPoliciesIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientAuthorizationUserPolicy(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getAuthzUserPoliciesIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	return getAuthzPolicyIDByIdentifyingProperties(ctx, parameters, kcClient)
}

var authzRegexPoliciesIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "resource_server_id", "name"},
	GetIDByExternalName:          getAuthzRegexPoliciesIDByExternalName,
	GetIDByIdentifyingProperties: getAuthzRegexPoliciesIDByIdentifyingProperties,
}

// AuthzRegexPoliciesIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var AuthzRegexPoliciesIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(authzRegexPoliciesIdentifyingPropertiesLookup)

func getAuthzRegexPoliciesIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientAuthorizationRegexPolicy(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getAuthzRegexPoliciesIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	return getAuthzPolicyIDByIdentifyingProperties(ctx, parameters, kcClient)
}

var authzResourceIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "resource_server_id", "name"},
	GetIDByExternalName:          getAuthzResourceIDByExternalName,
	GetIDByIdentifyingProperties: getAuthzResourceIDByIdentifyingProperties,
}

// AuthzResourceIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var AuthzResourceIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(authzResourceIdentifyingPropertiesLookup)

func getAuthzResourceIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientAuthorizationResource(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getAuthzResourceIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientAuthorizationResourceByName(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), parameters["name"].(string))
	if err != nil {
		return "", err
	}
	if found == nil {
		return "", nil
	}
	return found.Id, nil
}

var authzPermissionIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "resource_server_id", "name"},
	GetIDByExternalName:          getAuthzPermissionIDByExternalName,
	GetIDByIdentifyingProperties: getAuthzPermissionIDByIdentifyingProperties,
}

// AuthzPermissionIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var AuthzPermissionIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(authzPermissionIdentifyingPropertiesLookup)

func getAuthzPermissionIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOpenidClientAuthorizationPermission(ctx, parameters["realm_id"].(string), parameters["resource_server_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getAuthzPermissionIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	// Permissions are stored as authorization policies in Keycloak, so they can
	// be resolved by name through the policy endpoint.
	return getAuthzPolicyIDByIdentifyingProperties(ctx, parameters, kcClient)
}
