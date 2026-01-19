package samlclient

import (
	"context"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
)

const (
	// Group is the short group for this provider.
	Group = "samlclient"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_saml_client", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group

		if s, ok := r.TerraformResource.Schema["encryption_certificate"]; ok {
			s.Sensitive = true
		}
		if s, ok := r.TerraformResource.Schema["signing_certificate"]; ok {
			s.Sensitive = true
		}
		if s, ok := r.TerraformResource.Schema["signing_private_key"]; ok {
			s.Sensitive = true
		}

		// Avoid removing BrowserIdRef
		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"authentication_flow_binding_overrides"},
		}
	})

	p.AddResourceConfigurator("keycloak_saml_client_default_scopes", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group

		r.References["client_id"] = config.Reference{
			TerraformName: "keycloak_saml_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_saml_client_scope", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = Group
	})

}

var clientIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "client_id"},
	GetIDByExternalName:          getClientIDByExternalName,
	GetIDByIdentifyingProperties: getClientIDByIdentifyingProperties,
}

// ClientIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var ClientIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(clientIdentifyingPropertiesLookup)

func getClientIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetGenericClient(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getClientIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetGenericClientByClientId(ctx, parameters["realm_id"].(string), parameters["client_id"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return "", nil
		}

		return "", err
	}
	return found.Id, nil
}

var clientScopeIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
	GetIDByExternalName:          getClientScopeIDByExternalName,
	GetIDByIdentifyingProperties: getClientScopeIDByIdentifyingProperties,
}

// ClientScopeIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var ClientScopeIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(clientScopeIdentifyingPropertiesLookup)

func getClientScopeIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetSamlClientScope(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getClientScopeIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.ListSamlClientScopesWithFilter(ctx, parameters["realm_id"].(string), func(scope *keycloak.SamlClientScope) bool {
		return scope.Name == parameters["name"].(string)
	})

	if err != nil {
		return "", err
	}

	return lookup.SingleOrEmpty(found, func(scope *keycloak.SamlClientScope) string {
		return scope.Id
	})
}
