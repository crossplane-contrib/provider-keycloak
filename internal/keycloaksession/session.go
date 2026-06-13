/*
Copyright 2024 Upbound Inc.
*/

package keycloaksession

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
	"unsafe"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// logoutURLTemplate is the OIDC end-session endpoint pattern.
// Parameters: base URL (including optional base path), realm name.
const logoutURLTemplate = "%s/realms/%s/protocol/openid-connect/logout"

// logoutTimeout caps how long a single logout HTTP call may take.
const logoutTimeout = 10 * time.Second

// ConfigCacheKey returns a stable SHA-256 hex digest for the given
// key-value configuration map. Keys are sorted before hashing so
// that two maps with identical contents always produce the same key.
func ConfigCacheKey(config map[string]any) string {
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%v\n", k, config[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// IsPasswordGrant returns true when the configuration uses the
// resource-owner password grant, which creates a server-side Keycloak
// session that should be explicitly logged out.
func IsPasswordGrant(config map[string]any) bool {
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)
	return username != "" && password != ""
}

// ExtractRefreshToken reads the current refresh token from a
// KeycloakClient instance. The upstream library keeps the token in
// the unexported field clientCredentials (*ClientCredentials).
// Because ClientCredentials is an exported type with an exported
// RefreshToken field, we only need unsafe to cross the unexported
// pointer boundary. If the struct layout changes in a future version
// the function silently returns "".
func ExtractRefreshToken(kcClient *keycloak.KeycloakClient) (token string) {
	if kcClient == nil {
		return ""
	}
	defer func() {
		if r := recover(); r != nil {
			token = ""
		}
	}()

	v := reflect.ValueOf(kcClient).Elem()
	credsField := v.FieldByName("clientCredentials")
	if !credsField.IsValid() || credsField.IsNil() {
		return ""
	}
	// credsField is an unexported *ClientCredentials pointer.
	// ClientCredentials and its RefreshToken field are exported.
	creds := *(**keycloak.ClientCredentials)(unsafe.Pointer(credsField.UnsafeAddr())) //nolint:gosec // required to access unexported pointer field for session cleanup
	if creds == nil {
		return ""
	}
	return creds.RefreshToken
}

// LogoutSession ends the Keycloak session associated with the given
// client by posting to the OIDC logout endpoint with the client's
// current refresh token. It is a best-effort operation: errors are
// silently ignored. It is a no-op for non-password-grant configs or
// when the refresh token cannot be obtained.
func LogoutSession(ctx context.Context, config map[string]any, kcClient *keycloak.KeycloakClient) {
	if !IsPasswordGrant(config) {
		return
	}

	refreshToken := ExtractRefreshToken(kcClient)
	if refreshToken == "" {
		return
	}

	urlStr, _ := config["url"].(string)
	basePath, _ := config["base_path"].(string)
	realm, _ := config["realm"].(string)
	if realm == "" {
		realm = "master"
	}
	clientID, _ := config["client_id"].(string)
	clientSecret, _ := config["client_secret"].(string)

	logoutURL := fmt.Sprintf(logoutURLTemplate, urlStr+basePath, realm)

	data := url.Values{
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
	}
	if clientSecret != "" {
		data.Set("client_secret", clientSecret)
	}

	logoutCtx, cancel := context.WithTimeout(ctx, logoutTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(logoutCtx, http.MethodPost, logoutURL, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
