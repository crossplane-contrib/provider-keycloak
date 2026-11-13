/*
Copyright 2026 The Crossplane Authors.
*/

package keycloakcompat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const testClientUUID = "e325dcb7-5459-4fa4-87a5-c34adbe529dc"

func TestIsInternalAdminClient(t *testing.T) {
	cases := map[string]struct {
		realm    string
		clientID string
		want     bool
	}{
		"RealmManagement":       {realm: "test", clientID: "realm-management", want: true},
		"MasterRealmManagement": {realm: "master", clientID: "realm-management", want: true},
		"MasterRealmClient":     {realm: "master", clientID: "test-realm", want: true},
		"NonMasterRealmSuffix":  {realm: "test", clientID: "my-realm", want: false},
		"RegularClient":         {realm: "test", clientID: "my-app", want: false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsInternalAdminClient(tc.realm, tc.clientID); got != tc.want {
				t.Errorf("IsInternalAdminClient(%q, %q) = %v, want %v", tc.realm, tc.clientID, got, tc.want)
			}
		})
	}
}

// newKeycloak26_8Server emulates Keycloak 26.8: reading the client secret is
// forbidden for every client, regardless of the caller's roles.
func newKeycloak26_8Server(t *testing.T, clientID string, secretCalls *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/clients/"+testClientUUID+"/client-secret"):
			secretCalls.Add(1)
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"HTTP 403 Forbidden"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/clients/"+testClientUUID):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": testClientUUID, "clientId": clientID, "bearerOnly": true})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newTestClient(t *testing.T, url string) *keycloak.KeycloakClient {
	t.Helper()
	kc, err := keycloak.NewKeycloakClient(
		context.Background(),
		url, "", "", "admin-cli", "", "master",
		"", "", "token", "", "", "", "",
		true, 5, "", false, "", "", "test", false, nil, "",
	)
	if err != nil {
		t.Fatalf("NewKeycloakClient: %v", err)
	}
	return kc
}

func TestShimReturnsEmptySecretForInternalAdminClient(t *testing.T) {
	var secretCalls atomic.Int32
	srv := newKeycloak26_8Server(t, "realm-management", &secretCalls)
	defer srv.Close()

	kc := newTestClient(t, srv.URL)
	InstallInternalAdminClientSecretShim(kc)
	// Installing twice must not stack transports.
	InstallInternalAdminClientSecretShim(kc)

	client, err := kc.GetOpenidClient(context.Background(), "test", testClientUUID)
	if err != nil {
		t.Fatalf("GetOpenidClient: %v", err)
	}
	if client.ClientId != "realm-management" || client.ClientSecret != "" {
		t.Errorf("unexpected client: clientId=%q secret=%q", client.ClientId, client.ClientSecret)
	}
	if got := secretCalls.Load(); got != 1 {
		t.Errorf("client-secret requests = %d, want 1", got)
	}
}

func TestShimKeeps403ForRegularClient(t *testing.T) {
	var secretCalls atomic.Int32
	srv := newKeycloak26_8Server(t, "my-app", &secretCalls)
	defer srv.Close()

	kc := newTestClient(t, srv.URL)
	InstallInternalAdminClientSecretShim(kc)

	_, err := kc.GetOpenidClient(context.Background(), "test", testClientUUID)
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("GetOpenidClient error = %v, want 403", err)
	}
}

func TestWithoutShimInternalAdminClientFails(t *testing.T) {
	var secretCalls atomic.Int32
	srv := newKeycloak26_8Server(t, "realm-management", &secretCalls)
	defer srv.Close()

	kc := newTestClient(t, srv.URL)
	_, err := kc.GetOpenidClient(context.Background(), "test", testClientUUID)
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("GetOpenidClient error = %v, want 403", err)
	}
}

func TestShimIgnoresNilClient(t *testing.T) {
	InstallInternalAdminClientSecretShim(nil)
}
