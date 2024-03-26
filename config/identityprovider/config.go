package identityprovider

import "github.com/crossplane/upjet/pkg/config"

const (
	// Group is the short group for this provider.
	Group = "identityprovider"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_custom_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			Type: "github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm",
		}
	})
}
