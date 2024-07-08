package ldap

import "github.com/crossplane/upjet/pkg/config"

// Group is the short group name for the resources in this package
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
			TerraformName: "keycloak_ldap_user_federation",
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_role_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_group_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_role_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
		r.References["role"] = config.Reference{
			TerraformName: "keycloak_role",
			Extractor:     `github.com/crossplane/upjet/pkg/resource.ExtractParamPath("name", false)`,
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_group_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
		r.References["group"] = config.Reference{
			TerraformName: "keycloak_group",
			Extractor:     `github.com/crossplane/upjet/pkg/resource.ExtractParamPath("name", false)`,
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_msad_user_account_control_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_msad_lds_user_account_control_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_attribute_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_full_name_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_custom_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
	})
}
