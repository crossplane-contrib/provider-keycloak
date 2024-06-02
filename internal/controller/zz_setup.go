/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	openidclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/client/openidclient"
	samlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/client/samlclient"
	defaultgroups "github.com/crossplane-contrib/provider-keycloak/internal/controller/defaults/defaultgroups"
	roles "github.com/crossplane-contrib/provider-keycloak/internal/controller/defaults/roles"
	group "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/group"
	memberships "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/memberships"
	permissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/permissions"
	rolesgroup "github.com/crossplane-contrib/provider-keycloak/internal/controller/group/roles"
	identityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/idp/identityprovider"
	identityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/idp/identityprovidermapper"
	openididentityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/idp/openididentityprovider"
	custommapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/custommapper"
	fullnamemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/fullnamemapper"
	groupmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/groupmapper"
	hardcodedattributemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/hardcodedattributemapper"
	hardcodedgroupmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/hardcodedgroupmapper"
	hardcodedrolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/hardcodedrolemapper"
	msadldsuseraccountcontrolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/msadldsuseraccountcontrolmapper"
	msaduseraccountcontrolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/msaduseraccountcontrolmapper"
	rolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/rolemapper"
	userattributemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/userattributemapper"
	userfederation "github.com/crossplane-contrib/provider-keycloak/internal/controller/ldap/userfederation"
	rolemappermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/mapper/rolemapper"
	samlprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/mapper/samlprotocolmapper"
	clientclientpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientclientpolicy"
	clientdefaultscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientdefaultscopes"
	clientgrouppolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientgrouppolicy"
	clientpermissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientpermissions"
	clientrolepolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientrolepolicy"
	clientscope "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientscope"
	clientserviceaccountrealmrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientserviceaccountrealmrole"
	clientserviceaccountrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientserviceaccountrole"
	clientuserpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/clientuserpolicy"
	groupmembershipprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/openid/groupmembershipprotocolmapper"
	providerconfig "github.com/crossplane-contrib/provider-keycloak/internal/controller/providerconfig"
	keystorersa "github.com/crossplane-contrib/provider-keycloak/internal/controller/realm/keystorersa"
	realm "github.com/crossplane-contrib/provider-keycloak/internal/controller/realm/realm"
	requiredaction "github.com/crossplane-contrib/provider-keycloak/internal/controller/realm/requiredaction"
	role "github.com/crossplane-contrib/provider-keycloak/internal/controller/role/role"
	clientdefaultscopessaml "github.com/crossplane-contrib/provider-keycloak/internal/controller/saml/clientdefaultscopes"
	clientscopesaml "github.com/crossplane-contrib/provider-keycloak/internal/controller/saml/clientscope"
	groups "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/groups"
	permissionsuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/permissions"
	rolesuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/roles"
	user "github.com/crossplane-contrib/provider-keycloak/internal/controller/user/user"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		openidclient.Setup,
		samlclient.Setup,
		defaultgroups.Setup,
		roles.Setup,
		group.Setup,
		memberships.Setup,
		permissions.Setup,
		rolesgroup.Setup,
		identityprovider.Setup,
		identityprovidermapper.Setup,
		openididentityprovider.Setup,
		custommapper.Setup,
		fullnamemapper.Setup,
		groupmapper.Setup,
		hardcodedattributemapper.Setup,
		hardcodedgroupmapper.Setup,
		hardcodedrolemapper.Setup,
		msadldsuseraccountcontrolmapper.Setup,
		msaduseraccountcontrolmapper.Setup,
		rolemapper.Setup,
		userattributemapper.Setup,
		userfederation.Setup,
		rolemappermapper.Setup,
		samlprotocolmapper.Setup,
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
		clientdefaultscopessaml.Setup,
		clientscopesaml.Setup,
		groups.Setup,
		permissionsuser.Setup,
		rolesuser.Setup,
		user.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
