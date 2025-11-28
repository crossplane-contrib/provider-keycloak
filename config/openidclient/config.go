package openidclient

import (
	"context"
	"strings"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
)

const (
	// Group is the short group for this provider.
	Group = "openidclient"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_openid_client", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group

		r.References["authentication_flow_binding_overrides.browser_id"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
		}
		r.References["authentication_flow_binding_overrides.direct_grant_id"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
		}

	})

	p.AddResourceConfigurator("keycloak_openid_client_default_scopes", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_openid_client_optional_scopes", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
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
	})

	p.AddResourceConfigurator("keycloak_openid_client_client_policy", func(r *config.Resource) {
		r.ShortGroup = Group

		r.References["clients"] = config.Reference{
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

	p.AddResourceConfigurator("keycloak_openid_client_group_policy", func(r *config.Resource) {
		r.ShortGroup = Group

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

		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"groups.id"},
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_role_policy", func(r *config.Resource) {
		r.ShortGroup = Group

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

		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"role.id"},
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_user_policy", func(r *config.Resource) {
		r.ShortGroup = Group

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
	})

	p.AddResourceConfigurator("keycloak_openid_client_authorization_resource", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_openid_client_authorization_permission", func(r *config.Resource) {
		r.ShortGroup = Group
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
