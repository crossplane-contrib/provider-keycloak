/*
Copyright 2022 Upbound Inc.
*/

package config

import (
	"github.com/crossplane-contrib/provider-keycloak/config/authentication"
	"github.com/crossplane-contrib/provider-keycloak/config/group"
	"github.com/crossplane-contrib/provider-keycloak/config/identityprovider"
	"github.com/crossplane-contrib/provider-keycloak/config/ldap"
	"github.com/crossplane-contrib/provider-keycloak/config/mapper"
	"github.com/crossplane-contrib/provider-keycloak/config/oidc"
	"github.com/crossplane-contrib/provider-keycloak/config/openidclient"
	"github.com/crossplane-contrib/provider-keycloak/config/openidgroup"
	"github.com/crossplane-contrib/provider-keycloak/config/realm"
	"github.com/crossplane-contrib/provider-keycloak/config/role"
	"github.com/crossplane-contrib/provider-keycloak/config/saml"
	"github.com/crossplane-contrib/provider-keycloak/config/samlclient"
	"github.com/crossplane-contrib/provider-keycloak/config/user"
	"github.com/crossplane/upjet/v2/pkg/config"
)

// ExternalNameConfigs contains all external name configurations for this
// provider.
var ExternalNameConfigs = map[string]config.ExternalName{
	// Import requires using a randomly generated ID from provider: nl-2e21sda
	"keycloak_generic_protocol_mapper":                   mapper.ProtocolMapperIdentifierFromIdentifyingProperties,                // {UUid}
	"keycloak_generic_role_mapper":                       config.IdentifierFromProvider,                                           // {realm}/client|client-scope/{Client.UUid}/scope-mappings/{Client.UUid}/{Group.UUid}
	"keycloak_group_memberships":                         config.IdentifierFromProvider,                                           // {realm}/group-memberships/{Group.UUid}
	"keycloak_group_permissions":                         config.IdentifierFromProvider,                                           // {realm}/{Group.UUid}
	"keycloak_group_roles":                               config.IdentifierFromProvider,                                           // {realm}/{Group.UUid}
	"keycloak_group":                                     group.GroupIdentifierFromIdentifyingProperties,                          // {UUid}
	"keycloak_openid_client_client_policy":               openidclient.AuthzClientPoliciesIdentifierFromIdentifyingProperties,     // {UUid}
	"keycloak_openid_client_group_policy":                openidclient.AuthzGroupPoliciesIdentifierFromIdentifyingProperties,      // {UUid}
	"keycloak_openid_client_permissions":                 config.IdentifierFromProvider,                                           // {realm}/{Client.UUid}
	"keycloak_openid_client_role_policy":                 openidclient.AuthzRolePoliciesIdentifierFromIdentifyingProperties,       // {UUid}
	"keycloak_openid_client_user_policy":                 openidclient.AuthzUserPoliciesIdentifierFromIdentifyingProperties,       // {UUid}
	"keycloak_openid_client_default_scopes":              config.IdentifierFromProvider,                                           // {realm}/{Client.UUid}
	"keycloak_openid_client_optional_scopes":             config.IdentifierFromProvider,                                           // {realm}/{Client.UUid}
	"keycloak_openid_client_scope":                       openidclient.ClientScopeIdentifierFromIdentifyingProperties,             // {UUid}
	"keycloak_openid_client":                             openidclient.ClientIdentifierFromIdentifyingProperties,                  // {UUid}
	"keycloak_openid_group_membership_protocol_mapper":   openidgroup.IdentifierFromIdentifyingProperties,                         // {UUid}
	"keycloak_openid_client_service_account_realm_role":  config.IdentifierFromProvider,                                           // {serviceAccountUserId.UUid}/{role.UUid}
	"keycloak_openid_client_service_account_role":        config.IdentifierFromProvider,                                           // {serviceAccountUserId.UUid}/{role.UUid}
	"keycloak_organization":                              config.IdentifierFromProvider,                                           // {UUid}
	"keycloak_realm":                                     realm.RealmIdentifierFromIdentifyingProperties,                          // {realm}
	"keycloak_required_action":                           config.IdentifierFromProvider,                                           // {realm}/{alias}
	"keycloak_role":                                      role.IdentifierFromIdentifyingProperties,                                // {UUid}
	"keycloak_user_groups":                               config.IdentifierFromProvider,                                           // {realm}/{User.UUid}
	"keycloak_user_roles":                                config.IdentifierFromProvider,                                           // {realm}/{User.UUid}
	"keycloak_users_permissions":                         config.IdentifierFromProvider,                                           // {realm}
	"keycloak_user":                                      user.UserIdentifierFromIdentifyingProperties,                            // {UUid}
	"keycloak_oidc_identity_provider":                    oidc.IdentifierFromIdentifyingProperties,                                // {alias}
	"keycloak_oidc_google_identity_provider":             oidc.IdentifierFromIdentifyingProperties,                                // {alias}
	"keycloak_saml_identity_provider":                    saml.IdentifierFromIdentifyingProperties,                                // {alias}
	"keycloak_custom_identity_provider_mapper":           identityprovider.IdentifierFromIdentifyingProperties,                    // {UUid}
	"keycloak_saml_client":                               samlclient.ClientIdentifierFromIdentifyingProperties,                    // {UUid}
	"keycloak_saml_client_default_scopes":                config.IdentifierFromProvider,                                           // {realm}/{Client.UUid}
	"keycloak_saml_client_scope":                         samlclient.ClientScopeIdentifierFromIdentifyingProperties,               // {UUid}
	"keycloak_realm_keystore_rsa":                        realm.KeystoreRsaIdentifierFromIdentifyingProperties,                    // {UUid}
	"keycloak_realm_user_profile":                        config.IdentifierFromProvider,                                           // {realm}
	"keycloak_realm_default_client_scopes":               config.IdentifierFromProvider,                                           // {realm}
	"keycloak_realm_optional_client_scopes":              config.IdentifierFromProvider,                                           // {realm}
	"keycloak_realm_events":                              realm.EventsRealmIdentifierFromIdentifyingProperties,                    // {realm}
	"keycloak_authentication_flow":                       authentication.FlowIdentifierFromIdentifyingProperties,                  // {UUid}
	"keycloak_authentication_subflow":                    authentication.SubFlowIdentifierFromIdentifyingProperties,               // {UUid}
	"keycloak_authentication_execution":                  authentication.ExecutionIdentifierFromIdentifyingProperties,             // {UUid}
	"keycloak_authentication_execution_config":           authentication.ExecutionConfigIdentifierFromIdentifyingProperties,       // {UUid}
	"keycloak_authentication_bindings":                   config.IdentifierFromProvider,                                           // {realm}
	"keycloak_default_roles":                             config.IdentifierFromProvider,                                           // {UUid}
	"keycloak_default_groups":                            config.IdentifierFromProvider,                                           // {realm}/default-groups
	"keycloak_ldap_user_federation":                      ldap.UserFederationIdentifierFromIdentifyingProperties,                  // {UUid}
	"keycloak_ldap_user_attribute_mapper":                ldap.UserAttributeMapperIdentifierFromIdentifyingProperties,             // {UUid}
	"keycloak_ldap_role_mapper":                          ldap.RoleMapperIdentifierFromIdentifyingProperties,                      // {UUid}
	"keycloak_ldap_group_mapper":                         ldap.GroupMapperIdentifierFromIdentifyingProperties,                     // {UUid}
	"keycloak_ldap_hardcoded_role_mapper":                ldap.HardcodedRoleMapperIdentifierFromIdentifyingProperties,             // {UUid}
	"keycloak_ldap_hardcoded_group_mapper":               ldap.HardcodedGroupMapperIdentifierFromIdentifyingProperties,            // {UUid}
	"keycloak_ldap_msad_user_account_control_mapper":     ldap.MsadUserAccountControlMapperIdentifierFromIdentifyingProperties,    // {UUid}
	"keycloak_ldap_msad_lds_user_account_control_mapper": ldap.MsadLdsUserAccountControlMapperIdentifierFromIdentifyingProperties, // {UUid}
	"keycloak_ldap_hardcoded_attribute_mapper":           ldap.HardcodedAttributeMapperIdentifierFromIdentifyingProperties,        // {UUid}
	"keycloak_ldap_full_name_mapper":                     ldap.FullNameMapperIdentifierFromIdentifyingProperties,                  // {UUid}
	"keycloak_ldap_custom_mapper":                        ldap.CustomMapperIdentifierFromIdentifyingProperties,                    // {UUid}
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
