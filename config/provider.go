/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	"github.com/crossplane/upjet/pkg/config"
	ujconfig "github.com/crossplane/upjet/pkg/config"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/defaults"
	"github.com/crossplane-contrib/provider-keycloak/config/group"
	"github.com/crossplane-contrib/provider-keycloak/config/mapper"
	"github.com/crossplane-contrib/provider-keycloak/config/oidc"
	"github.com/crossplane-contrib/provider-keycloak/config/openidclient"
	"github.com/crossplane-contrib/provider-keycloak/config/openidgroup"
	"github.com/crossplane-contrib/provider-keycloak/config/realm"
	"github.com/crossplane-contrib/provider-keycloak/config/role"
	"github.com/crossplane-contrib/provider-keycloak/config/saml"
	"github.com/crossplane-contrib/provider-keycloak/config/user"
)

const (
	resourcePrefix = "keycloak"
	modulePath     = "github.com/crossplane-contrib/provider-keycloak"
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
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
			KnownReferencers(),
		),
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
					Type: "github.com/crossplane-contrib/provider-keycloak/apis/realm/v1alpha1.Realm",
				}
			case "client_id":
				r.References["client_id"] = config.Reference{
					Type: "github.com/crossplane-contrib/provider-keycloak/apis/openidclient/v1alpha1.Client",
				}
			case "service_account_user_id":
				r.References["service_account_user_id"] = config.Reference{
					Type:      "github.com/crossplane-contrib/provider-keycloak/apis/openidclient/v1alpha1.Client",
					Extractor: common.PathServiceAccountRoleIDExtractor,
				}
				r.LateInitializer = config.LateInitializer{
					IgnoredFields: []string{"service_account_user_id"},
				}

			}
		}
	}
}
