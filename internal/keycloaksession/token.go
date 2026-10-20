/*
Copyright 2024 Upbound Inc.
*/

package keycloaksession

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// tokenRefreshLeeway is how long before its expiry an access token is
// proactively refreshed. It absorbs the delay between the token being
// issued and observed as well as the duration of the request that is
// about to be sent. For short-lived tokens it is capped at half the
// token lifetime so that a token is not refreshed on every use.
const tokenRefreshLeeway = 30 * time.Second

// tokenState is the tracked expiry information of a client's current
// access token. Only a fingerprint of the token is kept, never the token.
type tokenState struct {
	fingerprint [sha256.Size]byte
	refreshAt   time.Time
}

// tokenTracker tracks, per *keycloak.KeycloakClient, when its current
// access token must be refreshed.
//
// The upstream client only refreshes its token after a request has been
// rejected with 401/403, which produces one failed request (and a failed
// authentication entry in Keycloak's logs) per token lifetime. The tracker
// lets callers refresh the token before it expires instead, keeping the
// reactive refresh as a fallback only.
//
// The expiry is computed on the local clock: the token lifetime is taken
// from its "exp" and "iat" claims (i.e. expires_in) and added to the local
// time at which the token was first observed, so clock skew between the
// provider and Keycloak does not matter.
type tokenTracker struct {
	mu     sync.Mutex
	states map[*keycloak.KeycloakClient]tokenState
	now    func() time.Time
}

var defaultTokenTracker = newTokenTracker(time.Now)

func newTokenTracker(now func() time.Time) *tokenTracker {
	return &tokenTracker{
		states: map[*keycloak.KeycloakClient]tokenState{},
		now:    now,
	}
}

// ObserveToken records the expiry of the client's current access token if
// it has not been seen before. It should be called right after the client
// may have obtained a new token (creation, end of a request) so that the
// token's lifetime is anchored to an accurate local time. The caller must
// have exclusive use of kcClient.
func ObserveToken(kcClient *keycloak.KeycloakClient) {
	defaultTokenTracker.observe(kcClient)
}

// RefreshTokenIfExpiring proactively refreshes the client's access token
// when it is about to expire. It is a no-op when the token is still fresh
// or its expiry is unknown (e.g. no token yet, or not a JWT). Any error is
// returned to the caller; since the upstream client still refreshes on a
// 401, callers may treat it as best-effort. The caller must have exclusive
// use of kcClient.
func RefreshTokenIfExpiring(ctx context.Context, kcClient *keycloak.KeycloakClient) error {
	return defaultTokenTracker.refreshIfExpiring(ctx, kcClient)
}

// ForgetToken drops any tracked state for the client.
func ForgetToken(kcClient *keycloak.KeycloakClient) {
	defaultTokenTracker.forget(kcClient)
}

func (t *tokenTracker) refreshIfExpiring(ctx context.Context, kcClient *keycloak.KeycloakClient) error {
	refreshAt, ok := t.observe(kcClient)
	if !ok || t.now().Before(refreshAt) {
		return nil
	}
	if err := kcClient.Refresh(ctx); err != nil {
		return err
	}
	t.observe(kcClient)
	return nil
}

// observe returns the time at which the client's current access token
// should be refreshed, and false if it is unknown.
func (t *tokenTracker) observe(kcClient *keycloak.KeycloakClient) (time.Time, bool) {
	if kcClient == nil {
		return time.Time{}, false
	}
	token := ExtractAccessToken(kcClient)

	t.mu.Lock()
	defer t.mu.Unlock()

	if token == "" {
		delete(t.states, kcClient)
		return time.Time{}, false
	}
	fp := sha256.Sum256([]byte(token))
	if s, ok := t.states[kcClient]; ok && s.fingerprint == fp {
		return s.refreshAt, true
	}

	now := t.now()
	expiry, ok := tokenExpiry(token, now)
	if !ok {
		delete(t.states, kcClient)
		return time.Time{}, false
	}
	leeway := tokenRefreshLeeway
	if half := max(expiry.Sub(now)/2, 0); half < leeway {
		leeway = half
	}
	refreshAt := expiry.Add(-leeway)
	t.states[kcClient] = tokenState{fingerprint: fp, refreshAt: refreshAt}
	return refreshAt, true
}

func (t *tokenTracker) forget(kcClient *keycloak.KeycloakClient) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.states, kcClient)
}

// tokenExpiry returns the local-clock expiry of a JWT access token observed
// at observedAt. When the token carries both "iat" and "exp", the expiry is
// observedAt plus the token lifetime (exp - iat), which is independent of
// clock skew. Otherwise "exp" is used as is. The token signature is not
// verified: the token is only inspected to schedule its refresh.
func tokenExpiry(token string, observedAt time.Time) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return time.Time{}, false
	}
	var claims struct {
		Exp *float64 `json:"exp"`
		Iat *float64 `json:"iat"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == nil {
		return time.Time{}, false
	}
	exp := time.Unix(0, int64(*claims.Exp*float64(time.Second)))
	if claims.Iat == nil {
		return exp, true
	}
	lifetime := time.Duration((*claims.Exp - *claims.Iat) * float64(time.Second))
	if lifetime <= 0 {
		return time.Time{}, false
	}
	return observedAt.Add(lifetime), true
}
