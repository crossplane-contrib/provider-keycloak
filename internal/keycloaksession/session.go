/*
Copyright 2024 Upbound Inc.
*/

package keycloaksession

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
// key-value configuration map. Keys are sorted before hashing and
// values are JSON-encoded to ensure deterministic output even when
// values are maps (whose iteration order is randomized in Go).
func ConfigCacheKey(config map[string]any) string {
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		v := config[k]
		// Use JSON marshaling for composite types to guarantee stable
		// output regardless of map iteration order.
		var valStr string
		switch v.(type) {
		case map[string]any, map[string]string, []any:
			if b, err := json.Marshal(v); err == nil {
				valStr = string(b)
			} else {
				valStr = fmt.Sprintf("%v", v)
			}
		default:
			valStr = fmt.Sprintf("%v", v)
		}
		_, _ = fmt.Fprintf(h, "%s=%s\n", k, valStr)
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

// logoutConfigKeys lists the only configuration keys cached for logout.
// Notably, "password" is NOT retained—only a pre-computed boolean
// "is_password_grant" signals that the session needs explicit logout.
var logoutConfigKeys = []string{
	"url",
	"base_path",
	"realm",
	"client_id",
	"client_secret",
	"is_password_grant",
}

// LogoutConfig returns a minimal copy of config containing only the
// fields required to perform a session logout. The actual password is
// not retained; instead, a boolean "is_password_grant" is computed and
// stored to signal whether the session requires explicit logout.
func LogoutConfig(config map[string]any) map[string]any {
	m := make(map[string]any, len(logoutConfigKeys))
	for _, k := range logoutConfigKeys {
		if v, ok := config[k]; ok {
			m[k] = v
		}
	}
	// Compute and store the grant-type boolean so that password is
	// never retained in the cache.
	m["is_password_grant"] = IsPasswordGrant(config)
	return m
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
	// Check the pre-computed boolean first (from LogoutConfig), fall
	// back to computing from username/password for raw configs.
	if isPwGrant, ok := config["is_password_grant"].(bool); ok {
		if !isPwGrant {
			return
		}
	} else if !IsPasswordGrant(config) {
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

	// Normalize by trimming whitespace, trimming trailing slash from
	// URL, and ensuring base_path has a leading slash (if non-empty)
	// to avoid double-slash.
	urlStr = strings.TrimSpace(urlStr)
	basePath = strings.TrimSpace(basePath)
	urlStr = strings.TrimRight(urlStr, "/")
	if basePath != "" && !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	basePath = strings.TrimRight(basePath, "/")

	logoutURL := fmt.Sprintf(logoutURLTemplate, urlStr+basePath, url.PathEscape(realm))

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
	_ = resp.Body.Close()
}
