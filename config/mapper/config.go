package mapper

import (
	"context"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane/upjet/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_generic_protocol_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"
		r.References["client_scope_id"] = config.Reference{
			TerraformName: "keycloak_openid_client_scope",
		}
	})

	p.AddResourceConfigurator("keycloak_generic_role_mapper", func(r *config.Resource) {
		r.ShortGroup = "client"
		r.References["role_id"] = config.Reference{
			TerraformName: "keycloak_role",
		}
		r.References["client_scope_id"] = config.Reference{
			TerraformName: "keycloak_openid_client_scope",
		}
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

func getProtocolMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetGenericProtocolMapper(ctx, parameters["realm_id"].(string), parameters["client_id"].(string), parameters["client_scope_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getProtocolMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
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
