package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const internalAdminClientUUID = "e325dcb7-5459-4fa4-87a5-c34adbe529dc"

func TestGetOpenidClientWithForbiddenSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/clients/"+internalAdminClientUUID+"/client-secret"):
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"HTTP 403 Forbidden"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/clients/"+internalAdminClientUUID):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": internalAdminClientUUID, "clientId": "realm-management", "bearerOnly": true,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	kc, err := keycloak.NewKeycloakClient(
		context.Background(),
		srv.URL, "", "", "admin-cli", "", "master",
		"", "", "token", "", "", "", "",
		true, 5, "", false, "", "", "test", false, nil, "",
	)
	if err != nil {
		t.Fatalf("NewKeycloakClient: %v", err)
	}

	client, err := kc.GetOpenidClient(context.Background(), "test", internalAdminClientUUID)
	if err != nil {
		t.Fatalf("GetOpenidClient: %v", err)
	}
	if !client.ClientSecretUnreadable || client.ClientSecret != "" {
		t.Errorf("unexpected client secret state: unreadable=%v secret=%q", client.ClientSecretUnreadable, client.ClientSecret)
	}
}
