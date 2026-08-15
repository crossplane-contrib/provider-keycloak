/*
Copyright 2022 Upbound Inc.
*/

// Package connectionsecrettransform provides a no-op stub of the cluster-scope
// connection-secret-transform controller for the namespaced provider variant.
// The full implementation lives in
// internal/controller/cluster/connectionsecrettransform; it operates on
// cluster-scoped ConnectionSecretTransform CRDs and cluster-scoped managed
// resources, so it is not applicable to the namespaced provider variant.
package connectionsecrettransform

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"
)

// Setup is a no-op for the namespaced provider variant.
func Setup(_ ctrl.Manager, _ controller.Options) error {
	return nil
}

// SetupGated is a no-op for the namespaced provider variant.
func SetupGated(_ ctrl.Manager, _ controller.Options) error {
	return nil
}
