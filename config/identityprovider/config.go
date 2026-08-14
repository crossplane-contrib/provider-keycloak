package identityprovider

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane-contrib/provider-keycloak/config/multitypes"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

const (
	// Group is the short group for this provider.
	Group = "identityprovider"
)

// additionalIdentityProviderTypes lists all identity provider resources besides
// keycloak_oidc_identity_provider that can be referenced by an alias field.
// The prefix is prepended to the original field name to build the synthetic field name,
// e.g. "saml" + "provider_alias" => "saml_provider_alias".
var additionalIdentityProviderTypes = []struct {
	prefix        string
	terraformName string
}{
	{"saml", "keycloak_saml_identity_provider"},
	{"google", "keycloak_oidc_google_identity_provider"},
	{"github", "keycloak_oidc_github_identity_provider"},
	{"facebook", "keycloak_oidc_facebook_identity_provider"},
	{"microsoft", "keycloak_oidc_microsoft_identity_provider"},
	{"kubernetes", "keycloak_kubernetes_identity_provider"},
	{"spiffe", "keycloak_spiffe_identity_provider"},
	{"openshift_v4", "keycloak_oidc_openshift_v4_identity_provider"},
}

// identityProviderAliasInstances builds the multitypes instances for a field holding an
// identity provider alias. Keycloak identifies identity providers by realm and alias only,
// so any identity provider type is a valid target.
//
// The first instance reuses the original field name and keeps referencing
// keycloak_oidc_identity_provider for backward compatibility, all other types get their
// own synthetic field.
func identityProviderAliasInstances(field string) []multitypes.Instance {
	instances := []multitypes.Instance{
		{
			Name: field,
			Reference: config.Reference{
				TerraformName: "keycloak_oidc_identity_provider",
				Extractor:     common.PathIdentityProviderAliasExtractor,
			},
		},
	}
	for _, t := range additionalIdentityProviderTypes {
		instances = append(instances, multitypes.Instance{
			Name: t.prefix + "_" + field,
			Reference: config.Reference{
				TerraformName: t.terraformName,
				Extractor:     common.PathIdentityProviderAliasExtractor,
			},
		})
	}
	return instances
}

// configureIdentityProviderAliasReferences wires the given alias field of the resource to all
// identity provider types, keeping the original field settable for backward compatibility.
func configureIdentityProviderAliasReferences(r *config.Resource, field string) {
	multitypes.ApplyToWithOptions(r, field,
		&multitypes.Options{KeepOriginalField: true},
		identityProviderAliasInstances(field)...,
	)
}

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_custom_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		configureIdentityProviderAliasReferences(r, "identity_provider_alias")
	})

	p.AddResourceConfigurator("keycloak_attribute_importer_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		configureIdentityProviderAliasReferences(r, "identity_provider_alias")
	})

	p.AddResourceConfigurator("keycloak_attribute_to_role_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		configureIdentityProviderAliasReferences(r, "identity_provider_alias")
	})

	p.AddResourceConfigurator("keycloak_hardcoded_attribute_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		configureIdentityProviderAliasReferences(r, "identity_provider_alias")
	})

	p.AddResourceConfigurator("keycloak_hardcoded_group_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		configureIdentityProviderAliasReferences(r, "identity_provider_alias")
	})

	p.AddResourceConfigurator("keycloak_hardcoded_role_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		configureIdentityProviderAliasReferences(r, "identity_provider_alias")
	})

	p.AddResourceConfigurator("keycloak_user_template_importer_identity_provider_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		configureIdentityProviderAliasReferences(r, "identity_provider_alias")
	})

	p.AddResourceConfigurator("keycloak_identity_provider_token_exchange_scope_permission", func(r *config.Resource) {
		r.ShortGroup = Group
		configureIdentityProviderAliasReferences(r, "provider_alias")
		r.References["clients"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_kubernetes_identity_provider", func(r *config.Resource) {
		r.ShortGroup = Group
		r.Kind = "KubernetesIdentityProvider"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		r.References["organization_id"] = config.Reference{
			TerraformName: "keycloak_organization",
		}
		r.References["first_broker_login_flow_alias"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor:     common.PathAuthenticationFlowAliasExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_spiffe_identity_provider", func(r *config.Resource) {
		r.ShortGroup = Group
		r.Kind = "SpiffeIdentityProvider"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		r.References["first_broker_login_flow_alias"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor:     common.PathAuthenticationFlowAliasExtractor,
		}
		r.References["post_broker_login_flow_alias"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor:     common.PathAuthenticationFlowAliasExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_oidc_openshift_v4_identity_provider", func(r *config.Resource) {
		r.ShortGroup = Group
		r.Kind = "OidcOpenShiftV4IdentityProvider"
		r.References["realm"] = config.Reference{
			TerraformName: "keycloak_realm",
		}
		r.References["first_broker_login_flow_alias"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor:     common.PathAuthenticationFlowAliasExtractor,
		}
		r.References["post_broker_login_flow_alias"] = config.Reference{
			TerraformName: "keycloak_authentication_flow",
			Extractor:     common.PathAuthenticationFlowAliasExtractor,
		}
	})
}

var identifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm", "identity_provider_alias", "name"},
	GetIDByExternalName:          getIDByExternalName,
	GetIDByIdentifyingProperties: getIDByIdentifyingProperties,
}

// IdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var IdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(identifyingPropertiesLookup)

func getIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetIdentityProviderMapper(ctx, parameters["realm"].(string), parameters["identity_provider_alias"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetIdentityProviderMappers(ctx, parameters["realm"].(string), parameters["identity_provider_alias"].(string))
	if err != nil {
		return "", err
	}

	filtered := lookup.Filter(found, func(mapper *keycloak.IdentityProviderMapper) bool {
		return mapper.Name == parameters["name"].(string)
	})

	return lookup.SingleOrEmpty(filtered, func(mapper *keycloak.IdentityProviderMapper) string {
		return mapper.Id
	})
}
