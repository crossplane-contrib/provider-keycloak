package multitypes

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

type Instance struct {
	Name      string
	Reference config.Reference
}

type Options struct {
	// KeepOriginalField, when true, keeps the original Terraform field as a
	// settable input field (in spec) instead of making it computed-only (status).
	// This is useful for backward compatibility when adding multi-type support
	// to an existing field - the original field remains usable.
	// Default is false (original field becomes computed-only).
	KeepOriginalField bool
}

func apply(r *config.Resource, name string, opts *Options, types ...Instance) {
	// Check if any instance reuses the original field name (for backward compatibility)
	hasOriginalName := false
	for _, t := range types {
		if t.Name == name {
			hasOriginalName = true
			break
		}
	}

	for _, t := range types {
		if t.Name == name {
			// This instance reuses the original field name
			// Set the reference on the original field
			r.References[t.Name] = t.Reference
		} else {
			// Create a synthetic field for other types
			cp := *r.TerraformResource.Schema[name]
			r.TerraformResource.Schema[t.Name] = &cp
			r.References[t.Name] = t.Reference
		}
	}

	// The original field is kept settable when an instance reuses its name
	// Otherwise, make it computed-only (appears only in status)
	if !hasOriginalName {
		// Not Optional & Computed => Appear only in status
		// See: https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md#overriding-terraform-resource-schema
		r.TerraformResource.Schema[name].Optional = false
		r.TerraformResource.Schema[name].Computed = true
		delete(r.References, name)
	}
}

func ApplyTo(r *config.Resource, name string, types ...Instance) {
	ApplyToWithOptions(r, name, nil, types...)
}

func ApplyToWithOptions(r *config.Resource, name string, opts *Options, types ...Instance) {
	apply(r, name, opts, types...)

	r.TerraformConfigurationInjector = wrapFuncAndConsolidate(r.TerraformConfigurationInjector, name, types)
}

func wrapFuncAndConsolidate(ci config.ConfigurationInjector, name string, types []Instance) config.ConfigurationInjector {
	return func(jsonMap map[string]any, tfMap map[string]any) error {
		if ci != nil {
			err := ci(jsonMap, tfMap)
			if err != nil {
				return err
			}
		}

		// Find the first non-nil synthetic field value and use it
		// jsonMap might use either snake_case (tf tag) or the field is in tfMap
		// Check both jsonMap and tfMap for the synthetic field values
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

		// If we found a value, consolidate it to the original field
		if selectedValue != nil {
			// Set the value in both tfMap and jsonMap for Terraform operations and parameter lookups
			// Both maps use the original Terraform field name (snake_case)
			tfMap[name] = selectedValue
			jsonMap[name] = selectedValue

			// Clean up: delete all synthetic field entries from both maps
			for _, t := range types {
				// Skip if this is the original field name (we just set it)
				if t.Name != name {
					delete(tfMap, t.Name)
					delete(jsonMap, t.Name)
				}
			}
		} else if jsonMap[name] == nil {
			// If no synthetic field has a value and the original field is also nil,
			// this might be a case where the resource hasn't resolved references yet
			// Don't error, just let it continue
		}

		return nil
	}
}
