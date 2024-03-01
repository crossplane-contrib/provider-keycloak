package group

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_group", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "group"

		r.References["parent_id"] = config.Reference{
			Type: "Group",
		}
	})
	p.AddResourceConfigurator("keycloak_group_memberships", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "group"
		r.References["group_id"] = config.Reference{
			Type: "Group",
		}

	})
	p.AddResourceConfigurator("keycloak_group_roles", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "group"
		r.References["group_id"] = config.Reference{
			Type: "Group",
		}
		r.References["role_ids"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/role/v1alpha1.Role",
		}
	})
	p.AddResourceConfigurator("keycloak_group_permissions", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "group"
		r.References["group_id"] = config.Reference{
			Type: "Group",
		}
	})
}
