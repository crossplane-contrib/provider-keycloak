/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	bindings "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/authenticationflow/bindings"
	execution "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/authenticationflow/execution"
	executionconfig "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/authenticationflow/executionconfig"
	flow "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/authenticationflow/flow"
	subflow "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/authenticationflow/subflow"
	protocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/client/protocolmapper"
	rolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/client/rolemapper"
	defaultgroups "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/defaults/defaultgroups"
	roles "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/defaults/roles"
	group "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/group/group"
	memberships "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/group/memberships"
	permissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/group/permissions"
	rolesgroup "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/group/roles"
	attributeidentityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/attributeidentityprovidermapper"
	groupidentityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/groupidentityprovidermapper"
	identityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/identityprovidermapper"
	importeridentityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/importeridentityprovidermapper"
	kubernetesidentityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/kubernetesidentityprovider"
	oidcopenshiftv4identityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/oidcopenshiftv4identityprovider"
	providertokenexchangescopepermission "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/providertokenexchangescopepermission"
	roleidentityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/roleidentityprovidermapper"
	spiffeidentityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/spiffeidentityprovider"
	templateimporteridentityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/templateimporteridentityprovidermapper"
	toroleidentityprovidermapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/identityprovider/toroleidentityprovidermapper"
	custommapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/custommapper"
	fullnamemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/fullnamemapper"
	groupmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/groupmapper"
	hardcodedattributemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/hardcodedattributemapper"
	hardcodedgroupmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/hardcodedgroupmapper"
	hardcodedrolemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/hardcodedrolemapper"
	msadldsuseraccountcontrolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/msadldsuseraccountcontrolmapper"
	msaduseraccountcontrolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/msaduseraccountcontrolmapper"
	rolemapperldap "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/rolemapper"
	userattributemapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/userattributemapper"
	userfederation "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/ldap/userfederation"
	googleidentityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/oidc/googleidentityprovider"
	identityprovider "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/oidc/identityprovider"
	client "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/client"
	clientauthorizationpermission "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientauthorizationpermission"
	clientauthorizationresource "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientauthorizationresource"
	clientclientpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientclientpolicy"
	clientdefaultscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientdefaultscopes"
	clientgrouppolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientgrouppolicy"
	clientoptionalscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientoptionalscopes"
	clientpermissions "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientpermissions"
	clientregexpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientregexpolicy"
	clientrolepolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientrolepolicy"
	clientscope "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientscope"
	clientserviceaccountrealmrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientserviceaccountrealmrole"
	clientserviceaccountrole "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientserviceaccountrole"
	clientuserpolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidclient/clientuserpolicy"
	groupmembershipprotocolmapper "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/openidgroup/groupmembershipprotocolmapper"
	organization "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/organization/organization"
	providerconfig "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/providerconfig"
	clientpolicyprofile "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/clientpolicyprofile"
	clientpolicyprofilepolicy "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/clientpolicyprofilepolicy"
	defaultclientscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/defaultclientscopes"
	keystorersa "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/keystorersa"
	optionalclientscopes "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/optionalclientscopes"
	realm "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/realm"
	realmevents "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/realmevents"
	realmlocalization "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/realmlocalization"
	requiredaction "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/requiredaction"
	userprofile "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/realm/userprofile"
	role "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/role/role"
	identityprovidersaml "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/saml/identityprovider"
	clientsamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/samlclient/client"
	clientdefaultscopessamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/samlclient/clientdefaultscopes"
	clientscopesamlclient "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/samlclient/clientscope"
	groups "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/user/groups"
	permissionsuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/user/permissions"
	rolesuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/user/roles"
	user "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/user/user"
	userfederationuser "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/user/userfederation"
	workflow "github.com/crossplane-contrib/provider-keycloak/internal/controller/namespaced/workflow/workflow"
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
		attributeidentityprovidermapper.Setup,
		groupidentityprovidermapper.Setup,
		identityprovidermapper.Setup,
		importeridentityprovidermapper.Setup,
		kubernetesidentityprovider.Setup,
		oidcopenshiftv4identityprovider.Setup,
		providertokenexchangescopepermission.Setup,
		roleidentityprovidermapper.Setup,
		spiffeidentityprovider.Setup,
		templateimporteridentityprovidermapper.Setup,
		toroleidentityprovidermapper.Setup,
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
		protocolmapper.SetupGated,
		rolemapper.SetupGated,
		defaultgroups.SetupGated,
		roles.SetupGated,
		group.SetupGated,
		memberships.SetupGated,
		permissions.SetupGated,
		rolesgroup.SetupGated,
		attributeidentityprovidermapper.SetupGated,
		groupidentityprovidermapper.SetupGated,
		identityprovidermapper.SetupGated,
		importeridentityprovidermapper.SetupGated,
		kubernetesidentityprovider.SetupGated,
		oidcopenshiftv4identityprovider.SetupGated,
		providertokenexchangescopepermission.SetupGated,
		roleidentityprovidermapper.SetupGated,
		spiffeidentityprovider.SetupGated,
		templateimporteridentityprovidermapper.SetupGated,
		toroleidentityprovidermapper.SetupGated,
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
		protocolmapper.SetupWebhookWithManager,
		rolemapper.SetupWebhookWithManager,
		defaultgroups.SetupWebhookWithManager,
		roles.SetupWebhookWithManager,
		group.SetupWebhookWithManager,
		memberships.SetupWebhookWithManager,
		permissions.SetupWebhookWithManager,
		rolesgroup.SetupWebhookWithManager,
		attributeidentityprovidermapper.SetupWebhookWithManager,
		groupidentityprovidermapper.SetupWebhookWithManager,
		identityprovidermapper.SetupWebhookWithManager,
		importeridentityprovidermapper.SetupWebhookWithManager,
		kubernetesidentityprovider.SetupWebhookWithManager,
		oidcopenshiftv4identityprovider.SetupWebhookWithManager,
		providertokenexchangescopepermission.SetupWebhookWithManager,
		roleidentityprovidermapper.SetupWebhookWithManager,
		spiffeidentityprovider.SetupWebhookWithManager,
		templateimporteridentityprovidermapper.SetupWebhookWithManager,
		toroleidentityprovidermapper.SetupWebhookWithManager,
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
