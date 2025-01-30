package role

import (
	"bytes"
	"context"
	"errors"
	"github.com/crossplane-contrib/provider-keycloak/internal/clients"
	"github.com/crossplane/upjet/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strings"
	"text/template"
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

// IdentifierByNameLookup is used to find the existing resource by it´s identifying properties
var IdentifierByNameLookup = config.ExternalName{
	SetIdentifierArgumentFn: config.NopSetIdentifierArgument,
	GetExternalNameFn:       getExternalNameFromRole,
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

	role, err := kcClient.GetRoleByName(ctx, realmID.(string), clientID.(string), name.(string))
	if err != nil {
		var apiErr *keycloak.ApiError
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return "", nil
		}

		return "", err
	}

	return role.Id, nil
}

func getExternalNameFromRole(tfState map[string]any) (string, error) {
	t, err := template.New("getExternalName").Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
		"ToUpper": strings.ToUpper,
	}).Parse(`{{if eq .client_id ""}}{{ .realm_id }}/{{ .name }}{{else}}{{ .realm_id }}/{{ .client_id }}/{{ .name }}{{end}}`)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, tfState)
	if err != nil {
		return "", err
	}
	externalName := buf.String()
	return externalName, nil
}
