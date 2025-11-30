// Package multitypes provides utilities for creating multiple strongly-typed
// reference fields from a single Terraform field that accepts different resource types.
//
// # Problem Statement
//
// Some Terraform resources have fields that can reference multiple different
// resource types. For example, a Keycloak authentication flow execution might
// have an "authenticator" field that can reference either an authenticator name
// or a flow alias. In the Terraform schema, this is a single string field.
//
// However, in Crossplane's resource model, we want to provide strongly-typed
// cross-resource references (using Ref/Selector fields) for each possible
// resource type. This creates better user experience and type safety.
//
// # Solution
//
// This package implements a "multi-type reference" pattern by:
//
//  1. Creating synthetic fields in the CRD schema for each reference type
//     (e.g., "authenticatorRef" and "flowAliasRef")
//  2. Configuring cross-resource references on these synthetic fields
//  3. At runtime, consolidating values from synthetic fields back to the
//     original Terraform field name before sending to Terraform
//
// # Usage Example
//
// In Keycloak, an authentication execution's parent_flow_alias field can reference
// either an authentication flow or a subflow. We use multitypes to provide
// strongly-typed references for both:
//
//	multitypes.ApplyToWithOptions(r, "parent_flow_alias",
//	    &multitypes.Options{KeepOriginalField: true},  // Explicit: keep for backward compatibility
//	    multitypes.Instance{
//	        Name: "parent_flow_alias",  // Reuses original name
//	        Reference: config.Reference{
//	            TerraformName:     "keycloak_authentication_flow",
//	            Extractor:         common.PathAuthenticationFlowAliasExtractor,
//	            RefFieldName:      "ParentFlowAliasRef",
//	            SelectorFieldName: "ParentFlowAliasSelector",
//	        },
//	    },
//	    multitypes.Instance{
//	        Name: "parent_subflow_alias",  // New synthetic field for subflow
//	        Reference: config.Reference{
//	            TerraformName:     "keycloak_authentication_subflow",
//	            Extractor:         common.PathAuthenticationFlowAliasExtractor,
//	            RefFieldName:      "ParentSubflowAliasRef",
//	            SelectorFieldName: "ParentSubflowAliasSelector",
//	        },
//	    },
//	)
//
// This generates fields in the CRD:
//   - spec.forProvider.parentFlowAlias (with parentFlowAliasRef/Selector for Flow)
//   - spec.forProvider.parentSubflowAlias (with parentSubflowAliasRef/Selector for Subflow)
//   - status.atProvider.parentFlowAlias (the actual value from Terraform)
//
// Users can then reference either a Flow or Subflow:
//
//	apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
//	kind: Execution
//	metadata:
//	  name: my-execution
//	spec:
//	  forProvider:
//	    realmId: my-realm
//	    authenticator: identity-provider-redirector
//	    # Option 1: Reference a Flow
//	    parentFlowAliasRef:
//	      name: my-authentication-flow
//	    # OR Option 2: Reference a Subflow
//	    # parentSubflowAliasRef:
//	    #   name: my-authentication-subflow
//
// # References
//
// - Cross-resource referencing: https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md#cross-resource-referencing
// - Overriding schema: https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md#overriding-terraform-resource-schema
// - ConfigurationInjector: github.com/crossplane/upjet/v2/pkg/config.ConfigurationInjector
package multitypes

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Instance represents a single typed variant of a multi-type field.
// Each Instance creates a separate field in the generated CRD with its own
// cross-resource reference configuration.
//
// For example, for Keycloak's parent_flow_alias field that can reference either
// an authentication flow or a subflow, you would create two Instances:
//   - Instance{Name: "parent_flow_alias", Reference: config.Reference{TerraformName: "keycloak_authentication_flow"}}
//   - Instance{Name: "parent_subflow_alias", Reference: config.Reference{TerraformName: "keycloak_authentication_subflow"}}
type Instance struct {
	// Name is the field name for this typed variant in snake_case.
	// This will be converted to CamelCase in the generated CRD.
	// If Name matches the original Terraform field name, the original field
	// will remain in spec.forProvider. Otherwise, the original field becomes
	// computed-only (status-only).
	//
	// Example: "parent_subflow_alias" becomes "parentSubflowAlias" in the CRD
	Name string

	// Reference configures the cross-resource reference for this typed variant.
	// See: https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md#cross-resource-referencing
	//
	// The most important field is TerraformName, which identifies the target
	// Terraform resource type (e.g., "keycloak_authentication_flow", "keycloak_authentication_subflow").
	Reference config.Reference
}

// Options configures optional behavior for multi-type field generation.
type Options struct {
	// KeepOriginalField, when true, allows an Instance to reuse the original
	// Terraform field name, keeping it as a settable input field (in spec)
	// instead of making it computed-only (status).
	//
	// This is useful for backward compatibility when adding multi-type support
	// to an existing field - the original field remains usable.
	//
	// When false (default), if any Instance tries to reuse the original field
	// name, an error will be raised to prevent unclear behavior.
	// The original field will be moved to status-only (computed).
	//
	// Set this to true explicitly when you want backward compatibility.
	KeepOriginalField bool
}

// apply modifies the Terraform resource schema to support multi-type references.
// This is an internal function called by ApplyTo and ApplyToWithOptions.
//
// # What it does:
//
//  1. For each Instance, either:
//     a. If Instance.Name matches the original field name: configures cross-resource
//     reference on the existing field
//     b. Otherwise: creates a synthetic copy of the field with the Instance.Name
//     and configures its cross-resource reference
//
//  2. Determines the fate of the original Terraform field:
//     a. If any Instance reuses the original name: field remains in spec.forProvider
//     b. Otherwise: marks the field as Computed-only (moves to status.atProvider)
//
// # Why mark as Computed?
//
// When the original field becomes status-only (Optional=false, Computed=true),
// it follows upjet's schema override pattern. This tells upjet to:
// - Generate the field in status.atProvider (for observing the value)
// - NOT generate it in spec.forProvider (users can't set it directly)
//
// This is necessary because users will set values via the typed synthetic fields
// (e.g., loadBalancerIdRef), not the original field. At runtime, our
// ConfigurationInjector (wrapFuncAndConsolidate) will copy the resolved value
// back to the original field name for Terraform.
//
// See upjet documentation on schema overrides:
// https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md#overriding-terraform-resource-schema
//
// # Schema modification details:
//
// upjet uses Terraform's schema.Schema structure where:
// - Optional=true, Computed=false: field in spec (user sets value)
// - Optional=false, Computed=true: field in status (provider computes value)
// - Optional=true, Computed=true: field in spec, late-initialized from status
//
// See: github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema.Schema
func apply(r *config.Resource, name string, opts *Options, types ...Instance) {
	// Check if any instance reuses the original field name
	hasOriginalName := false
	for _, t := range types {
		if t.Name == name {
			hasOriginalName = true
			break
		}
	}

	// Validate: if an Instance reuses the original name but KeepOriginalField is not set,
	// this is likely unintentional and we should error
	if hasOriginalName && (opts == nil || !opts.KeepOriginalField) {
		panic("multitypes: Instance reuses original field name '" + name + "' but Options.KeepOriginalField is not set to true. " +
			"Set KeepOriginalField=true explicitly for backward compatibility, or use a different name for the Instance.")
	}

	// Validate: if KeepOriginalField is set but no Instance reuses the name, warn about misconfiguration
	if opts != nil && opts.KeepOriginalField && !hasOriginalName {
		panic("multitypes: Options.KeepOriginalField is true but no Instance reuses the original field name '" + name + "'. " +
			"Either add an Instance with Name='" + name + "', or set KeepOriginalField=false.")
	}

	// Configure each typed variant
	for _, t := range types {
		if t.Name == name {
			// This instance reuses the original field name (with explicit permission via Options)
			// Set the reference on the original field
			// The field remains Optional in spec.forProvider for backward compatibility
			r.References[t.Name] = t.Reference
		} else {
			// Create a synthetic field for other types
			// This copies the schema from the original field, inheriting
			// all its properties (type, description, validation, etc.)
			cp := *r.TerraformResource.Schema[name]
			r.TerraformResource.Schema[t.Name] = &cp

			// Configure cross-resource reference for the synthetic field
			// This enables generation of Ref/Selector fields in the CRD
			// See: github.com/crossplane/upjet/v2/pkg/config.Reference
			r.References[t.Name] = t.Reference
		}
	}

	// Determine the fate of the original Terraform field
	if !hasOriginalName {
		// No instance reuses the original name, so users will only interact
		// with the synthetic typed fields. Make the original field computed-only
		// so it appears only in status.atProvider for observation.
		//
		// Schema behavior: Optional=false, Computed=true => status-only field
		// See: https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md#overriding-terraform-resource-schema
		r.TerraformResource.Schema[name].Optional = false
		r.TerraformResource.Schema[name].Computed = true

		// Remove any reference configuration from the original field since
		// it's now observation-only (users can't set it)
		delete(r.References, name)
	}
	// else: hasOriginalName is true AND KeepOriginalField was explicitly set,
	// the original field remains Optional and stays in spec.forProvider for backward compatibility
}

// ApplyTo configures multi-type reference support for a single Terraform field.
//
// This is the main entry point for adding multi-type references to a resource.
// It modifies both the schema and runtime behavior to support multiple strongly-typed
// reference fields that all map to the same underlying Terraform field.
//
// # Parameters:
//
//   - r: The upjet Resource being configured
//   - name: The original Terraform field name (in snake_case, e.g., "parent_flow_alias")
//   - types: One or more Instance definitions, each representing a typed variant
//
// # Example:
//
// For Keycloak authentication execution's parent_flow_alias field (with backward compatibility):
//
//	multitypes.ApplyToWithOptions(r, "parent_flow_alias",
//	    &multitypes.Options{KeepOriginalField: true},  // Explicit backward compatibility
//	    multitypes.Instance{
//	        Name: "parent_flow_alias",  // Reuses original name
//	        Reference: config.Reference{
//	            TerraformName: "keycloak_authentication_flow",
//	            Extractor:     common.PathAuthenticationFlowAliasExtractor,
//	        },
//	    },
//	    multitypes.Instance{
//	        Name: "parent_subflow_alias",  // Add new typed field
//	        Reference: config.Reference{
//	            TerraformName: "keycloak_authentication_subflow",
//	            Extractor:     common.PathAuthenticationFlowAliasExtractor,
//	        },
//	    },
//	)
//
// Or for a clean break without backward compatibility (original field moves to status):
//
//	multitypes.ApplyTo(r, "parent_flow_alias",
//	    multitypes.Instance{
//	        Name: "parent_flow_ref",  // Different name
//	        Reference: config.Reference{
//	            TerraformName: "keycloak_authentication_flow",
//	        },
//	    },
//	    multitypes.Instance{
//	        Name: "parent_subflow_ref",  // Different name
//	        Reference: config.Reference{
//	            TerraformName: "keycloak_authentication_subflow",
//	        },
//	    },
//	)
//
// This creates fields in the CRD like:
//   - spec.forProvider.parentFlowAlias (with parentFlowAliasRef/Selector)
//   - spec.forProvider.parentSubflowAlias (with parentSubflowAliasRef/Selector)
//   - status.atProvider.parentFlowAlias (the actual value observed from Terraform)
//
// # How it works:
//
//  1. Schema modification: Creates synthetic fields and configures references
//  2. Runtime injection: Installs a ConfigurationInjector that consolidates
//     values from synthetic fields back to the original Terraform field
//
// See also: ApplyToWithOptions for configuration with Options.
func ApplyTo(r *config.Resource, name string, types ...Instance) {
	ApplyToWithOptions(r, name, nil, types...)
}

// ApplyToWithOptions is like ApplyTo but accepts Options for explicit behavior control.
//
// Use this when you need explicit control over whether the original field should
// be kept for backward compatibility (via Options.KeepOriginalField).
//
// # Implementation details:
//
//  1. Calls apply() to modify the resource schema and configure references
//  2. Wraps the existing TerraformConfigurationInjector (if any) with
//     wrapFuncAndConsolidate to handle runtime value consolidation
//
// The TerraformConfigurationInjector is called during the resource reconciliation
// process by upjet's external client implementations:
// - github.com/crossplane/upjet/v2/pkg/controller/external_tfpluginsdk.go
// - github.com/crossplane/upjet/v2/pkg/controller/external_tfpluginfw.go
//
// See also: config.ConfigurationInjector documentation in upjet
func ApplyToWithOptions(r *config.Resource, name string, opts *Options, types ...Instance) {
	apply(r, name, opts, types...)

	// Wrap any existing ConfigurationInjector with our consolidation logic
	// This ensures values from synthetic fields are consolidated to the original
	// Terraform field name before being sent to Terraform
	r.TerraformConfigurationInjector = wrapFuncAndConsolidate(r.TerraformConfigurationInjector, name, types)
}

// wrapFuncAndConsolidate returns a ConfigurationInjector that wraps an existing
// injector and adds logic to consolidate multi-type field values.
//
// # The ConfigurationInjector Contract:
//
// A ConfigurationInjector is called by upjet during resource reconciliation to
// modify the Terraform configuration before it's sent to the Terraform provider.
// It receives two maps:
//
//  1. jsonMap: Deserialized from spec.forProvider using JSON tags (may use original field names)
//  2. tfMap: Converted using Terraform field tags (always uses snake_case Terraform names)
//
// The injector can modify these maps to inject additional configuration or
// transform values. Both maps eventually get merged and sent to Terraform.
//
// See upjet documentation:
// - ConfigurationInjector type: github.com/crossplane/upjet/v2/pkg/config.ConfigurationInjector
// - Called from: github.com/crossplane/upjet/v2/pkg/controller/external_tfpluginsdk.go (getExtendedParameters)
// - Called from: github.com/crossplane/upjet/v2/pkg/controller/external_tfpluginfw.go (getFrameworkExtendedParameters)
//
// # What this function does:
//
// Returns a new ConfigurationInjector that:
//  1. Calls the wrapped injector (ci) first, if present
//  2. Looks for the first non-nil value among the synthetic typed fields
//  3. Copies that value to the original Terraform field name
//  4. Removes the synthetic field entries to avoid sending them to Terraform
//
// # Why consolidation is necessary:
//
// Users interact with typed fields like "parentFlowAliasRef" or "parentSubflowAliasRef"
// in the CRD, but Terraform expects the original field name like "parent_flow_alias".
// This function bridges the gap by:
// - Finding which typed field the user populated (via direct value or reference)
// - Moving its resolved value to the original Terraform field name
// - Cleaning up the synthetic field names so they don't confuse Terraform
//
// # Reference Resolution:
//
// By the time this ConfigurationInjector runs, Crossplane has already resolved
// any Ref/Selector fields. So if the user specified:
//
//	spec:
//	  forProvider:
//	    parentFlowAliasRef:
//	      name: my-authentication-flow
//
// The jsonMap/tfMap will already contain the resolved alias:
//
//	{"parent_flow_alias": "browser"}
//
// This function finds that resolved value and ensures it's set on "parent_flow_alias".
//
// # Parameters:
//
//   - ci: An existing ConfigurationInjector to wrap (may be nil)
//   - name: The original Terraform field name
//   - types: The list of typed Instance definitions
//
// # Returns:
//
// A new ConfigurationInjector that performs consolidation after calling ci.
func wrapFuncAndConsolidate(ci config.ConfigurationInjector, name string, types []Instance) config.ConfigurationInjector {
	return func(jsonMap map[string]any, tfMap map[string]any) error {
		// First, call any existing injector that was configured
		// This preserves the existing behavior and allows chaining injectors
		if ci != nil {
			err := ci(jsonMap, tfMap)
			if err != nil {
				return err
			}
		}

		// Find the first non-nil synthetic field value and use it
		//
		// We check both maps because:
		// - jsonMap: Contains values as deserialized from spec.forProvider
		//   Field names might use the original struct field names
		// - tfMap: Contains values with Terraform field tags applied
		//   Field names use snake_case Terraform naming
		//
		// At this point, both maps should have the synthetic field names
		// (e.g., "parent_flow_alias", "parent_subflow_alias") with resolved values
		// from reference resolution or direct user input.
		var selectedValue any

		for _, t := range types {
			// Try jsonMap first (might use snake_case tf tags)
			if val := jsonMap[t.Name]; val != nil {
				selectedValue = val
				break
			}
			// Also try tfMap (definitely uses snake_case)
			if val := tfMap[t.Name]; val != nil {
				selectedValue = val
				break
			}
		}

		// Consolidate: copy the selected value to the original Terraform field name
		if selectedValue != nil {
			// Set the value in both tfMap and jsonMap using the original field name
			// This ensures Terraform receives the value under the expected field name
			//
			// Both maps are used by upjet when building the Terraform configuration:
			// - tfMap is the primary source for Terraform HCL generation
			// - jsonMap is used for certain parameter lookups and validations
			//
			// Both maps use snake_case field names (Terraform convention)
			tfMap[name] = selectedValue
			jsonMap[name] = selectedValue

			// Clean up: delete all synthetic field entries from both maps
			// These synthetic fields don't exist in the Terraform schema, so
			// sending them would cause Terraform to reject the configuration
			for _, t := range types {
				// Skip if this is the original field name (we just set it above)
				if t.Name != name {
					delete(tfMap, t.Name)
					delete(jsonMap, t.Name)
				}
			}
		}
		// If no synthetic field has a value and the original field is also nil,
		// this might be a case where:
		// - The resource is being created and references haven't resolved yet
		// - All fields are optional and the user didn't set any
		// - The field has a default value that will be populated later
		//
		// Don't treat this as an error - let Terraform validation handle it
		// if the field is actually required.

		return nil
	}
}
