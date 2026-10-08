/*
Copyright 2024 Upbound Inc.
*/

package keycloaksession

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func testJWT(claims string) string {
	enc := base64.RawURLEncoding
	return enc.EncodeToString([]byte(`{"alg":"none"}`)) + "." + enc.EncodeToString([]byte(claims)) + ".sig"
}

func TestTokenExpiry(t *testing.T) {
	observedAt := time.Unix(10_000, 0)

	t.Run("lifetime from iat and exp is anchored to the local clock", func(t *testing.T) {
		// Server clock is far off; only the lifetime (exp - iat) matters.
		got, ok := tokenExpiry(testJWT(`{"iat":500,"exp":800}`), observedAt)
		if !ok {
			t.Fatal("expected expiry to be parsed")
		}
		if want := observedAt.Add(300 * time.Second); !got.Equal(want) {
			t.Fatalf("expiry = %v, want %v", got, want)
		}
	})

	t.Run("exp without iat is used as is", func(t *testing.T) {
		got, ok := tokenExpiry(testJWT(`{"exp":12345}`), observedAt)
		if !ok || !got.Equal(time.Unix(12345, 0)) {
			t.Fatalf("expiry = %v, %v; want %v, true", got, ok, time.Unix(12345, 0))
		}
	})

	for name, token := range map[string]string{
		"empty":           "",
		"not a jwt":       "opaque-token",
		"invalid base64":  "a.!!!.c",
		"invalid json":    "a." + base64.RawURLEncoding.EncodeToString([]byte("nope")) + ".c",
		"no exp":          testJWT(`{"iat":500}`),
		"non-positive tl": testJWT(`{"iat":800,"exp":800}`),
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := tokenExpiry(token, observedAt); ok {
				t.Fatal("expected no expiry")
			}
		})
	}
}

// newTokenServer returns a fake Keycloak that issues a distinct JWT with the
// given lifetime on every token request and counts those requests.
func newTokenServer(t *testing.T, lifetime time.Duration, tokenRequests *int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/protocol/openid-connect/token"):
			n := atomic.AddInt32(tokenRequests, 1)
			iat := time.Now().Unix()
			token := testJWT(fmt.Sprintf(`{"jti":"%d","iat":%d,"exp":%d}`, n, iat, iat+int64(lifetime/time.Second)))
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"refresh_token":"r","token_type":"bearer","expires_in":%d}`, token, int(lifetime/time.Second))
		case strings.Contains(r.URL.Path, "/serverinfo"):
			_, _ = w.Write([]byte(`{"systemInfo":{"version":"26.0.0"}}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestTokenTrackerRefreshesBeforeExpiry(t *testing.T) {
	var tokenRequests int32
	srv := newTokenServer(t, 300*time.Second, &tokenRequests)

	ctx := context.Background()
	kc, err := keycloak.NewKeycloakClient(ctx, srv.URL, "", "", "admin-cli", "", "master",
		"admin", "admin", "", "", "", "", "", true, 5, "", true, "", "", "test", false, nil, "")
	if err != nil {
		t.Fatalf("NewKeycloakClient: %v", err)
	}
	if got := atomic.LoadInt32(&tokenRequests); got != 1 {
		t.Fatalf("token requests after login = %d, want 1", got)
	}

	now := time.Unix(1_000_000, 0)
	tracker := newTokenTracker(func() time.Time { return now })
	tracker.observe(kc)
	firstToken := ExtractAccessToken(kc)

	// Well before expiry: no refresh.
	now = now.Add(200 * time.Second)
	if err := tracker.refreshIfExpiring(ctx, kc); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&tokenRequests); got != 1 {
		t.Fatalf("token requests = %d, want 1 (token still fresh)", got)
	}

	// Within the refresh leeway: refreshed proactively.
	now = now.Add(75 * time.Second)
	if err := tracker.refreshIfExpiring(ctx, kc); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&tokenRequests); got != 2 {
		t.Fatalf("token requests = %d, want 2 (proactive refresh)", got)
	}
	if ExtractAccessToken(kc) == firstToken {
		t.Fatal("expected a new access token after proactive refresh")
	}

	// The new token is tracked from the time it was obtained.
	now = now.Add(200 * time.Second)
	if err := tracker.refreshIfExpiring(ctx, kc); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&tokenRequests); got != 2 {
		t.Fatalf("token requests = %d, want 2 (new token still fresh)", got)
	}
}

func TestTokenTrackerShortLivedTokenLeeway(t *testing.T) {
	var tokenRequests int32
	srv := newTokenServer(t, 20*time.Second, &tokenRequests)

	ctx := context.Background()
	kc, err := keycloak.NewKeycloakClient(ctx, srv.URL, "", "", "admin-cli", "", "master",
		"admin", "admin", "", "", "", "", "", true, 5, "", true, "", "", "test", false, nil, "")
	if err != nil {
		t.Fatalf("NewKeycloakClient: %v", err)
	}

	now := time.Unix(1_000_000, 0)
	tracker := newTokenTracker(func() time.Time { return now })
	refreshAt, ok := tracker.observe(kc)
	if !ok {
		t.Fatal("expected token expiry to be tracked")
	}
	// Leeway is capped at half the lifetime for short-lived tokens.
	if want := now.Add(10 * time.Second); !refreshAt.Equal(want) {
		t.Fatalf("refreshAt = %v, want %v", refreshAt, want)
	}
	if err := tracker.refreshIfExpiring(ctx, kc); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&tokenRequests); got != 1 {
		t.Fatalf("token requests = %d, want 1 (no refresh on every use)", got)
	}
}

func TestTokenTrackerUnknownExpiryIsNoop(t *testing.T) {
	// Client without a token (initial_login=false, offline): nothing to do.
	kc, err := keycloak.NewKeycloakClient(context.Background(), "http://127.0.0.1:1", "", "", "admin-cli", "", "master",
		"", "", "", "", "", "", "", false, 5, "", true, "", "", "test", false, nil, "")
	if err != nil {
		t.Fatalf("NewKeycloakClient: %v", err)
	}
	tracker := newTokenTracker(time.Now)
	if _, ok := tracker.observe(kc); ok {
		t.Fatal("expected no tracked expiry for a client without a token")
	}
	if err := tracker.refreshIfExpiring(context.Background(), kc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracker.refreshIfExpiring(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error for nil client: %v", err)
	}
}
