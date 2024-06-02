package openid

import (
	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane/upjet/pkg/config"
)

// Group is the short group name for the resources in this package
var Group = "openid"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {

	p.AddResourceConfigurator("keycloak_openid_client_default_scopes", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
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
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
		// The id of the service account that is assigned the role (the service account of the client that "consumes" the role).
		r.References["service_account_user_id"] = config.Reference{
			Type:              "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor:         common.PathServiceAccountRoleIDExtractor,
			RefFieldName:      "ServiceAccountUserClientIDRef",
			SelectorFieldName: "ServiceAccountUserClientIDSelector",
		}
		// The name of the role that is assigned.
		r.References["role"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/role/v1alpha1.Role",
			Extractor: `github.com/crossplane/upjet/pkg/resource.ExtractParamPath("name", false)`,
		}
		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"service_account_user_id"},
		}

	})

	p.AddResourceConfigurator("keycloak_openid_client_service_account_realm_role", func(r *config.Resource) {
		r.ShortGroup = Group

		// The id of the service account that is assigned the role (the service account of the client that "consumes" the role).
		r.References["service_account_user_id"] = config.Reference{
			Type:              "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor:         common.PathServiceAccountRoleIDExtractor,
			RefFieldName:      "ServiceAccountUserClientIDRef",
			SelectorFieldName: "ServiceAccountUserClientIDSelector",
		}
		// The name of the role that is assigned.
		r.References["role"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/role/v1alpha1.Role",
			Extractor: `github.com/crossplane/upjet/pkg/resource.ExtractParamPath("name", false)`,
		}
		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"service_account_user_id"},
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_client_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["resource_server_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
		r.References["clients"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_role_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_user_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_client_permissions", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_openid_group_membership_protocol_mapper", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
		r.References["client_id"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/client/v1alpha1.OpenIdClient",
			Extractor: common.PathUUIDExtractor,
		}
	})

}
