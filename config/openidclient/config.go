package openidclient

import (
	"context"
	"errors"
	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/utils"
	"github.com/crossplane/upjet/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
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
			Extractor:     `github.com/crossplane/upjet/pkg/resource.ExtractParamPath("name", false)`,
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
}

// IdentifierLookupForOidcClient is used to find the existing resource by it´s identifying properties
var IdentifierLookupForOidcClient = config.ExternalName{
	SetIdentifierArgumentFn: config.NopSetIdentifierArgument,
	GetExternalNameFn:       config.IDAsExternalName,
	GetIDFn:                 getIdFromOidcClientProperties,
	DisableNameInitializer:  true,
}

func getIdFromOidcClientProperties(ctx context.Context, externalName string, parameters map[string]any, terraformProviderConfig map[string]any) (string, error) {
	kcClient, err := utils.NewKeycloakClient(ctx, terraformProviderConfig)
	if err != nil {
		return "", err
	}

	realmID, realmIdExists := parameters["realm_id"]
	if !realmIdExists {
		return "", errors.New("realmId not set")
	}

	clientID, clientIdExists := parameters["client_id"]
	if !clientIdExists {
		return "", errors.New("clientId not set")
	}

	if externalName != "" {
		found, err := kcClient.GetGenericClient(ctx, realmID.(string), externalName)
		if err != nil {
			var apiErr *keycloak.ApiError
			if !(errors.As(err, &apiErr) && apiErr.Code == 404) {
				return "", err
			}
		} else {
			return found.Id, nil
		}
	}

	found, err := kcClient.GetGenericClientByClientId(ctx, realmID.(string), clientID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return "", nil
		}

		return "", err
	}

	return found.Id, nil
}
