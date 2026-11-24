/*
Copyright 2022 Upbound Inc.
*/

// Package resilience keeps a single failing controller from taking down the
// whole provider process.
package resilience

import (
	"context"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// recorderRetryInterval is the delay before a failed MR state recorder is
// restarted.
const recorderRetryInterval = 10 * time.Second

// WrapManager returns a manager whose Add isolates controller runnables: a
// controller whose Start fails (e.g. its informer cache never syncs because
// the stored objects of its kind cannot be listed or converted) is logged and
// dropped instead of aborting the manager, which would also tear down the CRD
// conversion webhook served by the same process and turn one broken kind into
// an outage for all of them (crossplane-contrib/provider-keycloak#669).
//
// MR state metric recorders are restarted when they fail instead of aborting
// the manager: right after a CRD becomes Established, API discovery may not yet
// serve its kind, so listing it fails with "no matches for kind" and would
// otherwise crashloop the provider on first start
// (crossplane-contrib/provider-keycloak#807).
func WrapManager(mgr manager.Manager, log logging.Logger) manager.Manager {
	return &tolerantManager{Manager: mgr, log: log}
}

type tolerantManager struct {
	manager.Manager
	log logging.Logger
}

func (m *tolerantManager) Add(r manager.Runnable) error {
	switch v := r.(type) {
	case controller.Controller:
		return m.Manager.Add(&tolerantController{Controller: v, log: m.log})
	case *statemetrics.MRStateRecorder:
		return m.Manager.Add(&retryingRunnable{Runnable: v, log: m.log, interval: recorderRetryInterval})
	}
	return m.Manager.Add(r)
}

// retryingRunnable restarts the wrapped runnable after a failure until the
// manager context is cancelled.
type retryingRunnable struct {
	manager.Runnable
	log      logging.Logger
	interval time.Duration
}

func (r *retryingRunnable) Start(ctx context.Context) error {
	for {
		err := r.Runnable.Start(ctx)
		if err == nil || ctx.Err() != nil {
			return err
		}
		if meta.IsNoMatchError(err) {
			r.log.Debug("Kind is not served yet, retrying", "error", err)
		} else {
			r.log.Info("Runnable failed, retrying", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(r.interval):
		}
	}
}

// NeedLeaderElection preserves the wrapped runnable's leader election
// requirement, which the manager sniffs on the runnable it is given.
func (r *retryingRunnable) NeedLeaderElection() bool {
	if le, ok := r.Runnable.(manager.LeaderElectionRunnable); ok {
		return le.NeedLeaderElection()
	}
	return true
}

type tolerantController struct {
	controller.Controller
	log logging.Logger
}

func (c *tolerantController) Start(ctx context.Context) error {
	err := c.Controller.Start(ctx)
	if err == nil || ctx.Err() != nil {
		return err
	}
	c.log.Info("Controller failed to start, the provider keeps running without it. Its managed resources are not reconciled until the underlying problem is fixed and the provider is restarted.", "error", err)
	return nil
}

// NeedLeaderElection preserves the wrapped controller's leader election
// requirement, which the manager sniffs on the runnable it is given.
func (c *tolerantController) NeedLeaderElection() bool {
	if le, ok := c.Controller.(manager.LeaderElectionRunnable); ok {
		return le.NeedLeaderElection()
	}
	return true
}
