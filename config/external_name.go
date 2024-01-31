/*
Copyright 2022 Upbound Inc.
*/

package config

import "github.com/crossplane/upjet/pkg/config"

// ExternalNameConfigs contains all external name configurations for this
// provider.
var ExternalNameConfigs = map[string]config.ExternalName{
	// Import requires using a randomly generated ID from provider: nl-2e21sda
	"keycloak_generic_protocol_mapper":                 config.IdentifierFromProvider,
	"keycloak_generic_role_mapper":                     config.IdentifierFromProvider,
	"keycloak_group_memberships":                       config.IdentifierFromProvider,
	"keycloak_group_roles":                             config.IdentifierFromProvider,
	"keycloak_group":                                   config.IdentifierFromProvider,
	"keycloak_openid_client_default_scopes":            config.IdentifierFromProvider,
	"keycloak_openid_client_scope":                     config.IdentifierFromProvider,
	"keycloak_openid_client":                           config.IdentifierFromProvider,
	"keycloak_openid_group_membership_protocol_mapper": config.IdentifierFromProvider,
	"keycloak_realm":                                   config.IdentifierFromProvider,
	"keycloak_required_action":                         config.IdentifierFromProvider,
	"keycloak_role":                                    config.IdentifierFromProvider,
	"keycloak_user_groups":                             config.IdentifierFromProvider,
	"keycloak_user":                                    config.IdentifierFromProvider,
	"keycloak_default_roles":                           config.TemplatedStringAsIdentifier("", "{{ .parameters.realm_id }}/{{ .parameters.id }}"),
	"keycloak_oidc_identity_provider":                  config.ParameterAsIdentifier("alias"),
	"keycloak_saml_identity_provider":                  config.ParameterAsIdentifier("alias"),
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
