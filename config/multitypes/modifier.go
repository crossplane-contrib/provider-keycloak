package multitypes

import (
	"github.com/crossplane/upjet/pkg/config"
	n "github.com/crossplane/upjet/pkg/types/name"
	"github.com/pkg/errors"
)

type Instance struct {
	Name      string
	Reference config.Reference
}

func apply(r *config.Resource, name string, types ...Instance) {
	for _, t := range types {
		cp := *r.TerraformResource.Schema[name]
		r.TerraformResource.Schema[t.Name] = &cp
		r.References[t.Name] = t.Reference
	}

	// Not Optional & Computed => Appear only in status
	// https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md#overriding-terraform-resource-schema
	r.TerraformResource.Schema[name].Optional = false
	r.TerraformResource.Schema[name].Computed = true

	delete(r.References, name)
}

func ApplyTo(r *config.Resource, name string, types ...Instance) {
	apply(r, name, types...)
	r.TerraformConfigurationInjector = wrapFuncAndConsolidate(r.TerraformConfigurationInjector, name, types)
}

func ApplyToList(r *config.Resource, name string, types ...Instance) {
	apply(r, name, types...)
	r.TerraformConfigurationInjector = wrapFuncAndConsolidateList(r.TerraformConfigurationInjector, name, types)
}

func wrapFuncAndConsolidate(ci config.ConfigurationInjector, name string, types []Instance) config.ConfigurationInjector {
	return func(jsonMap map[string]any, tfMap map[string]any) error {
		if ci != nil {
			err := ci(jsonMap, tfMap)
			if err != nil {
				return err
			}
		}

		isSetCount := 0
		allFields := ""
		setFields := ""
		for _, t := range types {
			tName := n.NewFromSnake(t.Name)

			if allFields == "" {
				allFields += tName.LowerCamelComputed
			} else {
				allFields += ", " + tName.LowerCamelComputed
			}

			if jsonMap[tName.LowerCamelComputed] != nil {
				if setFields == "" {
					setFields += tName.LowerCamelComputed
				} else {
					setFields += ", " + tName.LowerCamelComputed
				}
				isSetCount++
			}
		}

		if isSetCount > 1 {
			return errors.Errorf("Only one of these fields must be present '%s', but following are set '%s'!", allFields, setFields)
		}

		for _, t := range types {
			tName := n.NewFromSnake(t.Name)
			if jsonMap[tName.LowerCamelComputed] != nil {
				tfMap[name] = jsonMap[tName.LowerCamelComputed]
				delete(tfMap, tName.Snake)
			}
		}
		return nil
	}
}

func wrapFuncAndConsolidateList(ci config.ConfigurationInjector, name string, types []Instance) config.ConfigurationInjector {
	return func(jsonMap map[string]any, tfMap map[string]any) error {
		if ci != nil {
			err := ci(jsonMap, tfMap)
			if err != nil {
				return err
			}
		}

		var union []any
		for _, t := range types {
			tName := n.NewFromSnake(t.Name)
			value := jsonMap[tName.LowerCamelComputed]
			if value != nil {
				if list, ok := value.([]any); ok {
					union = append(union, list...)
					delete(tfMap, tName.Snake)
				}
			}
		}
		if union != nil {
			tfMap[name] = union
		}
		return nil
	}
}
