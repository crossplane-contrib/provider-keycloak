package workflow

import (
	"context"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
)

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

var identifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm", "name"},
	GetIDByExternalName:          getIDByExternalName,
	GetIDByIdentifyingProperties: getIDByIdentifyingProperties,
}

// IdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var IdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(identifyingPropertiesLookup)

func getIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetWorkflow(ctx, parameters["realm"].(string), id)
	if err != nil {
		if isWorkflowNotFoundError(err) {
			return "", &keycloak.ApiError{Code: 404, Message: err.Error()}
		}
		return "", err
	}
	return found.Id, nil
}

func getIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetWorkflowByName(ctx, parameters["realm"].(string), parameters["name"].(string))
	if err != nil {
		if isWorkflowNotFoundError(err) {
			return "", &keycloak.ApiError{Code: 404, Message: err.Error()}
		}
		return "", err
	}
	return found.Id, nil
}

func isWorkflowNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "workflow with") && strings.Contains(err.Error(), "not found in realm")
}
