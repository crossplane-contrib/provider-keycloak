package role

import (
	"context"
	"errors"
	"github.com/crossplane-contrib/provider-keycloak/internal/clients"
	"github.com/crossplane/upjet/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_role", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = "role"
		r.References["composite_roles"] = config.Reference{
			TerraformName: "keycloak_role",
		}
	})
}

// IdentifierLookupForRole is used to find the existing resource by it´s identifying properties
var IdentifierLookupForRole = config.ExternalName{
	SetIdentifierArgumentFn: config.NopSetIdentifierArgument,
	GetExternalNameFn:       config.IDAsExternalName,
	GetIDFn:                 getIdFromRole,
	DisableNameInitializer:  true,
}

func getIdFromRole(ctx context.Context, externalName string, parameters map[string]any, terraformProviderConfig map[string]any) (string, error) {

	kcClient, err := clients.NewKeycloakClient(ctx, terraformProviderConfig)
	if err != nil {
		return "", err
	}

	realmID, realmIdExists := parameters["realm_id"]
	if !realmIdExists {
		return "", errors.New("realmId not set")
	}

	name, nameExists := parameters["name"]
	if !nameExists {
		return "", errors.New("name not set")
	}

	clientID, clientIdExists := parameters["client_id"]
	if !clientIdExists {
		clientID = ""
	}

	if externalName != "" {
		found, err := kcClient.GetRole(ctx, realmID.(string), externalName)
		if err != nil {
			var apiErr *keycloak.ApiError
			if !(errors.As(err, &apiErr) && apiErr.Code == 404) {
				return "", err
			}
		} else {
			return found.Id, nil
		}
	}

	found, err := kcClient.GetRoleByName(ctx, realmID.(string), clientID.(string), name.(string))
	if err != nil {
		var apiErr *keycloak.ApiError
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return "", nil
		}

		return "", err
	}

	return found.Id, nil
}
