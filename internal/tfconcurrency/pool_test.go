/*
Copyright 2021 Upbound Inc.
*/

package tfconcurrency

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// newOfflineClient builds a real *keycloak.KeycloakClient that performs no
// network I/O: with no credentials and initialLogin=false the constructor skips
// the login round-trip.
func newOfflineClient(t *testing.T) *keycloak.KeycloakClient {
	t.Helper()
	c, err := newOfflineClientErr(context.Background())
	if err != nil {
		t.Fatalf("NewKeycloakClient: %v", err)
	}
	return c
}

func newOfflineClientErr(ctx context.Context) (*keycloak.KeycloakClient, error) {
	return keycloak.NewKeycloakClient(
		ctx,
		"http://127.0.0.1:1", "", "", "admin-cli", "", "master",
		"", "", "", "", "", "", "",
		false, 5, "", true, "", "", "test", false, nil,
	)
}

func offlineFactory() ClientFactory {
	return func(ctx context.Context) (*keycloak.KeycloakClient, error) {
		return newOfflineClientErr(ctx)
	}
}

func TestPoolReusesReturnedClient(t *testing.T) {
	p := NewPool(2, offlineFactory())
	ctx := context.Background()

	c1, err := p.Borrow(ctx)
	if err != nil {
		t.Fatal(err)
	}
	p.Return(c1)

	c2, err := p.Borrow(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if c1 != c2 {
		t.Fatal("expected the returned client to be reused")
	}
	p.Return(c2)

	if got := len(p.Clients()); got != 1 {
		t.Fatalf("pool created %d clients, want 1", got)
	}
}

func TestPoolBoundsToSize(t *testing.T) {
	p := NewPool(2, offlineFactory())
	ctx := context.Background()

	c1, _ := p.Borrow(ctx)
	c2, _ := p.Borrow(ctx)

	res := make(chan *keycloak.KeycloakClient, 1)
	go func() {
		c, _ := p.Borrow(ctx)
		res <- c
	}()

	select {
	case <-res:
		t.Fatal("third borrow should block while the pool is at capacity")
	case <-time.After(50 * time.Millisecond):
	}

	p.Return(c1)
	select {
	case <-res:
	case <-time.After(time.Second):
		t.Fatal("third borrow should unblock after a return")
	}
	p.Return(c2)

	if got := len(p.Clients()); got > 2 {
		t.Fatalf("pool created %d clients, want <= 2", got)
	}
}

func TestPoolBorrowRespectsContextCancellation(t *testing.T) {
	p := NewPool(1, offlineFactory())
	ctx := context.Background()

	c1, _ := p.Borrow(ctx)
	defer p.Return(c1)

	cctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := p.Borrow(cctx); err == nil {
		t.Fatal("expected Borrow to fail when context is cancelled and pool is saturated")
	}
}

func TestPoolConcurrentBorrowReturn(t *testing.T) {
	const size = 4
	p := NewPool(size, offlineFactory())

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := p.Borrow(context.Background())
			if err != nil {
				return
			}
			time.Sleep(time.Millisecond)
			p.Return(c)
		}()
	}
	wg.Wait()

	if got := len(p.Clients()); got > size {
		t.Fatalf("pool created %d clients, want <= %d", got, size)
	}
}
