/*
Copyright 2021 Upbound Inc.
*/

package tfconcurrency

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// registry maps a provider's primary (cached) *keycloak.KeycloakClient — the
// value upjet passes as the schema callback `meta` — to the per-configuration
// Pool that backs it. It is populated by the provider setup for every unique
// configuration.
var registry sync.Map // map[*keycloak.KeycloakClient]*Pool

// fallbackMu holds a per-client mutex used only when no pool is registered for
// a given meta client (should not happen in production). It guarantees the
// wrapped callbacks are still race-free (serialized) as a defensive fallback.
var fallbackMu sync.Map // map[*keycloak.KeycloakClient]*sync.Mutex

// Register associates a primary client (the schema callback meta) with its
// pool. Both must be non-nil.
func Register(meta *keycloak.KeycloakClient, p *Pool) {
	if meta == nil || p == nil {
		return
	}
	registry.Store(meta, p)
}

// Unregister removes the pool association for a primary client.
func Unregister(meta *keycloak.KeycloakClient) {
	if meta == nil {
		return
	}
	registry.Delete(meta)
	fallbackMu.Delete(meta)
}

func poolFor(meta any) *Pool {
	kc, ok := meta.(*keycloak.KeycloakClient)
	if !ok || kc == nil {
		return nil
	}
	if v, ok := registry.Load(kc); ok {
		return v.(*Pool)
	}
	return nil
}

// borrow selects the client a wrapped callback should use in place of the
// shared meta. When a pool is registered it borrows a dedicated client (so the
// callback runs concurrently with others, each on its own client). When no pool
// is registered it returns the original meta guarded by a per-client mutex so
// the callback is still serialized and therefore race-free. The returned
// release function must always be called (defer) exactly once.
func borrow(ctx context.Context, meta any) (client any, release func(), err error) {
	if pool := poolFor(meta); pool != nil {
		c, berr := pool.Borrow(ctx)
		if berr != nil {
			return nil, func() {}, berr
		}
		return c, func() { pool.Return(c) }, nil
	}

	kc, ok := meta.(*keycloak.KeycloakClient)
	if !ok || kc == nil {
		return meta, func() {}, nil
	}
	v, _ := fallbackMu.LoadOrStore(kc, &sync.Mutex{})
	m := v.(*sync.Mutex)
	m.Lock()
	return meta, m.Unlock, nil
}

// WrapProvider wraps every resource and data source in p so that their
// Terraform SDK callbacks run against a per-configuration pooled client instead
// of the single shared meta client. It must be called on the same
// *schema.Provider whose resources upjet drives at runtime.
func WrapProvider(p *schema.Provider) {
	if p == nil {
		return
	}
	for _, r := range p.ResourcesMap {
		WrapResource(r)
	}
	for _, r := range p.DataSourcesMap {
		WrapResource(r)
	}
}

// WrapResource wraps the context-aware CRUD, CustomizeDiff and import callbacks
// of a single resource. Each wrapped callback borrows a dedicated client for
// the (synchronous) duration of the call and substitutes it as the meta
// argument. Because each upjet operation (Apply/RefreshWithoutUpgrade/Diff)
// invokes exactly one such callback with no nesting, a single borrow per call
// is deadlock-free. The deprecated non-context CRUD fields are not used by
// terraform-provider-keycloak and are intentionally left untouched.
func WrapResource(r *schema.Resource) {
	if r == nil {
		return
	}

	r.CreateContext = wrapContext(r.CreateContext)
	r.ReadContext = wrapContext(r.ReadContext)
	r.UpdateContext = wrapContext(r.UpdateContext)
	r.DeleteContext = wrapContext(r.DeleteContext)

	r.CreateWithoutTimeout = wrapContext(r.CreateWithoutTimeout)
	r.ReadWithoutTimeout = wrapContext(r.ReadWithoutTimeout)
	r.UpdateWithoutTimeout = wrapContext(r.UpdateWithoutTimeout)
	r.DeleteWithoutTimeout = wrapContext(r.DeleteWithoutTimeout)

	r.CustomizeDiff = wrapCustomizeDiff(r.CustomizeDiff)

	if r.Importer != nil {
		r.Importer.StateContext = wrapStateContext(r.Importer.StateContext)
	}
}

// wrapContext wraps any context-aware CRUD callback. It uses the unnamed
// underlying signature so the result is assignable to each of the distinct
// named SDK types (CreateContextFunc, ReadContextFunc, ...).
func wrapContext(orig func(context.Context, *schema.ResourceData, any) diag.Diagnostics) func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
	if orig == nil {
		return nil
	}
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		client, release, err := borrow(ctx, meta)
		if err != nil {
			return diag.FromErr(err)
		}
		defer release()
		return orig(ctx, d, client)
	}
}

func wrapCustomizeDiff(orig func(context.Context, *schema.ResourceDiff, any) error) func(context.Context, *schema.ResourceDiff, any) error {
	if orig == nil {
		return nil
	}
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		client, release, err := borrow(ctx, meta)
		if err != nil {
			return err
		}
		defer release()
		return orig(ctx, d, client)
	}
}

func wrapStateContext(orig func(context.Context, *schema.ResourceData, any) ([]*schema.ResourceData, error)) func(context.Context, *schema.ResourceData, any) ([]*schema.ResourceData, error) {
	if orig == nil {
		return nil
	}
	return func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
		client, release, err := borrow(ctx, meta)
		if err != nil {
			return nil, err
		}
		defer release()
		return orig(ctx, d, client)
	}
}
