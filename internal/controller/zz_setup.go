// SPDX-FileCopyrightText: 2023 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	protocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/client/protocolmapper"
	rolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/client/rolemapper"
	group "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/group"
	memberships "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/memberships"
	roles "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/roles"
	client "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/client"
	clientdefaultscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientdefaultscopes"
	clientscope "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientscope"
	clientserviceaccountrealmrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientserviceaccountrealmrole"
	clientserviceaccountrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientserviceaccountrole"
	groupmembershipprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidgroup/groupmembershipprotocolmapper"
	providerconfig "github.com/crossplane-contrib/provider-keycloak/internal/controller/providerconfig"
	realm "github.com/crossplane-contrib/provider-keycloak/internal/controller/realm/realm"
	requiredaction "github.com/crossplane-contrib/provider-keycloak/internal/controller/realm/requiredaction"
	role "github.com/crossplane-contrib/provider-keycloak/internal/controller/role/role"
	groups "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/groups"
	user "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/user"
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
		clientdefaultscopes.Setup,
		clientscope.Setup,
		clientserviceaccountrealmrole.Setup,
		clientserviceaccountrole.Setup,
		groupmembershipprotocolmapper.Setup,
		providerconfig.Setup,
		realm.Setup,
		requiredaction.Setup,
		role.Setup,
		groups.Setup,
		user.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
