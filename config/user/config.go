package user

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_user", func(r *config.Resource) {
		r.ShortGroup = "user"

		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"required_actions", "initial_password.value", "initial_password.value", "initial_password.temporary"},
		}

	})

	p.AddResourceConfigurator("keycloak_user_groups", func(r *config.Resource) {
		r.ShortGroup = "user"

		r.References["user_id"] = config.Reference{
			TerraformName: "keycloak_user",
		}

		r.References["group_ids"] = config.Reference{}
	})

	p.AddResourceConfigurator("keycloak_user_roles", func(r *config.Resource) {
		r.ShortGroup = "user"

		r.References["user_id"] = config.Reference{
			TerraformName: "keycloak_user",
		}
	})

	p.AddResourceConfigurator("keycloak_users_permissions", func(r *config.Resource) {
		r.ShortGroup = "user"
	})
}
