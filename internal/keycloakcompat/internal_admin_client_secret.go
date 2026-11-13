/*
Copyright 2026 The Crossplane Authors.
*/

// Package keycloakcompat contains compatibility shims for Keycloak server
// behaviour changes that the pinned Terraform provider does not handle yet.
package keycloakcompat

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// emptyClientSecret replaces the 403 response body. It decodes into a zero
// keycloak.OpenidClientSecret, i.e. an empty client secret.
const emptyClientSecret = `{}`

// clientSecretPath matches the admin endpoint returning a client's secret and
// captures the (escaped) realm segment.
var clientSecretPath = regexp.MustCompile(`/admin/realms/([^/]+)/clients/[^/]+/client-secret$`)

// IsInternalAdminClient reports whether a client is one of Keycloak's internal
// admin clients: realm-management in every realm, or <realm>-realm in master.
// Since Keycloak 26.8 reading their secret requires manage-clients, which
// Keycloak never grants on these clients, so the request always returns 403.
func IsInternalAdminClient(realm, clientID string) bool {
	return clientID == "realm-management" || realm == "master" && strings.HasSuffix(clientID, "-realm")
}

// InstallInternalAdminClientSecretShim wraps the HTTP transport of the given
// Keycloak client so that a 403 on
// GET /admin/realms/{realm}/clients/{id}/client-secret is turned into an empty
// secret when {id} is an internal admin client (see IsInternalAdminClient).
//
// The pinned Terraform provider unconditionally fetches the client secret when
// reading or importing a keycloak_openid_client, so observing e.g. the
// realm-management client fails against Keycloak >= 26.8. This mirrors
// keycloak/terraform-provider-keycloak#1722 and becomes a no-op once the
// upstream provider skips the secret fetch for these clients; it can be
// removed after that fix is part of the pinned provider version.
func InstallInternalAdminClientSecretShim(kc *keycloak.KeycloakClient) {
	if kc == nil {
		return
	}
	httpClient := kc.GetHttpClient()
	if httpClient == nil {
		return
	}
	if _, ok := httpClient.Transport.(*internalAdminClientSecretTransport); ok {
		return
	}
	next := httpClient.Transport
	if next == nil {
		next = http.DefaultTransport
	}
	httpClient.Transport = &internalAdminClientSecretTransport{next: next}
}

type internalAdminClientSecretTransport struct {
	next http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (t *internalAdminClientSecretTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.next.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusForbidden || req.Method != http.MethodGet {
		return resp, err
	}
	m := clientSecretPath.FindStringSubmatch(req.URL.EscapedPath())
	if m == nil {
		return resp, nil
	}
	if !t.isInternalAdminClient(req, m[1]) {
		return resp, nil
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return &http.Response{
		Status:        "200 OK",
		StatusCode:    http.StatusOK,
		Proto:         resp.Proto,
		ProtoMajor:    resp.ProtoMajor,
		ProtoMinor:    resp.ProtoMinor,
		Header:        http.Header{"Content-Type": []string{"application/json"}},
		Body:          io.NopCloser(bytes.NewReader([]byte(emptyClientSecret))),
		ContentLength: int64(len(emptyClientSecret)),
		Request:       req,
	}, nil
}

// isInternalAdminClient fetches the client representation the secret request
// refers to, using the same headers (credentials), and checks its clientId.
func (t *internalAdminClientSecretTransport) isInternalAdminClient(req *http.Request, escapedRealm string) bool {
	realm, err := url.PathUnescape(escapedRealm)
	if err != nil {
		return false
	}

	clientURL := *req.URL
	clientURL.Path = strings.TrimSuffix(clientURL.Path, "/client-secret")
	clientURL.RawPath = strings.TrimSuffix(clientURL.RawPath, "/client-secret")

	lookup := req.Clone(req.Context())
	lookup.URL = &clientURL
	lookup.RequestURI = ""

	resp, err := t.next.RoundTrip(lookup)
	if err != nil {
		return false
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return false
	}

	var client struct {
		ClientID string `json:"clientId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&client); err != nil {
		return false
	}
	return IsInternalAdminClient(realm, client.ClientID)
}
