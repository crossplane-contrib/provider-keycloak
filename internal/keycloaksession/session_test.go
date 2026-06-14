/*
Copyright 2024 Upbound Inc.
*/

package keycloaksession

import (
	"testing"
)

func TestConfigCacheKey(t *testing.T) {
	t.Run("same config produces same key", func(t *testing.T) {
		cfg := map[string]any{
			"url":       "https://keycloak.example.com",
			"client_id": "admin-cli",
			"username":  "admin",
			"password":  "s3cr3t",
		}
		k1 := ConfigCacheKey(cfg)
		k2 := ConfigCacheKey(cfg)
		if k1 != k2 {
			t.Fatalf("expected same key, got %q and %q", k1, k2)
		}
	})

	t.Run("key is order-independent", func(t *testing.T) {
		cfg1 := map[string]any{"url": "https://keycloak.example.com", "client_id": "admin-cli"}
		cfg2 := map[string]any{"client_id": "admin-cli", "url": "https://keycloak.example.com"}
		if ConfigCacheKey(cfg1) != ConfigCacheKey(cfg2) {
			t.Fatal("expected same key for maps with identical contents but different insertion order")
		}
	})

	t.Run("different configs produce different keys", func(t *testing.T) {
		cfg1 := map[string]any{"url": "https://keycloak1.example.com", "client_id": "admin-cli"}
		cfg2 := map[string]any{"url": "https://keycloak2.example.com", "client_id": "admin-cli"}
		if ConfigCacheKey(cfg1) == ConfigCacheKey(cfg2) {
			t.Fatal("expected different keys for different configurations")
		}
	})

	t.Run("stable with nested map values", func(t *testing.T) {
		cfg := map[string]any{
			"url":       "https://keycloak.example.com",
			"client_id": "admin-cli",
			"additional_headers": map[string]any{
				"X-Custom-Header": "value1",
				"Authorization":   "******",
				"X-Request-Id":    "12345",
			},
		}
		// Run multiple times to exercise Go's randomized map iteration.
		first := ConfigCacheKey(cfg)
		for i := 0; i < 100; i++ {
			if got := ConfigCacheKey(cfg); got != first {
				t.Fatalf("iteration %d: expected stable key %q, got %q", i, first, got)
			}
		}
	})
}

func TestIsPasswordGrant(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		want   bool
	}{
		{
			name:   "password grant",
			config: map[string]any{"username": "admin", "password": "secret"},
			want:   true,
		},
		{
			name:   "client credentials",
			config: map[string]any{"client_id": "cli", "client_secret": "s"},
			want:   false,
		},
		{
			name:   "no auth fields",
			config: map[string]any{"client_id": "cli"},
			want:   false,
		},
		{
			name:   "username only",
			config: map[string]any{"username": "admin"},
			want:   false,
		},
		{
			name:   "password only",
			config: map[string]any{"password": "secret"},
			want:   false,
		},
		{
			name:   "empty strings",
			config: map[string]any{"username": "", "password": ""},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPasswordGrant(tt.config); got != tt.want {
				t.Fatalf("IsPasswordGrant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractRefreshToken_nil(t *testing.T) {
	if token := ExtractRefreshToken(nil); token != "" {
		t.Fatalf("expected empty token for nil client, got %q", token)
	}
}

func TestLogoutConfig(t *testing.T) {
	full := map[string]any{
		"url":           "https://keycloak.example.com",
		"base_path":     "/auth",
		"realm":         "myrealm",
		"client_id":     "admin-cli",
		"client_secret": "secret",
		"username":      "admin",
		"password":      "pass",
		"access_token":  "should-not-be-retained",
		"jwt_token":     "should-not-be-retained",
		"additional_headers": map[string]any{
			"X-Custom": "value",
		},
	}
	got := LogoutConfig(full)

	// Should contain only logout-relevant keys
	expectedKeys := []string{"url", "base_path", "realm", "client_id", "client_secret", "username", "password"}
	if len(got) != len(expectedKeys) {
		t.Fatalf("expected %d keys, got %d: %v", len(expectedKeys), len(got), got)
	}
	for _, k := range expectedKeys {
		if _, ok := got[k]; !ok {
			t.Fatalf("expected key %q in logout config", k)
		}
	}
	// Sensitive keys not needed for logout should be absent
	for _, k := range []string{"access_token", "jwt_token", "additional_headers"} {
		if _, ok := got[k]; ok {
			t.Fatalf("key %q should not be in logout config", k)
		}
	}
}
