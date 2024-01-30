/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	"github.com/crossplane/upjet/pkg/config"
	ujconfig "github.com/crossplane/upjet/pkg/config"

	"github.com/stakater/provider-keycloak/config/defaults"
	"github.com/stakater/provider-keycloak/config/group"
	"github.com/stakater/provider-keycloak/config/mapper"
	"github.com/stakater/provider-keycloak/config/oidc"
	"github.com/stakater/provider-keycloak/config/openidclient"
	"github.com/stakater/provider-keycloak/config/openidgroup"
	"github.com/stakater/provider-keycloak/config/realm"
	"github.com/stakater/provider-keycloak/config/role"
	"github.com/stakater/provider-keycloak/config/saml"
	"github.com/stakater/provider-keycloak/config/user"
)

const (
	resourcePrefix = "keycloak"
	modulePath     = "github.com/stakater/provider-keycloak"
	rootGroup      = "keycloak.crossplane.io"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
			KnownReferencers(),
		),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithRootGroup(rootGroup))

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
		realm.Configure,
		group.Configure,
		role.Configure,
		openidclient.Configure,
		openidgroup.Configure,
		mapper.Configure,
		user.Configure,
		defaults.Configure,
		oidc.Configure,
		saml.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}

// KnownReferencers adds referencers for fields that are known and common among
// more than a few resources.
func KnownReferencers() config.ResourceOption { //nolint:gocyclo
	return func(r *config.Resource) {
		for k, s := range r.TerraformResource.Schema {
			// We shouldn't add referencers for status fields and sensitive fields
			// since they already have secret referencer.
			if (s.Computed && !s.Optional) || s.Sensitive {
				continue
			}
			switch k {
			case "realm_id":
				r.References["realm_id"] = config.Reference{
					Type: "github.com/stakater/provider-keycloak/apis/realm/v1alpha1.Realm",
				}
			case "client_id":
				r.References["client_id"] = config.Reference{
					Type: "github.com/stakater/provider-keycloak/apis/openidclient/v1alpha1.Client",
				}
			}
		}
	}
}
