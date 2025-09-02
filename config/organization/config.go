package organization

import (
	"context"
	"strings"

	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_organization", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		r.ShortGroup = "organization"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
	})
}

var organizationIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm", "name", "domain"},
	GetIDByExternalName:          getOrganizationIDByExternalName,
	GetIDByIdentifyingProperties: getOrganizationIDByIdentifyingProperties,
}

// OrganizationIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var OrganizationIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(organizationIdentifyingPropertiesLookup)

func getOrganizationIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOrganization(ctx, parameters["realm"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getOrganizationIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetOrganizationByName(ctx, parameters["realm"].(string), parameters["name"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "organization with name") {
			return "", nil
		}

		return "", err
	}

	return found.Id, nil
}
