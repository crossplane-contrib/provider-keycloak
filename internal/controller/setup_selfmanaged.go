/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	client "github.com/crossplane-contrib/provider-keycloak/internal/controller/generic"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func SetupSelfmanaged(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		client.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
