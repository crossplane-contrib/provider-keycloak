package ldap

import "github.com/crossplane/upjet/pkg/config"

var Group = "ldap"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {

	// ldap
	p.AddResourceConfigurator("keycloak_ldap_user_federation", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_ldap_user_attribute_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_role_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_group_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_role_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}
		r.References["role"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/role/v1alpha1.Role",
			Extractor: `github.com/crossplane/upjet/pkg/resource.ExtractParamPath("name", false)`,
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_group_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}
		r.References["group"] = config.Reference{
			Type:      "github.com/crossplane-contrib/provider-keycloak/apis/group/v1alpha1.Group",
			Extractor: `github.com/crossplane/upjet/pkg/resource.ExtractParamPath("name", false)`,
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_msad_user_account_control_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_msad_lds_user_account_control_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_attribute_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_full_name_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_custom_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/ldap/v1alpha1.UserFederation",
		}
	})
}
