package workflow

import "github.com/crossplane/upjet/v2/pkg/config"

const (
	// Group is the short group for this provider.
	Group = "workflow"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_workflow", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
	})
}
