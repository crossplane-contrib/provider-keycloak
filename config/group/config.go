package group

import "github.com/crossplane/upjet/pkg/config"

// Group is the short group name for the resources in this package
var Group = "group"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_group", func(r *config.Resource) {
		r.ShortGroup = Group

		r.References["parent_id"] = config.Reference{
			Type: "Group",
		}
	})
	p.AddResourceConfigurator("keycloak_group_memberships", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["group_id"] = config.Reference{
			Type: "Group",
		}

	})
	p.AddResourceConfigurator("keycloak_group_roles", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["group_id"] = config.Reference{
			Type: "Group",
		}
	})
	p.AddResourceConfigurator("keycloak_group_permissions", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["group_id"] = config.Reference{
			Type: "Group",
		}
	})
}
