/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	bindings "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/authenticationflow/bindings"
	execution "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/authenticationflow/execution"
	executionconfig "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/authenticationflow/executionconfig"
	flow "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/authenticationflow/flow"
	subflow "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/authenticationflow/subflow"
	protocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/client/protocolmapper"
	rolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/client/rolemapper"
	defaultgroups "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/defaults/defaultgroups"
	roles "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/defaults/roles"
	group "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/group/group"
	memberships "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/group/memberships"
	permissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/group/permissions"
	rolesgroup "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/group/roles"
	identityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/identityprovider/identityprovidermapper"
	custommapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/custommapper"
	fullnamemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/fullnamemapper"
	groupmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/groupmapper"
	hardcodedattributemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/hardcodedattributemapper"
	hardcodedgroupmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/hardcodedgroupmapper"
	hardcodedrolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/hardcodedrolemapper"
	msadldsuseraccountcontrolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/msadldsuseraccountcontrolmapper"
	msaduseraccountcontrolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/msaduseraccountcontrolmapper"
	rolemapperldap "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/rolemapper"
	userattributemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/userattributemapper"
	userfederation "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/ldap/userfederation"
	googleidentityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/oidc/googleidentityprovider"
	identityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/oidc/identityprovider"
	client "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/client"
	clientclientpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientclientpolicy"
	clientdefaultscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientdefaultscopes"
	clientgrouppolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientgrouppolicy"
	clientoptionalscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientoptionalscopes"
	clientpermissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientpermissions"
	clientrolepolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientrolepolicy"
	clientscope "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientscope"
	clientserviceaccountrealmrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientserviceaccountrealmrole"
	clientserviceaccountrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientserviceaccountrole"
	clientuserpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientuserpolicy"
	groupmembershipprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidgroup/groupmembershipprotocolmapper"
	organization "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/organization/organization"
	providerconfig "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/providerconfig"
	defaultclientscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/defaultclientscopes"
	keystorersa "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/keystorersa"
	optionalclientscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/optionalclientscopes"
	realm "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/realm"
	realmevents "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/realmevents"
	requiredaction "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/requiredaction"
	userprofile "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/userprofile"
	role "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/role/role"
	identityprovidersaml "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/saml/identityprovider"
	clientsamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/samlclient/client"
	clientdefaultscopessamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/samlclient/clientdefaultscopes"
	clientscopesamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/samlclient/clientscope"
	groups "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/groups"
	permissionsuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/permissions"
	rolesuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/roles"
	user "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/user"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		bindings.Setup,
		execution.Setup,
		executionconfig.Setup,
		flow.Setup,
		subflow.Setup,
		protocolmapper.Setup,
		rolemapper.Setup,
		defaultgroups.Setup,
		roles.Setup,
		group.Setup,
		memberships.Setup,
		permissions.Setup,
		rolesgroup.Setup,
		identityprovidermapper.Setup,
		custommapper.Setup,
		fullnamemapper.Setup,
		groupmapper.Setup,
		hardcodedattributemapper.Setup,
		hardcodedgroupmapper.Setup,
		hardcodedrolemapper.Setup,
		msadldsuseraccountcontrolmapper.Setup,
		msaduseraccountcontrolmapper.Setup,
		rolemapperldap.Setup,
		userattributemapper.Setup,
		userfederation.Setup,
		googleidentityprovider.Setup,
		identityprovider.Setup,
		client.Setup,
		clientclientpolicy.Setup,
		clientdefaultscopes.Setup,
		clientgrouppolicy.Setup,
		clientoptionalscopes.Setup,
		clientpermissions.Setup,
		clientrolepolicy.Setup,
		clientscope.Setup,
		clientserviceaccountrealmrole.Setup,
		clientserviceaccountrole.Setup,
		clientuserpolicy.Setup,
		groupmembershipprotocolmapper.Setup,
		organization.Setup,
		providerconfig.Setup,
		defaultclientscopes.Setup,
		keystorersa.Setup,
		optionalclientscopes.Setup,
		realm.Setup,
		realmevents.Setup,
		requiredaction.Setup,
		userprofile.Setup,
		role.Setup,
		identityprovidersaml.Setup,
		clientsamlclient.Setup,
		clientdefaultscopessamlclient.Setup,
		clientscopesamlclient.Setup,
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

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		bindings.SetupGated,
		execution.SetupGated,
		executionconfig.SetupGated,
		flow.SetupGated,
		subflow.SetupGated,
		protocolmapper.SetupGated,
		rolemapper.SetupGated,
		defaultgroups.SetupGated,
		roles.SetupGated,
		group.SetupGated,
		memberships.SetupGated,
		permissions.SetupGated,
		rolesgroup.SetupGated,
		identityprovidermapper.SetupGated,
		custommapper.SetupGated,
		fullnamemapper.SetupGated,
		groupmapper.SetupGated,
		hardcodedattributemapper.SetupGated,
		hardcodedgroupmapper.SetupGated,
		hardcodedrolemapper.SetupGated,
		msadldsuseraccountcontrolmapper.SetupGated,
		msaduseraccountcontrolmapper.SetupGated,
		rolemapperldap.SetupGated,
		userattributemapper.SetupGated,
		userfederation.SetupGated,
		googleidentityprovider.SetupGated,
		identityprovider.SetupGated,
		client.SetupGated,
		clientclientpolicy.SetupGated,
		clientdefaultscopes.SetupGated,
		clientgrouppolicy.SetupGated,
		clientoptionalscopes.SetupGated,
		clientpermissions.SetupGated,
		clientrolepolicy.SetupGated,
		clientscope.SetupGated,
		clientserviceaccountrealmrole.SetupGated,
		clientserviceaccountrole.SetupGated,
		clientuserpolicy.SetupGated,
		groupmembershipprotocolmapper.SetupGated,
		organization.SetupGated,
		providerconfig.SetupGated,
		defaultclientscopes.SetupGated,
		keystorersa.SetupGated,
		optionalclientscopes.SetupGated,
		realm.SetupGated,
		realmevents.SetupGated,
		requiredaction.SetupGated,
		userprofile.SetupGated,
		role.SetupGated,
		identityprovidersaml.SetupGated,
		clientsamlclient.SetupGated,
		clientdefaultscopessamlclient.SetupGated,
		clientscopesamlclient.SetupGated,
		groups.SetupGated,
		permissionsuser.SetupGated,
		rolesuser.SetupGated,
		user.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
