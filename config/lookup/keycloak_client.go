package lookup

import (
	"context"
	"errors"
	"fmt"
	"github.com/crossplane/upjet/pkg/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Component is a generic keycloak data model
// This needs to be removed in the future. See comments on GetComponents method
type Component struct {
	Id           string              `json:"id,omitempty"`
	Name         string              `json:"name"`
	ProviderId   string              `json:"providerId"`
	ProviderType string              `json:"providerType"`
	ParentId     string              `json:"parentId"`
	Config       map[string][]string `json:"config"`
}

// newKeycloakClient creates a new keycloak client based on the settings in the provider configuration
// (This can be removed once this issue is resolved: https://github.com/crossplane/upjet/issues/464)
func newKeycloakClient(ctx context.Context, terraformProviderConfig map[string]any) (*keycloak.KeycloakClient, error) {
	c := terraformProviderConfig["configuration"].(terraform.ProviderConfiguration)

	url := tryGetString(c, "url", "")
	basePath := tryGetString(c, "base_path", "")
	clientID := tryGetString(c, "client_id", "")
	clientSecret := tryGetString(c, "client_secret", "")
	username := tryGetString(c, "username", "")
	password := tryGetString(c, "password", "")
	realm := tryGetString(c, "realm", "master")
	initialLogin := tryGetBool(c, "initial_login", true)
	clientTimeout := tryGetInt(c, "client_timeout", 15)
	tlsInsecureSkipVerify := tryGetBool(c, "tls_insecure_skip_verify", false)
	rootCaCertificate := tryGetString(c, "root_ca_certificate", "")
	redHatSSO := tryGetBool(c, "initial_login", false)
	additionalHeaders := tryGetMap(c, "additional_headers")
	userAgent := "Crossplane Keycloak Provider"

	keycloakClient, err := keycloak.NewKeycloakClient(ctx, url, basePath, clientID, clientSecret, realm, username, password, initialLogin, clientTimeout, rootCaCertificate, tlsInsecureSkipVerify, userAgent, redHatSSO, additionalHeaders)
	if err != nil {
		return nil, err
	}
	return keycloakClient, nil
}

func tryGetString(m map[string]any, key string, defaultValue string) string {
	value, ok := m[key]
	if ok {
		return value.(string)
	}
	return defaultValue
}

func tryGetBool(m map[string]any, key string, defaultValue bool) bool {
	value, ok := m[key]
	if ok {
		return value.(bool)
	}
	return defaultValue
}

func tryGetInt(m map[string]any, key string, defaultValue int) int {
	value, ok := m[key]
	if ok {
		return value.(int)
	}
	return defaultValue
}

func tryGetMap(m map[string]any, key string) map[string]string {
	value, ok := m[key]
	result := make(map[string]string)
	if ok {
		for k, v := range value.(map[string]interface{}) {
			result[k] = v.(string)
		}
	}
	return result
}

// GetComponents returns the components of the specified realm, type and name
// This needs to be removed in the future.
// We need to clarify with terraform-provider-keycloak maintainers if we could add a GetComponents method
// Currently we need this i.e. because there is no method to list all RealmKeystoreRsa
// or to get the RealmKeystoreRsa by name
func GetComponents(kcClient *keycloak.KeycloakClient, ctx context.Context, realmId string, typ string, name string) ([]*Component, error) {
	params := make(map[string]string)
	params["type"] = typ
	params["name"] = name

	var components []*Component

	err := keycloakClientGet(kcClient, ctx, fmt.Sprintf("/realms/%s/components", realmId), &components, params)
	if err != nil {
		return nil, err
	}

	return components, nil
}

type GenericProtocolMappers struct {
	ProtocolMappers []*keycloak.GenericProtocolMapper
}

// GetGenericProtocolMappers returns the protocol mappers of the specified realm, clientId or clientScopeId
// We need to clarify with terraform-provider-keycloak maintainers if we could add a GetGenericProtocolMappers method
func GetGenericProtocolMappers(kcClient *keycloak.KeycloakClient, ctx context.Context, realmId string, clientId string, clientScopeId string) (*GenericProtocolMappers, error) {
	var genericProtocolMappers GenericProtocolMappers
	var typ string
	var id string

	if clientId == "" && clientScopeId == "" {
		return nil, errors.New("either clientId or clientScopeId must be present, but both are empty")
	}

	if clientId != "" && clientScopeId != "" {
		return nil, errors.New("either clientId or clientScopeId must be present, but both are not empty")
	}

	if clientId != "" {
		typ = "clients"
		id = clientId
	}

	if clientScopeId != "" {
		typ = "client-scopes"
		id = clientScopeId
	}

	err := keycloakClientGet(kcClient, ctx, fmt.Sprintf("/realms/%s/%s/%s", realmId, typ, id), &genericProtocolMappers, nil)
	if err != nil {
		return nil, err
	}

	for _, protocolMapper := range genericProtocolMappers.ProtocolMappers {
		protocolMapper.RealmId = realmId
		if clientId != "" {
			protocolMapper.ClientId = clientId
		}
		if clientScopeId != "" {
			protocolMapper.ClientId = clientScopeId
		}
	}

	return &genericProtocolMappers, nil

}
