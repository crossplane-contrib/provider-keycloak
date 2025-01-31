package clients

import (
	"context"
	"github.com/crossplane/upjet/pkg/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// NewKeycloakClient creates a new keycloak client based on the settings in the provider configuration
func NewKeycloakClient(ctx context.Context, terraformProviderConfig map[string]any) (*keycloak.KeycloakClient, error) {
	config := terraformProviderConfig["configuration"].(terraform.ProviderConfiguration)

	url := tryGetString(config, "url", "")
	basePath := tryGetString(config, "base_path", "")
	clientID := tryGetString(config, "client_id", "")
	clientSecret := tryGetString(config, "client_secret", "")
	username := tryGetString(config, "username", "")
	password := tryGetString(config, "password", "")
	realm := tryGetString(config, "realm", "master")
	initialLogin := tryGetBool(config, "initial_login", true)
	clientTimeout := tryGetInt(config, "client_timeout", 15)
	tlsInsecureSkipVerify := tryGetBool(config, "tls_insecure_skip_verify", false)
	rootCaCertificate := tryGetString(config, "root_ca_certificate", "")
	redHatSSO := tryGetBool(config, "initial_login", false)
	additionalHeaders := tryGetMap(config, "additional_headers")
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
