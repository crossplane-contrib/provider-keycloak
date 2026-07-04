/*
Copyright 2021 Upbound Inc.
*/

// Package tfconcurrency makes the Terraform-Plugin-SDK no-fork execution path
// safe for concurrent use in this provider.
//
// The embedded github.com/keycloak/terraform-provider-keycloak client
// (*keycloak.KeycloakClient) is NOT safe for concurrent use: login(),
// Refresh() and sendRequest() mutate shared fields (clientCredentials,
// version, initialLogin) without synchronization, and the upstream Mutex
// field is not wired into those paths. Because upjet shares a single cached
// client (the provider "meta") across all concurrently reconciling managed
// resources, concurrent operations race on that shared client and can crash
// the provider with "fatal error: concurrent map writes".
//
// The terraform-plugin-sdk *schema.Resource itself IS safe for concurrent
// use (Apply/RefreshWithoutUpgrade/Diff build fresh, local ResourceData and
// deep-copy the schema before any mutation), so the only thing that must be
// protected is the client. Rather than serializing everything behind a single
// lock, this package gives each in-flight operation its OWN client borrowed
// from a small, bounded, per-configuration pool. Concurrency is therefore
// preserved up to the pool size while every client instance only ever has a
// single user at a time.
package tfconcurrency

import (
	"context"
	"sync"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// ClientFactory creates a new, independent *keycloak.KeycloakClient for a
// single provider configuration. Implementations typically build the client
// exactly the way the provider setup does (same config, same login), so that
// pooled clients are indistinguishable from the primary one.
type ClientFactory func(ctx context.Context) (*keycloak.KeycloakClient, error)

// Pool is a bounded pool of *keycloak.KeycloakClient instances that all target
// the same provider configuration. At most Size clients are ever created and
// at most Size operations run concurrently; additional borrowers block until a
// client is returned (natural back-pressure). Clients are created lazily and
// their creation is serialized to avoid a login/token "storm" against Keycloak.
//
// A borrowed client is used by exactly one goroutine for the duration of a
// single, synchronous Terraform SDK callback, which is why per-client state
// never races.
type Pool struct {
	// sem bounds the number of simultaneously-borrowed clients to the pool
	// size. Acquiring a slot (sending) happens on Borrow; releasing (receiving)
	// happens on Return.
	sem chan struct{}

	factory ClientFactory

	// createMu serializes client construction so that a cold pool does not fire
	// N concurrent logins at once.
	createMu sync.Mutex

	// mu guards idle and all.
	mu   sync.Mutex
	idle []*keycloak.KeycloakClient // returned clients available for reuse
	all  []*keycloak.KeycloakClient // every client the pool owns (for Close)
}

// NewPool returns a pool that will hold at most size clients, created on
// demand by factory. size is clamped to a minimum of 1.
func NewPool(size int, factory ClientFactory) *Pool {
	if size < 1 {
		size = 1
	}
	return &Pool{
		sem:     make(chan struct{}, size),
		factory: factory,
	}
}

// Seed adds an already-constructed client to the pool as an idle, reusable
// member. It is used to donate the provider's primary (already logged-in)
// client to the pool so it is not left idle. Seed does not consume a borrow
// slot. It is a no-op for a nil client.
func (p *Pool) Seed(c *keycloak.KeycloakClient) {
	if p == nil || c == nil {
		return
	}
	p.mu.Lock()
	p.idle = append(p.idle, c)
	p.all = append(p.all, c)
	p.mu.Unlock()
}

// Borrow returns a client for exclusive use by the caller until Return is
// called. It blocks if the pool is at capacity, returning early only if ctx is
// cancelled. On any error the borrow slot is released so the pool does not leak
// capacity.
func (p *Pool) Borrow(ctx context.Context) (*keycloak.KeycloakClient, error) {
	// Acquire a capacity slot (bounds concurrency to the pool size).
	select {
	case p.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Reuse an idle client if one is available.
	p.mu.Lock()
	if n := len(p.idle); n > 0 {
		c := p.idle[n-1]
		p.idle[n-1] = nil
		p.idle = p.idle[:n-1]
		p.mu.Unlock()
		return c, nil
	}
	p.mu.Unlock()

	// None idle: create a new one (serialized). We hold the slot throughout, so
	// the total number of clients can never exceed the pool size.
	c, err := p.create(ctx)
	if err != nil {
		<-p.sem // release the slot we acquired
		return nil, err
	}
	return c, nil
}

func (p *Pool) create(ctx context.Context) (*keycloak.KeycloakClient, error) {
	p.createMu.Lock()
	defer p.createMu.Unlock()

	c, err := p.factory(ctx)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.all = append(p.all, c)
	p.mu.Unlock()
	return c, nil
}

// Return releases a previously borrowed client back to the pool for reuse and
// frees a capacity slot. Returning a nil client only frees the slot.
func (p *Pool) Return(c *keycloak.KeycloakClient) {
	if c != nil {
		p.mu.Lock()
		p.idle = append(p.idle, c)
		p.mu.Unlock()
	}
	<-p.sem
}

// Clients returns a snapshot of every client the pool owns. It is used for
// lifecycle operations such as logout on shutdown.
func (p *Pool) Clients() []*keycloak.KeycloakClient {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]*keycloak.KeycloakClient, len(p.all))
	copy(out, p.all)
	return out
}
