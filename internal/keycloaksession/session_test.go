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
