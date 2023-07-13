/*
Copyright 2021 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/upbound/upjet/pkg/controller"

	protocolmapper "github.com/corewire/provider-keycloak/internal/controller/client/protocolmapper"
	rolemapper "github.com/corewire/provider-keycloak/internal/controller/client/rolemapper"
	group "github.com/corewire/provider-keycloak/internal/controller/group/group"
	memberships "github.com/corewire/provider-keycloak/internal/controller/group/memberships"
	roles "github.com/corewire/provider-keycloak/internal/controller/group/roles"
	client "github.com/corewire/provider-keycloak/internal/controller/openidclient/client"
	providerconfig "github.com/corewire/provider-keycloak/internal/controller/providerconfig"
	realm "github.com/corewire/provider-keycloak/internal/controller/realm/realm"
	role "github.com/corewire/provider-keycloak/internal/controller/role/role"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		protocolmapper.Setup,
		rolemapper.Setup,
		group.Setup,
		memberships.Setup,
		roles.Setup,
		client.Setup,
		providerconfig.Setup,
		realm.Setup,
		role.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
