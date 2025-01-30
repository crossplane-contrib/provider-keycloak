package role

import (
	"context"
	"errors"
	role "github.com/crossplane-contrib/provider-keycloak/apis/role/v1alpha1"
	"github.com/crossplane-contrib/provider-keycloak/config/utils"
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

	r := role.RoleParameters{}
	err = utils.UnmarshalTerraformParamsToObject(parameters, &r)
	if err != nil {
		return "", err
	}

	if r.RealmID == nil {
		return "", errors.New("realmId not set")
	}

	if r.Name == nil {
		return "", errors.New("name not set")
	}

	clientId := ""
	if r.ClientID != nil {
		clientId = *r.ClientID
	}

	if externalName != "" {
		found, err := kcClient.GetRole(ctx, *r.RealmID, externalName)
		if err != nil {
			var apiErr *keycloak.ApiError
			if !(errors.As(err, &apiErr) && apiErr.Code == 404) {
				return "", err
			}
		} else {
			return found.Id, nil
		}
	}

	found, err := kcClient.GetRoleByName(ctx, *r.RealmID, clientId, *r.Name)
	if err != nil {
		var apiErr *keycloak.ApiError
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return "", nil
		}

		return "", err
	}

	return found.Id, nil
}
