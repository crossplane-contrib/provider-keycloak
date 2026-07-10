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
	genericclientprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/client/genericclientprotocolmapper"
	genericclientrolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/client/genericclientrolemapper"
	protocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/client/protocolmapper"
	rolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/client/rolemapper"
	defaultgroups "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/defaults/defaultgroups"
	roles "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/defaults/roles"
	group "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/group/group"
	memberships "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/group/memberships"
	permissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/group/permissions"
	rolesgroup "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/group/roles"
	identityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/identityprovider/identityprovidermapper"
	kubernetesidentityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/identityprovider/kubernetesidentityprovider"
	oidcopenshiftv4identityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/identityprovider/oidcopenshiftv4identityprovider"
	providertokenexchangescopepermission "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/identityprovider/providertokenexchangescopepermission"
	spiffeidentityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/identityprovider/spiffeidentityprovider"
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
	clientauthorizationpermission "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientauthorizationpermission"
	clientauthorizationresource "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientauthorizationresource"
	clientclientpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientclientpolicy"
	clientdefaultscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientdefaultscopes"
	clientgrouppolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientgrouppolicy"
	clientoptionalscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientoptionalscopes"
	clientpermissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientpermissions"
	clientregexpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientregexpolicy"
	clientrolepolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientrolepolicy"
	clientscope "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientscope"
	clientserviceaccountrealmrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientserviceaccountrealmrole"
	clientserviceaccountrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientserviceaccountrole"
	clientuserpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidclient/clientuserpolicy"
	groupmembershipprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/openidgroup/groupmembershipprotocolmapper"
	organization "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/organization/organization"
	providerconfig "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/providerconfig"
	clientpolicyprofile "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/clientpolicyprofile"
	clientpolicyprofilepolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/clientpolicyprofilepolicy"
	defaultclientscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/defaultclientscopes"
	keystorersa "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/keystorersa"
	optionalclientscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/optionalclientscopes"
	realm "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/realm"
	realmevents "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/realmevents"
	realmlocalization "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/realmlocalization"
	requiredaction "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/requiredaction"
	userprofile "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/realm/userprofile"
	role "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/role/role"
	identityprovidersaml "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/saml/identityprovider"
	clientsamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/samlclient/client"
	clientdefaultscopessamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/samlclient/clientdefaultscopes"
	clientscopesamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/samlclient/clientscope"
	userattributeprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/samlclient/userattributeprotocolmapper"
	userpropertyprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/samlclient/userpropertyprotocolmapper"
	groups "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/groups"
	permissionsuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/permissions"
	rolesuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/roles"
	user "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/user"
	userfederationuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/user/userfederation"
	workflow "github.com/crossplane-contrib/provider-keycloak/internal/controller/cluster/workflow/workflow"
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
		genericclientprotocolmapper.Setup,
		genericclientrolemapper.Setup,
		protocolmapper.Setup,
		rolemapper.Setup,
		defaultgroups.Setup,
		roles.Setup,
		group.Setup,
		memberships.Setup,
		permissions.Setup,
		rolesgroup.Setup,
		identityprovidermapper.Setup,
		kubernetesidentityprovider.Setup,
		oidcopenshiftv4identityprovider.Setup,
		providertokenexchangescopepermission.Setup,
		spiffeidentityprovider.Setup,
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
		clientauthorizationpermission.Setup,
		clientauthorizationresource.Setup,
		clientclientpolicy.Setup,
		clientdefaultscopes.Setup,
		clientgrouppolicy.Setup,
		clientoptionalscopes.Setup,
		clientpermissions.Setup,
		clientregexpolicy.Setup,
		clientrolepolicy.Setup,
		clientscope.Setup,
		clientserviceaccountrealmrole.Setup,
		clientserviceaccountrole.Setup,
		clientuserpolicy.Setup,
		groupmembershipprotocolmapper.Setup,
		organization.Setup,
		providerconfig.Setup,
		clientpolicyprofile.Setup,
		clientpolicyprofilepolicy.Setup,
		defaultclientscopes.Setup,
		keystorersa.Setup,
		optionalclientscopes.Setup,
		realm.Setup,
		realmevents.Setup,
		realmlocalization.Setup,
		requiredaction.Setup,
		userprofile.Setup,
		role.Setup,
		identityprovidersaml.Setup,
		clientsamlclient.Setup,
		clientdefaultscopessamlclient.Setup,
		clientscopesamlclient.Setup,
		userattributeprotocolmapper.Setup,
		userpropertyprotocolmapper.Setup,
		groups.Setup,
		permissionsuser.Setup,
		rolesuser.Setup,
		user.Setup,
		userfederationuser.Setup,
		workflow.Setup,
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
		genericclientprotocolmapper.SetupGated,
		genericclientrolemapper.SetupGated,
		protocolmapper.SetupGated,
		rolemapper.SetupGated,
		defaultgroups.SetupGated,
		roles.SetupGated,
		group.SetupGated,
		memberships.SetupGated,
		permissions.SetupGated,
		rolesgroup.SetupGated,
		identityprovidermapper.SetupGated,
		kubernetesidentityprovider.SetupGated,
		oidcopenshiftv4identityprovider.SetupGated,
		providertokenexchangescopepermission.SetupGated,
		spiffeidentityprovider.SetupGated,
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
		clientauthorizationpermission.SetupGated,
		clientauthorizationresource.SetupGated,
		clientclientpolicy.SetupGated,
		clientdefaultscopes.SetupGated,
		clientgrouppolicy.SetupGated,
		clientoptionalscopes.SetupGated,
		clientpermissions.SetupGated,
		clientregexpolicy.SetupGated,
		clientrolepolicy.SetupGated,
		clientscope.SetupGated,
		clientserviceaccountrealmrole.SetupGated,
		clientserviceaccountrole.SetupGated,
		clientuserpolicy.SetupGated,
		groupmembershipprotocolmapper.SetupGated,
		organization.SetupGated,
		providerconfig.SetupGated,
		clientpolicyprofile.SetupGated,
		clientpolicyprofilepolicy.SetupGated,
		defaultclientscopes.SetupGated,
		keystorersa.SetupGated,
		optionalclientscopes.SetupGated,
		realm.SetupGated,
		realmevents.SetupGated,
		realmlocalization.SetupGated,
		requiredaction.SetupGated,
		userprofile.SetupGated,
		role.SetupGated,
		identityprovidersaml.SetupGated,
		clientsamlclient.SetupGated,
		clientdefaultscopessamlclient.SetupGated,
		clientscopesamlclient.SetupGated,
		userattributeprotocolmapper.SetupGated,
		userpropertyprotocolmapper.SetupGated,
		groups.SetupGated,
		permissionsuser.SetupGated,
		rolesuser.SetupGated,
		user.SetupGated,
		userfederationuser.SetupGated,
		workflow.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		bindings.SetupWebhookWithManager,
		execution.SetupWebhookWithManager,
		executionconfig.SetupWebhookWithManager,
		flow.SetupWebhookWithManager,
		subflow.SetupWebhookWithManager,
		genericclientprotocolmapper.SetupWebhookWithManager,
		genericclientrolemapper.SetupWebhookWithManager,
		protocolmapper.SetupWebhookWithManager,
		rolemapper.SetupWebhookWithManager,
		defaultgroups.SetupWebhookWithManager,
		roles.SetupWebhookWithManager,
		group.SetupWebhookWithManager,
		memberships.SetupWebhookWithManager,
		permissions.SetupWebhookWithManager,
		rolesgroup.SetupWebhookWithManager,
		identityprovidermapper.SetupWebhookWithManager,
		kubernetesidentityprovider.SetupWebhookWithManager,
		oidcopenshiftv4identityprovider.SetupWebhookWithManager,
		providertokenexchangescopepermission.SetupWebhookWithManager,
		spiffeidentityprovider.SetupWebhookWithManager,
		custommapper.SetupWebhookWithManager,
		fullnamemapper.SetupWebhookWithManager,
		groupmapper.SetupWebhookWithManager,
		hardcodedattributemapper.SetupWebhookWithManager,
		hardcodedgroupmapper.SetupWebhookWithManager,
		hardcodedrolemapper.SetupWebhookWithManager,
		msadldsuseraccountcontrolmapper.SetupWebhookWithManager,
		msaduseraccountcontrolmapper.SetupWebhookWithManager,
		rolemapperldap.SetupWebhookWithManager,
		userattributemapper.SetupWebhookWithManager,
		userfederation.SetupWebhookWithManager,
		googleidentityprovider.SetupWebhookWithManager,
		identityprovider.SetupWebhookWithManager,
		client.SetupWebhookWithManager,
		clientauthorizationpermission.SetupWebhookWithManager,
		clientauthorizationresource.SetupWebhookWithManager,
		clientclientpolicy.SetupWebhookWithManager,
		clientdefaultscopes.SetupWebhookWithManager,
		clientgrouppolicy.SetupWebhookWithManager,
		clientoptionalscopes.SetupWebhookWithManager,
		clientpermissions.SetupWebhookWithManager,
		clientregexpolicy.SetupWebhookWithManager,
		clientrolepolicy.SetupWebhookWithManager,
		clientscope.SetupWebhookWithManager,
		clientserviceaccountrealmrole.SetupWebhookWithManager,
		clientserviceaccountrole.SetupWebhookWithManager,
		clientuserpolicy.SetupWebhookWithManager,
		groupmembershipprotocolmapper.SetupWebhookWithManager,
		organization.SetupWebhookWithManager,
		providerconfig.SetupWebhookWithManager,
		clientpolicyprofile.SetupWebhookWithManager,
		clientpolicyprofilepolicy.SetupWebhookWithManager,
		defaultclientscopes.SetupWebhookWithManager,
		keystorersa.SetupWebhookWithManager,
		optionalclientscopes.SetupWebhookWithManager,
		realm.SetupWebhookWithManager,
		realmevents.SetupWebhookWithManager,
		realmlocalization.SetupWebhookWithManager,
		requiredaction.SetupWebhookWithManager,
		userprofile.SetupWebhookWithManager,
		role.SetupWebhookWithManager,
		identityprovidersaml.SetupWebhookWithManager,
		clientsamlclient.SetupWebhookWithManager,
		clientdefaultscopessamlclient.SetupWebhookWithManager,
		clientscopesamlclient.SetupWebhookWithManager,
		userattributeprotocolmapper.SetupWebhookWithManager,
		userpropertyprotocolmapper.SetupWebhookWithManager,
		groups.SetupWebhookWithManager,
		permissionsuser.SetupWebhookWithManager,
		rolesuser.SetupWebhookWithManager,
		user.SetupWebhookWithManager,
		userfederationuser.SetupWebhookWithManager,
		workflow.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
