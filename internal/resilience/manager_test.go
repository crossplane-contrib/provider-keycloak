/*
Copyright 2022 Upbound Inc.
*/

package resilience

import (
	"context"
	"errors"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type fakeController struct {
	start          func(ctx context.Context) error
	leaderElection bool
}

func (f *fakeController) Reconcile(context.Context, reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
func (f *fakeController) Watch(source.TypedSource[reconcile.Request]) error { return nil }
func (f *fakeController) Start(ctx context.Context) error                   { return f.start(ctx) }
func (f *fakeController) GetLogger() logr.Logger                            { return logr.Discard() }
func (f *fakeController) NeedLeaderElection() bool                          { return f.leaderElection }

var _ controller.Controller = &fakeController{}
var _ manager.LeaderElectionRunnable = &fakeController{}

func TestTolerantControllerStart(t *testing.T) {
	boom := errors.New("failed to wait for caches to sync")

	t.Run("StartupFailureIsNotFatal", func(t *testing.T) {
		c := &tolerantController{
			Controller: &fakeController{start: func(context.Context) error { return boom }},
			log:        logging.NewNopLogger(),
		}
		if err := c.Start(context.Background()); err != nil {
			t.Errorf("a controller startup failure must not propagate to the manager, got: %v", err)
		}
	})

	t.Run("ShutdownErrorPropagates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		c := &tolerantController{
			Controller: &fakeController{start: func(context.Context) error { return boom }},
			log:        logging.NewNopLogger(),
		}
		if err := c.Start(ctx); !errors.Is(err, boom) {
			t.Errorf("an error during shutdown must propagate, got: %v", err)
		}
	})

	t.Run("SuccessfulStart", func(t *testing.T) {
		c := &tolerantController{
			Controller: &fakeController{start: func(context.Context) error { return nil }},
			log:        logging.NewNopLogger(),
		}
		if err := c.Start(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("LeaderElectionIsPreserved", func(t *testing.T) {
		for _, need := range []bool{true, false} {
			c := &tolerantController{Controller: &fakeController{leaderElection: need}}
			if got := c.NeedLeaderElection(); got != need {
				t.Errorf("NeedLeaderElection: want %v, got %v", need, got)
			}
		}
	})
}
