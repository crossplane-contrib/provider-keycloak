package mapper

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane-contrib/provider-keycloak/config/multitypes"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_generic_protocol_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"

		// Skip late-initialization for the config map so that
		// server-side defaults (e.g. "introspection.token.claim",
		// "multivalued") added by Keycloak are not copied back into
		// spec.forProvider.config, which would cause perpetual drift.
		// See https://github.com/crossplane-contrib/provider-keycloak/issues/558
		if r.LateInitializer.IgnoredFields == nil {
			r.LateInitializer.IgnoredFields = []string{}
		}
		r.LateInitializer.IgnoredFields = append(r.LateInitializer.IgnoredFields, "config")

		multitypes.ApplyToWithOptions(r, "client_id",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_client_id",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "client_id",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client",
					Extractor:     common.PathUUIDExtractor,
				},
			})

		multitypes.ApplyToWithOptions(r, "client_scope_id",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_client_scope_id",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client_scope",
				},
			},
			multitypes.Instance{
				Name: "client_scope_id",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_scope",
				},
			})
	})

	p.AddResourceConfigurator("keycloak_generic_client_protocol_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"
		r.Kind = "GenericClientProtocolMapper"

		// Skip late-initialization for the config map so that
		// server-side defaults (e.g. "introspection.token.claim",
		// "multivalued") added by Keycloak are not copied back into
		// spec.forProvider.config, which would cause perpetual drift.
		// See https://github.com/crossplane-contrib/provider-keycloak/issues/558
		if r.LateInitializer.IgnoredFields == nil {
			r.LateInitializer.IgnoredFields = []string{}
		}
		r.LateInitializer.IgnoredFields = append(r.LateInitializer.IgnoredFields, "config")

		multitypes.ApplyToWithOptions(r, "client_id",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_client_id",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "client_id",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client",
					Extractor:     common.PathUUIDExtractor,
				},
			})

		multitypes.ApplyToWithOptions(r, "client_scope_id",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_client_scope_id",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client_scope",
				},
			},
			multitypes.Instance{
				Name: "client_scope_id",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_scope",
				},
			})
	})

	p.AddResourceConfigurator("keycloak_generic_role_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"
		r.References["role_id"] = config.Reference{
			TerraformName: "keycloak_role",
		}

		multitypes.ApplyToWithOptions(r, "client_id",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_client_id",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "client_id",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client",
					Extractor:     common.PathUUIDExtractor,
				},
			})

		multitypes.ApplyToWithOptions(r, "client_scope_id",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_client_scope_id",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client_scope",
				},
			},
			multitypes.Instance{
				Name: "client_scope_id",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_scope",
				},
			})
	})
	p.AddResourceConfigurator("keycloak_generic_client_role_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"
		r.Kind = "GenericClientRoleMapper"
		r.References["role_id"] = config.Reference{
			TerraformName: "keycloak_role",
		}

		multitypes.ApplyToWithOptions(r, "client_id",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_client_id",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client",
					Extractor:     common.PathUUIDExtractor,
				},
			},
			multitypes.Instance{
				Name: "client_id",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client",
					Extractor:     common.PathUUIDExtractor,
				},
			})

		multitypes.ApplyToWithOptions(r, "client_scope_id",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				Name: "saml_client_scope_id",
				Reference: config.Reference{
					TerraformName: "keycloak_saml_client_scope",
				},
			},
			multitypes.Instance{
				Name: "client_scope_id",
				Reference: config.Reference{
					TerraformName: "keycloak_openid_client_scope",
				},
			})
	})
}

var protocolMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
	OptionalParameters:           []string{"client_id", "client_scope_id"},
	GetIDByExternalName:          getProtocolMapperIDByExternalName,
	GetIDByIdentifyingProperties: getProtocolMapperIDByIdentifyingProperties,
}

// ProtocolMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var ProtocolMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(protocolMapperIdentifyingPropertiesLookup)

// GenericClientProtocolMapperIdentifierFromIdentifyingProperties reuses the
// generic protocol mapper lookup for the client-scoped variant.
var GenericClientProtocolMapperIdentifierFromIdentifyingProperties = ProtocolMapperIdentifierFromIdentifyingProperties

func getProtocolMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	if parameters["client_id"].(string) == "" && parameters["client_scope_id"].(string) == "" {
		return "", errors.New("Either client_id or client_scope_id must be set")
	}

	found, err := kcClient.GetGenericProtocolMapper(ctx, parameters["realm_id"].(string), parameters["client_id"].(string), parameters["client_scope_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getProtocolMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	if parameters["client_id"].(string) == "" && parameters["client_scope_id"].(string) == "" {
		return "", errors.New("Either client_id or client_scope_id must be set")
	}

	found, err := lookup.GetGenericProtocolMappers(kcClient, ctx, parameters["realm_id"].(string), parameters["client_id"].(string), parameters["client_scope_id"].(string))
	if err != nil {
		return "", err
	}

	filtered := lookup.Filter(found.ProtocolMappers, func(mapper *keycloak.GenericProtocolMapper) bool {
		return mapper.Name == parameters["name"].(string)
	})

	return lookup.SingleOrEmpty(filtered, func(mapper *keycloak.GenericProtocolMapper) string {
		return mapper.Id
	})
}
