// SPDX-FileCopyrightText: 2023 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	protocolmapper "github.com/stakater/provider-keycloak/internal/controller/client/protocolmapper"
	rolemapper "github.com/stakater/provider-keycloak/internal/controller/client/rolemapper"
	roles "github.com/stakater/provider-keycloak/internal/controller/defaults/roles"
	group "github.com/stakater/provider-keycloak/internal/controller/group/group"
	memberships "github.com/stakater/provider-keycloak/internal/controller/group/memberships"
	rolesgroup "github.com/stakater/provider-keycloak/internal/controller/group/roles"
	identityprovider "github.com/stakater/provider-keycloak/internal/controller/oidc/identityprovider"
	client "github.com/stakater/provider-keycloak/internal/controller/openidclient/client"
	clientdefaultscopes "github.com/stakater/provider-keycloak/internal/controller/openidclient/clientdefaultscopes"
	clientscope "github.com/stakater/provider-keycloak/internal/controller/openidclient/clientscope"
	groupmembershipprotocolmapper "github.com/stakater/provider-keycloak/internal/controller/openidgroup/groupmembershipprotocolmapper"
	providerconfig "github.com/stakater/provider-keycloak/internal/controller/providerconfig"
	keystorersa "github.com/stakater/provider-keycloak/internal/controller/realm/keystorersa"
	realm "github.com/stakater/provider-keycloak/internal/controller/realm/realm"
	requiredaction "github.com/stakater/provider-keycloak/internal/controller/realm/requiredaction"
	role "github.com/stakater/provider-keycloak/internal/controller/role/role"
	identityprovidersaml "github.com/stakater/provider-keycloak/internal/controller/saml/identityprovider"
	groups "github.com/stakater/provider-keycloak/internal/controller/user/groups"
	user "github.com/stakater/provider-keycloak/internal/controller/user/user"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		protocolmapper.Setup,
		rolemapper.Setup,
		roles.Setup,
		group.Setup,
		memberships.Setup,
		rolesgroup.Setup,
		identityprovider.Setup,
		client.Setup,
		clientdefaultscopes.Setup,
		clientscope.Setup,
		groupmembershipprotocolmapper.Setup,
		providerconfig.Setup,
		keystorersa.Setup,
		realm.Setup,
		requiredaction.Setup,
		role.Setup,
		identityprovidersaml.Setup,
		groups.Setup,
		user.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
