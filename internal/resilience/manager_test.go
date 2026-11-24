/*
Copyright 2022 Upbound Inc.
*/

package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

type fakeRunnable struct {
	start func(ctx context.Context) error
}

func (f *fakeRunnable) Start(ctx context.Context) error { return f.start(ctx) }

func TestRetryingRunnableStart(t *testing.T) {
	noMatch := &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: "openidgroup.keycloak.crossplane.io", Kind: "FullNameProtocolMapper"}}

	t.Run("MissingKindIsRetried", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		r := &retryingRunnable{
			Runnable: &fakeRunnable{start: func(context.Context) error {
				calls++
				if calls < 3 {
					return noMatch
				}
				cancel()
				return nil
			}},
			log: logging.NewNopLogger(),
		}
		if err := r.Start(ctx); err != nil {
			t.Errorf("a missing kind must not propagate to the manager, got: %v", err)
		}
		if calls != 3 {
			t.Errorf("want 3 starts, got %d", calls)
		}
	})

	t.Run("OtherErrorIsRetried", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		r := &retryingRunnable{
			Runnable: &fakeRunnable{start: func(context.Context) error {
				calls++
				if calls == 1 {
					return errors.New("failed to list managed resources")
				}
				return nil
			}},
			log: logging.NewNopLogger(),
		}
		if err := r.Start(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if calls != 2 {
			t.Errorf("want 2 starts, got %d", calls)
		}
	})

	t.Run("CancelledWhileWaiting", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		r := &retryingRunnable{
			Runnable: &fakeRunnable{start: func(context.Context) error {
				time.AfterFunc(10*time.Millisecond, cancel)
				return noMatch
			}},
			log:      logging.NewNopLogger(),
			interval: time.Hour,
		}
		if err := r.Start(ctx); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("ShutdownErrorPropagates", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		r := &retryingRunnable{
			Runnable: &fakeRunnable{start: func(context.Context) error {
				cancel()
				return noMatch
			}},
			log:      logging.NewNopLogger(),
			interval: time.Hour,
		}
		if err := r.Start(ctx); !errors.Is(err, noMatch) {
			t.Errorf("an error during shutdown must propagate, got: %v", err)
		}
	})

	t.Run("LeaderElectionIsPreserved", func(t *testing.T) {
		if got := (&retryingRunnable{Runnable: &fakeRunnable{}}).NeedLeaderElection(); !got {
			t.Error("a runnable without a preference must require leader election")
		}
		for _, need := range []bool{true, false} {
			r := &retryingRunnable{Runnable: &fakeController{leaderElection: need}}
			if got := r.NeedLeaderElection(); got != need {
				t.Errorf("NeedLeaderElection: want %v, got %v", need, got)
			}
		}
	})
}

func TestAddWrapsMRStateRecorder(t *testing.T) {
	m := &recordingManager{}
	tm := WrapManager(m, logging.NewNopLogger())
	rec := statemetrics.NewMRStateRecorder(nil, logging.NewNopLogger(), nil, nil, time.Second)
	if err := tm.Add(rec); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.added.(*retryingRunnable); !ok {
		t.Errorf("want MR state recorder wrapped in *retryingRunnable, got %T", m.added)
	}
	other := &fakeRunnable{}
	if err := tm.Add(other); err != nil {
		t.Fatal(err)
	}
	if m.added != other {
		t.Errorf("want other runnables added unchanged, got %T", m.added)
	}
}

type recordingManager struct {
	manager.Manager
	added manager.Runnable
}

func (m *recordingManager) Add(r manager.Runnable) error {
	m.added = r
	return nil
}
