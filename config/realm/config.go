package realm

import (
	"context"
	"fmt"
	"time"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
)

// Group is the short group name for the resources in this package
var Group = "realm"

// realmDurationFields lists all Terraform schema field names on the
// keycloak_realm resource that represent durations. Values must be valid
// Go duration strings (e.g. "300s", "5m", "1h30m") because the Terraform
// provider converts them to integer seconds before sending to the Keycloak API.
var realmDurationFields = []string{
	"sso_session_idle_timeout",
	"sso_session_idle_timeout_remember_me",
	"sso_session_max_lifespan",
	"sso_session_max_lifespan_remember_me",
	"offline_session_idle_timeout",
	"offline_session_max_lifespan",
	"client_session_idle_timeout",
	"client_session_max_lifespan",
	"access_token_lifespan",
	"access_token_lifespan_for_implicit_flow",
	"access_code_lifespan",
	"access_code_lifespan_login",
	"access_code_lifespan_user_action",
	"action_token_generated_by_user_lifespan",
	"action_token_generated_by_admin_lifespan",
	"oauth2_device_code_lifespan",
}

// validateDurationString validates that a string value is a valid Go duration
// (parseable by time.ParseDuration). Empty strings are allowed because the
// fields are optional/computed.
func validateDurationString(v interface{}, k string) (warnings []string, errors []error) {
	value, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}
	if value == "" {
		return warnings, errors
	}
	if _, err := time.ParseDuration(value); err != nil {
		errors = append(errors, fmt.Errorf("%q is not a valid duration string for %q: %w (valid examples: \"300s\", \"5m\", \"1h30m\")", value, k, err))
	}
	return warnings, errors
}

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_realm", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = Group

		// Add validation to duration fields to reject invalid values early,
		// preventing broken values from reaching the Keycloak API.
		// See: https://github.com/crossplane-contrib/provider-keycloak/issues/434
		for _, field := range realmDurationFields {
			if s, ok := r.TerraformResource.Schema[field]; ok {
				s.ValidateFunc = validateDurationString
			}
		}
	})

	p.AddResourceConfigurator("keycloak_required_action", func(r *config.Resource) {
		r.ShortGroup = Group
		r.Kind = "RequiredAction"
	})

	p.AddResourceConfigurator("keycloak_realm_keystore_rsa", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = Group
		if s, ok := r.TerraformResource.Schema["private_key"]; ok {
			s.Sensitive = true
		}
		if s, ok := r.TerraformResource.Schema["certificate"]; ok {
			s.Sensitive = true
		}
	})

	p.AddResourceConfigurator("keycloak_realm_user_profile", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_realm_localization", func(r *config.Resource) {
		r.ShortGroup = Group
		r.Kind = "RealmLocalization"
	})

	p.AddResourceConfigurator("keycloak_realm_default_client_scopes", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_realm_optional_client_scopes", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_realm_events", func(r *config.Resource) {
		r.ShortGroup = Group
		r.Kind = "RealmEvents"
	})

	p.AddResourceConfigurator("keycloak_realm_client_policy_profile", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_realm_client_policy_profile_policy", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["profiles"] = config.Reference{
			TerraformName: "keycloak_realm_client_policy_profile",
			Extractor:     common.PathNameExtractor,
		}
	})
}

var realmIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm"},
	GetIDByExternalName:          getRealmIDByExternalName,
	GetIDByIdentifyingProperties: getRealmIDByIdentifyingProperties,
}

// RealmIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var RealmIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(realmIdentifyingPropertiesLookup)

func getRealmIDByExternalName(ctx context.Context, _ string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	return getRealmIDByIdentifyingProperties(ctx, parameters, kcClient)
}

func getRealmIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetRealm(ctx, parameters["realm"].(string))
	if err != nil {
		return "", err
	}
	return found.Realm, nil
}

var keystoreRsaIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name", "provider_id"},
	GetIDByExternalName:          getKeystoreRsaIDByExternalName,
	GetIDByIdentifyingProperties: getKeystoreRsaIDByIdentifyingProperties,
}

// KeystoreRsaIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var KeystoreRsaIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(keystoreRsaIdentifyingPropertiesLookup)

func getKeystoreRsaIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetRealmKeystoreRsa(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getKeystoreRsaIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := "org.keycloak.keys.KeyProvider"
	providerId := parameters["provider_id"].(string)
	name := parameters["name"].(string)
	realmId := parameters["realm_id"].(string)

	return lookup.GetComponentId(kcClient, ctx, realmId, &typ, nil, &providerId, &name)
}

var eventsRealmIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id"},
	GetIDByExternalName:          getEventsRealmIDByExternalName,
	GetIDByIdentifyingProperties: getEventsRealmIDByIdentifyingProperties,
}

// EventsRealmIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var EventsRealmIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(eventsRealmIdentifyingPropertiesLookup)

func getEventsRealmIDByExternalName(ctx context.Context, _ string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	return getEventsRealmIDByIdentifyingProperties(ctx, parameters, kcClient)
}

func getEventsRealmIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetRealm(ctx, parameters["realm_id"].(string))
	if err != nil {
		return "", err
	}
	return found.Realm, nil
}
