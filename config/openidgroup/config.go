package openidgroup

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
)

const (
	// Group is the short group for this provider.
	Group = "openidgroup"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_openid_group_membership_protocol_mapper", func(r *config.Resource) {
		configureOpenIDProtocolMapper(r)
	})

	for _, name := range []string{
		"keycloak_openid_audience_protocol_mapper",
		"keycloak_openid_audience_resolve_protocol_mapper",
		"keycloak_openid_full_name_protocol_mapper",
		"keycloak_openid_hardcoded_claim_protocol_mapper",
		"keycloak_openid_hardcoded_role_protocol_mapper",
		"keycloak_openid_sub_protocol_mapper",
		"keycloak_openid_user_attribute_protocol_mapper",
		"keycloak_openid_user_client_role_protocol_mapper",
		"keycloak_openid_user_property_protocol_mapper",
		"keycloak_openid_user_realm_role_protocol_mapper",
		"keycloak_openid_user_session_note_protocol_mapper",
	} {
		resourceName := name
		p.AddResourceConfigurator(resourceName, func(r *config.Resource) {
			configureOpenIDProtocolMapper(r)
			if resourceName == "keycloak_openid_hardcoded_role_protocol_mapper" {
				r.References["role_id"] = config.Reference{
					TerraformName: "keycloak_role",
					Extractor:     common.PathUUIDExtractor,
				}
			}
			if resourceName == "keycloak_openid_user_client_role_protocol_mapper" {
				r.References["client_id_for_role_mappings"] = config.Reference{
					TerraformName: "keycloak_openid_client",
					Extractor:     common.PathUUIDExtractor,
				}
			}
		})
	}
}

func configureOpenIDProtocolMapper(r *config.Resource) {
	r.ShortGroup = Group
	r.References["client_id"] = config.Reference{
		TerraformName: "keycloak_openid_client",
		Extractor:     common.PathUUIDExtractor,
	}
	r.References["client_scope_id"] = config.Reference{
		TerraformName: "keycloak_openid_client_scope",
	}
}

var identifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
	OptionalParameters:           []string{"client_id", "client_scope_id"},
	GetIDByExternalName:          getIDByExternalName,
	GetIDByIdentifyingProperties: getIDByIdentifyingProperties,
}

// IdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var IdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(identifyingPropertiesLookup)

func getIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	clientID, clientScopeID, err := getProtocolMapperAttachmentIDs(parameters)
	if err != nil {
		return "", err
	}
	found, err := kcClient.GetOpenIdGroupMembershipProtocolMapper(ctx, parameters["realm_id"].(string), clientID, clientScopeID, id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	clientID, clientScopeID, err := getProtocolMapperAttachmentIDs(parameters)
	if err != nil {
		return "", err
	}
	found, err := lookup.GetGenericProtocolMappers(kcClient, ctx, parameters["realm_id"].(string), clientID, clientScopeID)
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

var openidProtocolMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
	OptionalParameters:           []string{"client_id", "client_scope_id"},
	GetIDByExternalName:          getOpenidProtocolMapperIDByExternalName,
	GetIDByIdentifyingProperties: getOpenidProtocolMapperIDByIdentifyingProperties,
}

// OpenidProtocolMapperIdentifierFromIdentifyingProperties is used to find existing OpenID protocol mappers by their identifying properties.
var OpenidProtocolMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(openidProtocolMapperIdentifyingPropertiesLookup)

func getOpenidProtocolMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	clientID, clientScopeID, err := getProtocolMapperAttachmentIDs(parameters)
	if err != nil {
		return "", err
	}
	found, err := kcClient.GetGenericProtocolMapper(ctx, parameters["realm_id"].(string), clientID, clientScopeID, id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getOpenidProtocolMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	clientID, clientScopeID, err := getProtocolMapperAttachmentIDs(parameters)
	if err != nil {
		return "", err
	}
	found, err := lookup.GetGenericProtocolMappers(kcClient, ctx, parameters["realm_id"].(string), clientID, clientScopeID)
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

func getProtocolMapperAttachmentIDs(parameters map[string]any) (string, string, error) {
	clientID, _ := parameters["client_id"].(string)
	clientScopeID, _ := parameters["client_scope_id"].(string)
	if clientID == "" && clientScopeID == "" {
		return "", "", errors.New("either client_id or client_scope_id must be set")
	}
	return clientID, clientScopeID, nil
}
