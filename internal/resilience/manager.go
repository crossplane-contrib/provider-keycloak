/*
Copyright 2022 Upbound Inc.
*/

// Package resilience keeps a single failing controller from taking down the
// whole provider process.
package resilience

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// WrapManager returns a manager whose Add isolates controller runnables: a
// controller whose Start fails (e.g. its informer cache never syncs because
// the stored objects of its kind cannot be listed or converted) is logged and
// dropped instead of aborting the manager, which would also tear down the CRD
// conversion webhook served by the same process and turn one broken kind into
// an outage for all of them (crossplane-contrib/provider-keycloak#669).
func WrapManager(mgr manager.Manager, log logging.Logger) manager.Manager {
	return &tolerantManager{Manager: mgr, log: log}
}

type tolerantManager struct {
	manager.Manager
	log logging.Logger
}

func (m *tolerantManager) Add(r manager.Runnable) error {
	if c, ok := r.(controller.Controller); ok {
		return m.Manager.Add(&tolerantController{Controller: c, log: m.log})
	}
	return m.Manager.Add(r)
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
