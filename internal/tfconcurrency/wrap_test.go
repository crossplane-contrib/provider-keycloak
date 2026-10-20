/*
Copyright 2021 Upbound Inc.
*/

package tfconcurrency

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestWrapResourceSubstitutesPooledClient(t *testing.T) {
	primary := newOfflineClient(t)
	pool := NewPool(2, offlineFactory())
	Register(primary, pool)
	defer Unregister(primary)

	var got any
	r := &schema.Resource{
		ReadContext: func(_ context.Context, _ *schema.ResourceData, meta any) diag.Diagnostics {
			got = meta
			return nil
		},
	}
	WrapResource(r)

	if diags := r.ReadContext(context.Background(), nil, primary); diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	gotClient, ok := got.(*keycloak.KeycloakClient)
	if !ok || gotClient == nil {
		t.Fatalf("callback did not receive a *keycloak.KeycloakClient, got %T", got)
	}
	if gotClient == primary {
		t.Fatal("callback received the shared primary client; expected a pooled substitute")
	}
}

// TestWrapResourceAllowsBoundedConcurrency proves the fix preserves concurrency
// (more than one callback runs at once) while bounding it to the pool size.
func TestWrapResourceAllowsBoundedConcurrency(t *testing.T) {
	const size = 3
	primary := newOfflineClient(t)
	pool := NewPool(size, offlineFactory())
	Register(primary, pool)
	defer Unregister(primary)

	var active, maxActive int32
	entered := make(chan struct{}, 16)
	release := make(chan struct{})

	r := &schema.Resource{
		ReadContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
			n := atomic.AddInt32(&active, 1)
			for {
				old := atomic.LoadInt32(&maxActive)
				if n <= old || atomic.CompareAndSwapInt32(&maxActive, old, n) {
					break
				}
			}
			entered <- struct{}{}
			<-release
			atomic.AddInt32(&active, -1)
			return nil
		},
	}
	WrapResource(r)

	const callers = 5
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.ReadContext(context.Background(), nil, primary)
		}()
	}

	for i := 0; i < size; i++ {
		select {
		case <-entered:
		case <-time.After(2 * time.Second):
			t.Fatalf("only %d/%d callbacks entered; expected %d concurrent", i, callers, size)
		}
	}
	select {
	case <-entered:
		t.Fatal("more than pool-size callbacks ran concurrently")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	wg.Wait()

	if maxActive != size {
		t.Fatalf("max concurrent callbacks = %d, want %d", maxActive, size)
	}
}

// TestWrapResourceNoRaceUnderConcurrentLogin drives many concurrent operations
// that each trigger a login + request through the wrapper. With the shared
// client this races on initialLogin/clientCredentials/version; with the pool
// each operation uses its own client, so `go test -race` stays clean.
func TestWrapResourceNoRaceUnderConcurrentLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/protocol/openid-connect/token"):
			_, _ = w.Write([]byte(`{"access_token":"a","refresh_token":"r","token_type":"bearer","expires_in":300}`))
		case strings.Contains(r.URL.Path, "/serverinfo"):
			_, _ = w.Write([]byte(`{"systemInfo":{"version":"26.0.0"}}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	factory := func(ctx context.Context) (*keycloak.KeycloakClient, error) {
		return keycloak.NewKeycloakClient(
			ctx,
			srv.URL, "", "", "admin-cli", "", "master",
			"admin", "admin", "", "", "", "", "",
			false, 5, "", true, "", "", "test", false, nil, "",
		)
	}

	primary, err := factory(context.Background())
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	pool := NewPool(4, factory)
	Register(primary, pool)
	defer Unregister(primary)

	r := &schema.Resource{
		ReadContext: func(ctx context.Context, _ *schema.ResourceData, meta any) diag.Diagnostics {
			kc := meta.(*keycloak.KeycloakClient)
			if _, err := kc.GetServerInfo(ctx); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
	WrapResource(r)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.ReadContext(context.Background(), nil, primary)
		}()
	}
	wg.Wait()
}

// TestWrapResourceRefreshesTokenBeforeExpiry proves that a borrowed client's
// access token is refreshed before it expires, so routine token expiry never
// produces a 401 from Keycloak (the upstream refresh-on-401 stays a fallback).
func TestWrapResourceRefreshesTokenBeforeExpiry(t *testing.T) {
	const lifetime = time.Second
	var tokenRequests, unauthorized int32
	enc := base64.RawURLEncoding
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/protocol/openid-connect/token") {
			n := atomic.AddInt32(&tokenRequests, 1)
			iat := float64(time.Now().UnixNano()) / float64(time.Second)
			claims := fmt.Sprintf(`{"jti":"%d","iat":%f,"exp":%f}`, n, iat, iat+lifetime.Seconds())
			token := enc.EncodeToString([]byte(`{"alg":"none"}`)) + "." + enc.EncodeToString([]byte(claims)) + ".sig"
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"bearer","expires_in":1}`, token)
			return
		}
		// Reject expired tokens like Keycloak does.
		var claims struct {
			Exp float64 `json:"exp"`
		}
		parts := strings.Split(strings.TrimPrefix(r.Header.Get("Authorization"), "bearer "), ".")
		payload, err := enc.DecodeString(parts[min(1, len(parts)-1)])
		if err != nil || json.Unmarshal(payload, &claims) != nil || float64(time.Now().UnixNano())/float64(time.Second) > claims.Exp {
			atomic.AddInt32(&unauthorized, 1)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"systemInfo":{"version":"26.0.0"}}`))
	}))
	defer srv.Close()

	factory := func(ctx context.Context) (*keycloak.KeycloakClient, error) {
		return keycloak.NewKeycloakClient(
			ctx,
			srv.URL, "", "", "admin-cli", "", "master",
			"admin", "admin", "", "", "", "", "",
			true, 5, "", true, "", "", "test", false, nil, "",
		)
	}

	primary, err := factory(context.Background())
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	pool := NewPool(1, factory)
	pool.Seed(primary)
	Register(primary, pool)
	defer Unregister(primary)

	r := &schema.Resource{
		ReadContext: func(ctx context.Context, _ *schema.ResourceData, meta any) diag.Diagnostics {
			if _, err := meta.(*keycloak.KeycloakClient).GetServerInfo(ctx); err != nil {
				return diag.FromErr(err)
			}
			return nil
		},
	}
	WrapResource(r)

	for i := 0; i < 3; i++ {
		if diags := r.ReadContext(context.Background(), nil, primary); diags.HasError() {
			t.Fatalf("read %d: %v", i, diags)
		}
		// Let the current token expire on the server side.
		time.Sleep(lifetime + 100*time.Millisecond)
	}

	if got := atomic.LoadInt32(&unauthorized); got != 0 {
		t.Fatalf("got %d 401 responses, want 0 (token should be refreshed proactively)", got)
	}
	if got := atomic.LoadInt32(&tokenRequests); got != 3 {
		t.Fatalf("token requests = %d, want 3 (initial login + 2 proactive refreshes)", got)
	}
}
