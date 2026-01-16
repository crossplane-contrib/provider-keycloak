package ldap

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
)

// Group is the short group name for the resources in this package
var Group = "ldap"

const ldapStorageMapperType = "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {

	// ldap
	p.AddResourceConfigurator("keycloak_ldap_user_federation", func(r *config.Resource) {
		r.ShortGroup = Group
	})

	p.AddResourceConfigurator("keycloak_ldap_user_attribute_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_role_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
		r.References["client_id"] = config.Reference{
			TerraformName: "keycloak_openid_client",
			Extractor:     common.PathUUIDExtractor,
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_group_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_role_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
		r.References["role"] = config.Reference{
			TerraformName: "keycloak_role",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("name", false)`,
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_group_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
		r.References["group"] = config.Reference{
			TerraformName: "keycloak_group",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("name", false)`,
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_msad_user_account_control_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_msad_lds_user_account_control_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_hardcoded_attribute_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}

	})

	p.AddResourceConfigurator("keycloak_ldap_full_name_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
	})

	p.AddResourceConfigurator("keycloak_ldap_custom_mapper", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["ldap_user_federation_id"] = config.Reference{
			TerraformName: "keycloak_ldap_user_federation",
		}
	})
}

var userFederationIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "name"},
	GetIDByExternalName:          getUserFederationIDByExternalName,
	GetIDByIdentifyingProperties: getUserFederationIDByIdentifyingProperties,
}

// UserFederationIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var UserFederationIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(userFederationIdentifyingPropertiesLookup)

func getUserFederationIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapUserFederation(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getUserFederationIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := "org.keycloak.storage.UserStorageProvider"
	name := parameters["name"].(string)
	components, err := lookup.GetComponents(kcClient, ctx, parameters["realm_id"].(string), &typ, nil, &name)
	if err != nil {
		return "", err
	}
	filtered := lookup.Filter(components, func(component *lookup.Component) bool {
		return component.ProviderId == "ldap"
	})

	// Currently the Keycloak API allows to add multiple LdapProviders with the SAME name
	// If this is the case an error would be thrown here
	return lookup.SingleOrEmpty(filtered, func(component *lookup.Component) string {
		return component.Id
	})
}

var userAttributeMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getUserAttributeMapperIDByExternalName,
	GetIDByIdentifyingProperties: getUserAttributeMapperIDByIdentifyingProperties,
}

// UserAttributeMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var UserAttributeMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(userAttributeMapperIdentifyingPropertiesLookup)

func getUserAttributeMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapUserAttributeMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getUserAttributeMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	components, err := lookup.GetComponents(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &name)
	if err != nil {
		return "", err
	}

	filtered := lookup.Filter(components, func(component *lookup.Component) bool {
		return component.ProviderId == "user-attribute-ldap-mapper"
	})

	// Currently the Keycloak API allows to add multiple UserAttributeMapper with the SAME name
	// If this is the case an error would be thrown here
	return lookup.SingleOrEmpty(filtered, func(component *lookup.Component) string {
		return component.Id
	})
}

var roleMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getRoleMapperIDByExternalName,
	GetIDByIdentifyingProperties: getRoleMapperIDByIdentifyingProperties,
}

// RoleMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var RoleMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(roleMapperIdentifyingPropertiesLookup)

func getRoleMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapUserAttributeMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getRoleMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	providerId := "role-ldap-mapper"
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}

var groupMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getGroupMapperIDByExternalName,
	GetIDByIdentifyingProperties: getGroupMapperIDByIdentifyingProperties,
}

// GroupMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var GroupMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(groupMapperIdentifyingPropertiesLookup)

func getGroupMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapGroupMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getGroupMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	providerId := "group-ldap-mapper"
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}

var HardcodedRoleMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getHardcodedRoleMapperIDByExternalName,
	GetIDByIdentifyingProperties: getHardcodedRoleMapperIDByIdentifyingProperties,
}

// HardcodedRoleMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var HardcodedRoleMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(HardcodedRoleMapperIdentifyingPropertiesLookup)

func getHardcodedRoleMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapHardcodedRoleMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getHardcodedRoleMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	providerId := "hardcoded-ldap-role-mapper"
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}

var hardcodedGroupMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getHardcodedGroupMapperIDByExternalName,
	GetIDByIdentifyingProperties: getHardcodedGroupMapperIDByIdentifyingProperties,
}

// HardcodedGroupMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var HardcodedGroupMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(hardcodedGroupMapperIdentifyingPropertiesLookup)

func getHardcodedGroupMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapHardcodedGroupMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getHardcodedGroupMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	providerId := "hardcoded-ldap-group-mapper"
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}

var msadUserAccountControlMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getMsadUserAccountControlMapperIDByExternalName,
	GetIDByIdentifyingProperties: getMsadUserAccountControlMapperIDByIdentifyingProperties,
}

// MsadUserAccountControlMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var MsadUserAccountControlMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(msadUserAccountControlMapperIdentifyingPropertiesLookup)

func getMsadUserAccountControlMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapMsadUserAccountControlMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getMsadUserAccountControlMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	providerId := "msad-user-account-control-mapper"
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}

var msadLdsUserAccountControlMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getMsadLdsUserAccountControlMapperIDByExternalName,
	GetIDByIdentifyingProperties: getMsadLdsUserAccountControlMapperIDByIdentifyingProperties,
}

// MsadLdsUserAccountControlMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var MsadLdsUserAccountControlMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(msadLdsUserAccountControlMapperIdentifyingPropertiesLookup)

func getMsadLdsUserAccountControlMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapMsadLdsUserAccountControlMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getMsadLdsUserAccountControlMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	providerId := "msad-lds-user-account-control-mapper"
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}

var hardcodedAttributeMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getHardcodedAttributeMapperIDByExternalName,
	GetIDByIdentifyingProperties: getHardcodedAttributeMapperIDByIdentifyingProperties,
}

// HardcodedAttributeMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var HardcodedAttributeMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(hardcodedAttributeMapperIdentifyingPropertiesLookup)

func getHardcodedAttributeMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapHardcodedAttributeMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getHardcodedAttributeMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	providerId := "hardcoded-ldap-attribute-mapper"
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}

var fullNameMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "name"},
	GetIDByExternalName:          getFullNameMapperIDByExternalName,
	GetIDByIdentifyingProperties: getFullNameMapperIDByIdentifyingProperties,
}

// FullNameMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var FullNameMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(fullNameMapperIdentifyingPropertiesLookup)

func getFullNameMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapFullNameMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getFullNameMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := ldapStorageMapperType
	providerId := "full-name-ldap-mapper"
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}

var customMapperIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "ldap_user_federation_id", "provider_type", "provider_id", "name"},
	GetIDByExternalName:          getCustomMapperIDByExternalName,
	GetIDByIdentifyingProperties: getCustomMapperIDByIdentifyingProperties,
}

// CustomMapperIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var CustomMapperIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(customMapperIdentifyingPropertiesLookup)

func getCustomMapperIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetLdapCustomMapper(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getCustomMapperIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	typ := parameters["provider_type"].(string)
	providerId := parameters["provider_id"].(string)
	parentId := parameters["ldap_user_federation_id"].(string)
	name := parameters["name"].(string)

	return lookup.GetComponentId(kcClient, ctx, parameters["realm_id"].(string), &typ, &parentId, &providerId, &name)
}
