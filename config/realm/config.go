package realm

import (
	"context"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane/upjet/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// Group is the short group name for the resources in this package
var Group = "realm"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_realm", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = Group
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
	return found.Id, nil
}

var keystoreRsaIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
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
	components, err := lookup.GetComponents(kcClient, ctx, parameters["realm_id"].(string), "org.keycloak.keys.KeyProvider", parameters["name"].(string))
	if err != nil {
		return "", err
	}

	// Currently the Keycloak API allows to add multiple KeyProvider with the SAME name
	// If this is the case an error would be thrown here
	return lookup.SingleOrEmpty(components, func(scope *lookup.Component) string {
		return scope.Id
	})
}
