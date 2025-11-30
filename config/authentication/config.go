package authentication

import (
	"context"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"

	"github.com/crossplane-contrib/provider-keycloak/config/common"
	"github.com/crossplane-contrib/provider-keycloak/config/lookup"
	"github.com/crossplane-contrib/provider-keycloak/config/multitypes"
)

const (
	// Group is the short group for this provider.
	Group = "authenticationflow"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("keycloak_authentication_flow", func(r *config.Resource) {
		r.ShortGroup = Group
	})
	p.AddResourceConfigurator("keycloak_authentication_subflow", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["parent_flow_alias"] = config.Reference{
			TerraformName:     "keycloak_authentication_flow",
			Extractor:         common.PathAuthenticationFlowAliasExtractor,
			RefFieldName:      "ParentFlowAliasRef",
			SelectorFieldName: "ParentFlowAliasSelector",
		}
	})
	p.AddResourceConfigurator("keycloak_authentication_execution", func(r *config.Resource) {
		r.ShortGroup = Group

		// Issue #163: parent_flow_alias can reference either a Flow or a Subflow in Keycloak.
		// For backward compatibility, we explicitly keep parentFlowAlias/Ref/Selector for Flow references.
		// We add parentSubflowAlias/Ref/Selector for the new Subflow reference capability.
		// The multitypes pattern ensures only one can be set, and both map to parent_flow_alias in Terraform.
		multitypes.ApplyToWithOptions(r, "parent_flow_alias",
			&multitypes.Options{KeepOriginalField: true}, // Explicit: maintain backward compatibility
			multitypes.Instance{
				// Use "parent_flow_alias" as the name so it generates "parentFlowAlias" in the API
				Name: "parent_flow_alias",
				Reference: config.Reference{
					TerraformName:     "keycloak_authentication_flow",
					Extractor:         common.PathAuthenticationFlowAliasExtractor,
					RefFieldName:      "ParentFlowAliasRef",
					SelectorFieldName: "ParentFlowAliasSelector",
				},
			},
			multitypes.Instance{
				Name: "parent_subflow_alias",
				Reference: config.Reference{
					TerraformName:     "keycloak_authentication_subflow",
					Extractor:         common.PathAuthenticationFlowAliasExtractor,
					RefFieldName:      "ParentSubflowAliasRef",
					SelectorFieldName: "ParentSubflowAliasSelector",
				},
			},
		)
	})
	p.AddResourceConfigurator("keycloak_authentication_execution_config", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["execution_id"] = config.Reference{
			TerraformName: "keycloak_authentication_execution",
		}
	})
	p.AddResourceConfigurator("keycloak_authentication_bindings", func(r *config.Resource) {
		r.ShortGroup = Group
		r.References["browser_flow"] = config.Reference{
			TerraformName:     "keycloak_authentication_flow",
			Extractor:         common.PathAuthenticationFlowAliasExtractor,
			RefFieldName:      "BrowserFlowRef",
			SelectorFieldName: "BrowserFlowSelector",
		}
		r.References["registration_flow"] = config.Reference{
			TerraformName:     "keycloak_authentication_flow",
			Extractor:         common.PathAuthenticationFlowAliasExtractor,
			RefFieldName:      "RegistrationFlowRef",
			SelectorFieldName: "RegistrationFlowSelector",
		}
		r.References["direct_grant_flow"] = config.Reference{
			TerraformName:     "keycloak_authentication_flow",
			Extractor:         common.PathAuthenticationFlowAliasExtractor,
			RefFieldName:      "DirectGrantFlowRef",
			SelectorFieldName: "DirectGrantFlowSelector",
		}
		r.References["reset_credentials_flow"] = config.Reference{
			TerraformName:     "keycloak_authentication_flow",
			Extractor:         common.PathAuthenticationFlowAliasExtractor,
			RefFieldName:      "ResetCredentialsFlowRef",
			SelectorFieldName: "ResetCredentialsFlowSelector",
		}
		r.References["client_authentication_flow"] = config.Reference{
			TerraformName:     "keycloak_authentication_flow",
			Extractor:         common.PathAuthenticationFlowAliasExtractor,
			RefFieldName:      "ClientAuthenticationFlowRef",
			SelectorFieldName: "ClientAuthenticationFlowSelector",
		}
		r.References["docker_authentication_flow"] = config.Reference{
			TerraformName:     "keycloak_authentication_flow",
			Extractor:         common.PathAuthenticationFlowAliasExtractor,
			RefFieldName:      "DockerAuthenticationFlowRef",
			SelectorFieldName: "DockerAuthenticationFlowSelector",
		}
	})
}

var flowIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "alias"},
	GetIDByExternalName:          getFlowIDByExternalName,
	GetIDByIdentifyingProperties: getFlowIDByIdentifyingProperties,
}

// FlowIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var FlowIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(flowIdentifyingPropertiesLookup)

func getFlowIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetAuthenticationFlow(ctx, parameters["realm_id"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getFlowIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetAuthenticationFlowFromAlias(ctx, parameters["realm_id"].(string), parameters["alias"].(string))
	if err != nil {
		if strings.Contains(err.Error(), "no authentication flow found for alias") {
			return "", nil
		}

		return "", err
	}

	return found.Id, nil
}

var subFlowIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "parent_flow_alias", "alias"},
	GetIDByExternalName:          getSubFlowIDByExternalName,
	GetIDByIdentifyingProperties: getSubFlowIDByIdentifyingProperties,
}

// SubFlowIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var SubFlowIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(subFlowIdentifyingPropertiesLookup)

func getSubFlowIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetAuthenticationSubFlow(ctx, parameters["realm_id"].(string), parameters["parent_flow_alias"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getSubFlowIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	executions, err := kcClient.ListAuthenticationExecutions(ctx, parameters["realm_id"].(string), parameters["parent_flow_alias"].(string))
	if err != nil {
		return "", err
	}

	filtered := lookup.Filter(executions, func(execution *keycloak.AuthenticationExecutionInfo) bool {
		return execution.AuthenticationFlow && execution.Level == 0
	})

	for _, flow := range filtered {
		subFlow, err := kcClient.GetAuthenticationSubFlow(ctx, parameters["realm_id"].(string), parameters["parent_flow_alias"].(string), flow.FlowId)
		if err != nil {
			return "", err
		}
		if subFlow != nil && subFlow.Alias == parameters["alias"].(string) {
			return subFlow.Id, nil
		}
	}

	return "", nil
}

var executionIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "parent_flow_alias", "authenticator"},
	GetIDByExternalName:          getExecutionIDByExternalName,
	GetIDByIdentifyingProperties: getExecutionIDByIdentifyingProperties,
}

// ExecutionIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var ExecutionIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(executionIdentifyingPropertiesLookup)

func getExecutionIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	found, err := kcClient.GetAuthenticationExecution(ctx, parameters["realm_id"].(string), parameters["parent_flow_alias"].(string), id)
	if err != nil {
		return "", err
	}
	return found.Id, nil
}

func getExecutionIDByIdentifyingProperties(ctx context.Context, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	executions, err := kcClient.ListAuthenticationExecutions(ctx, parameters["realm_id"].(string), parameters["parent_flow_alias"].(string))
	if err != nil {
		return "", err
	}

	// This limits the usage for authentication execution per flow to a single instance of the same ProviderId
	// A workaround is to encapsulate a duplicated execution with same ProviderId into a subFlow
	filtered := lookup.Filter(executions, func(execution *keycloak.AuthenticationExecutionInfo) bool {
		// execution.Level == 0 means that this execution is directly assigned to the parent_flow_alias
		// and not part of a nested subFlow
		return !execution.AuthenticationFlow && execution.ProviderId == parameters["authenticator"].(string) && execution.Level == 0
	})

	return lookup.SingleOrEmpty(filtered, func(execution *keycloak.AuthenticationExecutionInfo) string {
		return execution.Id
	})
}

var executionConfigIdentifyingPropertiesLookup = lookup.IdentifyingPropertiesLookupConfig{
	RequiredParameters:           []string{"realm_id", "execution_id"},
	GetIDByExternalName:          getExecutionConfigIDByExternalName,
	GetIDByIdentifyingProperties: getExecutionConfigIDByIdentifyingProperties,
}

// ExecutionConfigIdentifierFromIdentifyingProperties is used to find the existing resource by it´s identifying properties
var ExecutionConfigIdentifierFromIdentifyingProperties = lookup.BuildIdentifyingPropertiesLookup(executionConfigIdentifyingPropertiesLookup)

func getExecutionConfigIDByExternalName(ctx context.Context, id string, parameters map[string]any, kcClient *keycloak.KeycloakClient) (string, error) {
	executionConfig := keycloak.AuthenticationExecutionConfig{
		Id:      id,
		RealmId: parameters["realm_id"].(string),
	}

	err := kcClient.GetAuthenticationExecutionConfig(ctx, &executionConfig)
	if err != nil {
		return "", err
	}

	return executionConfig.Id, nil
}

func getExecutionConfigIDByIdentifyingProperties(_ context.Context, _ map[string]any, _ *keycloak.KeycloakClient) (string, error) {
	// If External-Name is not matching anymore we can simply create a new config
	// We do not need to try to find the existing one, because it´s a 1:1 relationship between Execution and ExecutionConfig
	// We can simply create it
	return "", nil
}
