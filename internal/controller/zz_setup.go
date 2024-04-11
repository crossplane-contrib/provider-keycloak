/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	protocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/client/protocolmapper"
	rolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/client/rolemapper"
	roles "github.com/crossplane-contrib/provider-keycloak/internal/controller/defaults/roles"
	group "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/group"
	memberships "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/memberships"
	permissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/permissions"
	rolesgroup "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/roles"
	identityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/identityprovider/identityprovidermapper"
	identityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/oidc/identityprovider"
	client "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/client"
	clientclientpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientclientpolicy"
	clientdefaultscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientdefaultscopes"
	clientgrouppolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientgrouppolicy"
	clientpermissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientpermissions"
	clientrolepolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientrolepolicy"
	clientscope "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientscope"
	clientserviceaccountrealmrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientserviceaccountrealmrole"
	clientserviceaccountrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientserviceaccountrole"
	clientuserpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidclient/clientuserpolicy"
	groupmembershipprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/openidgroup/groupmembershipprotocolmapper"
	providerconfig "github.com/crossplane-contrib/provider-keycloak/internal/controller/providerconfig"
	keystorersa "github.com/crossplane-contrib/provider-keycloak/internal/controller/realm/keystorersa"
	realm "github.com/crossplane-contrib/provider-keycloak/internal/controller/realm/realm"
	requiredaction "github.com/crossplane-contrib/provider-keycloak/internal/controller/realm/requiredaction"
	role "github.com/crossplane-contrib/provider-keycloak/internal/controller/role/role"
	identityprovidersaml "github.com/crossplane-contrib/provider-keycloak/internal/controller/saml/identityprovider"
	clientsamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/samlclient/client"
	clientdefaultscopessamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/samlclient/clientdefaultscopes"
	clientscopesamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/samlclient/clientscope"
	groups "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/groups"
	permissionsuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/permissions"
	user "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/user"
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
		permissions.Setup,
		rolesgroup.Setup,
		identityprovidermapper.Setup,
		identityprovider.Setup,
		client.Setup,
		clientclientpolicy.Setup,
		clientdefaultscopes.Setup,
		clientgrouppolicy.Setup,
		clientpermissions.Setup,
		clientrolepolicy.Setup,
		clientscope.Setup,
		clientserviceaccountrealmrole.Setup,
		clientserviceaccountrole.Setup,
		clientuserpolicy.Setup,
		groupmembershipprotocolmapper.Setup,
		providerconfig.Setup,
		keystorersa.Setup,
		realm.Setup,
		requiredaction.Setup,
		role.Setup,
		identityprovidersaml.Setup,
		clientsamlclient.Setup,
		clientdefaultscopessamlclient.Setup,
		clientscopesamlclient.Setup,
		groups.Setup,
		permissionsuser.Setup,
		user.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
