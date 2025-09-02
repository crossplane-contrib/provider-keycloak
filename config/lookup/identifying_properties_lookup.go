package lookup

import (
	"context"
	"errors"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
	"strconv"
)

type IdentifyingPropertiesLookupConfig struct {
	GetIDByExternalName          GetIDByExternalName
	GetIDByIdentifyingProperties GetIDByIdentifyingProperties
	RequiredParameters           []string
	OptionalParameters           []string
}

func BuildIdentifyingPropertiesLookupIDFn(lookupConfig IdentifyingPropertiesLookupConfig) config.GetIDFn {
	return func(ctx context.Context, externalName string, parameters map[string]any, terraformProviderConfig map[string]any) (string, error) {
		return GetIDFromIdentifyingProperties(ctx, externalName, parameters, terraformProviderConfig, lookupConfig)
	}
}

// BuildIdentifyingPropertiesLookup creates the ExternalName which contains all information that is necessary for naming operations
// It will set this specialized GetIDFn: GetIDFromIdentifyingProperties and return the ID as ExternalName
func BuildIdentifyingPropertiesLookup(lookupConfig IdentifyingPropertiesLookupConfig) config.ExternalName {
	return config.ExternalName{
		SetIdentifierArgumentFn: config.NopSetIdentifierArgument,
		GetExternalNameFn:       config.IDAsExternalName,
		GetIDFn:                 BuildIdentifyingPropertiesLookupIDFn(lookupConfig),
		DisableNameInitializer:  true,
	}
}

type GetIDByExternalName func(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error)

type GetIDByIdentifyingProperties func(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error)

// GetIDFromIdentifyingProperties is a specialized GetIDFn
// Check if external-name is set and try to resolve the resource by external-name (using GetIDByExternalName)
// If resource can NOT be resolved by external-name or external-name is NOT set
// then try to resolve resource by identifying properties like realmId, clientId, etc. (using GetIDByIdentifyingProperties)
func GetIDFromIdentifyingProperties(ctx context.Context, externalName string, parameters map[string]any, terraformProviderConfig map[string]any, lookupConfig IdentifyingPropertiesLookupConfig) (string, error) {
	kcClient, err := newKeycloakClient(ctx, terraformProviderConfig)
	if err != nil {
		return "", err
	}

	processedParameters := make(map[string]any)

	for _, reqParamName := range lookupConfig.RequiredParameters {
		reqParam, reqParamExists := parameters[reqParamName]
		if !reqParamExists {
			return "", errors.New("required param '" + reqParamName + "' not set")
		}
		processedParameters[reqParamName] = reqParam
	}

	for _, optParamName := range lookupConfig.OptionalParameters {
		optParam, optParamExists := parameters[optParamName]
		if !optParamExists {
			optParam = ""
		}
		processedParameters[optParamName] = optParam
	}

	if externalName != "" {
		foundID, err := lookupConfig.GetIDByExternalName(ctx, externalName, processedParameters, kcClient)
		if err != nil {
			var apiErr *keycloak.ApiError
			if !(errors.As(err, &apiErr) && apiErr.Code == 404) {
				return "", err
			}
		} else {
			return foundID, nil
		}
	}

	foundID, err := lookupConfig.GetIDByIdentifyingProperties(ctx, processedParameters, kcClient)
	if err != nil {
		var apiErr *keycloak.ApiError
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return "", nil
		}

		return "", err
	}

	return foundID, nil
}

func SingleOrEmpty[T any](list []*T, idFunc func(obj *T) string) (string, error) {
	if len(list) == 0 {
		return "", nil
	}

	if len(list) > 1 {
		return "", errors.New("Too many resources found, which match the identifying parameters. Expected 0 or 1, but was " + strconv.Itoa(len(list)))
	}

	return idFunc(list[0]), nil
}

func Filter[T any](list []*T, filterFunc func(obj *T) bool) []*T {
	var filtered []*T

	for _, item := range list {
		if filterFunc(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
