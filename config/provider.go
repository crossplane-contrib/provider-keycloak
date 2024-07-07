/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	"github.com/crossplane/upjet/pkg/config"
	ujconfig "github.com/crossplane/upjet/pkg/config"
	conversiontfjson "github.com/crossplane/upjet/pkg/types/conversion/tfjson"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	keycloakProvider "github.com/mrparkers/terraform-provider-keycloak/provider"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-keycloak/config/authentication"
	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/defaults"
	"github.com/crossplane-contrib/provider-keycloak/config/group"
	"github.com/crossplane-contrib/provider-keycloak/config/identityprovider"
	"github.com/crossplane-contrib/provider-keycloak/config/ldap"
	"github.com/crossplane-contrib/provider-keycloak/config/mapper"
	"github.com/crossplane-contrib/provider-keycloak/config/oidc"
	"github.com/crossplane-contrib/provider-keycloak/config/openidclient"
	"github.com/crossplane-contrib/provider-keycloak/config/openidgroup"
	"github.com/crossplane-contrib/provider-keycloak/config/realm"
	"github.com/crossplane-contrib/provider-keycloak/config/role"
	"github.com/crossplane-contrib/provider-keycloak/config/saml"
	"github.com/crossplane-contrib/provider-keycloak/config/samlclient"
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

// workaround for the TF Azure v3.57.0-based no-fork release: We would like to
// keep the types in the generated CRDs intact
// (prevent number->int type replacements).
func getProviderSchema(s string) (*schema.Provider, error) {
	ps := tfjson.ProviderSchemas{}
	if err := ps.UnmarshalJSON([]byte(s)); err != nil {
		panic(err)
	}
	if len(ps.Schemas) != 1 {
		return nil, errors.Errorf("there should exactly be 1 provider schema but there are %d", len(ps.Schemas))
	}
	var rs map[string]*tfjson.Schema
	for _, v := range ps.Schemas {
		rs = v.ResourceSchemas
		break
	}
	return &schema.Provider{
		ResourcesMap: conversiontfjson.GetV2ResourceMap(rs),
	}, nil
}

// GetProvider returns provider configuration
func GetProvider(generationProvider bool) (*ujconfig.Provider, error) {
	var p *schema.Provider
	var err error
	if generationProvider {
		p, err = getProviderSchema(providerSchema)
	} else {
		p = keycloakProvider.KeycloakProvider(nil)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get the Terraform provider schema with generation mode set to %t", generationProvider)
	}

	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithIncludeList([]string{}),
		ujconfig.WithTerraformPluginSDKIncludeList(ExternalNameConfigured()),
		ujconfig.WithTerraformPluginFrameworkIncludeList([]string{}), // For future resources
		ujconfig.WithTerraformProvider(p),
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
		identityprovider.Configure,
		ldap.Configure,
		samlclient.Configure,
		authentication.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc, nil
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
					TerraformName: "keycloak_realm",
				}
			case "client_id":
				r.References["client_id"] = config.Reference{
					TerraformName: "keycloak_openid_client",
					Extractor:     common.PathUUIDExtractor,
				}
			case "service_account_user_id":
				r.References["service_account_user_id"] = config.Reference{
					TerraformName:     "keycloak_openid_client",
					Extractor:         common.PathServiceAccountRoleIDExtractor,
					RefFieldName:      "ServiceAccountUserClientIDRef",
					SelectorFieldName: "ServiceAccountUserClientIDSelector",
				}
				r.LateInitializer = config.LateInitializer{
					IgnoredFields: []string{"service_account_user_id"},
				}

			case "role_ids":
				r.References["role_ids"] = config.Reference{
					TerraformName: "keycloak_role",
					Extractor:     common.PathUUIDExtractor,
				}

			case "role_id":
				r.References["role_id"] = config.Reference{
					TerraformName: "keycloak_role",
					Extractor:     common.PathUUIDExtractor,
				}
			}

		}
	}
}
