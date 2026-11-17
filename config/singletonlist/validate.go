// Package singletonlist rejects managed resources that set more than one entry
// for a Terraform nested block declared with MaxItems = 1.
//
// Invalid extra entries can cause perpetual reconciliation, as with UserProfile
// permissions in crossplane-contrib/provider-keycloak#745.
package singletonlist

import (
	"fmt"
	"sort"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	n "github.com/crossplane/upjet/v2/pkg/types/name"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

type block struct {
	tfName    string
	crdName   string
	singleton bool
	children  []block
}

// Validate returns a resource option that validates the Terraform parameters
// of a managed resource against the MaxItems = 1 constraint of its nested
// blocks before they are sent to Terraform.
func Validate() config.ResourceOption {
	return func(r *config.Resource) {
		if r.TerraformResource == nil {
			return
		}
		blocks := collect(r.TerraformResource.Schema)
		if len(blocks) == 0 {
			return
		}
		r.TerraformConfigurationInjector = wrap(r.TerraformConfigurationInjector, blocks)
	}
}

func wrap(ci config.ConfigurationInjector, blocks []block) config.ConfigurationInjector {
	return func(jsonMap map[string]any, tfMap map[string]any) error {
		if ci != nil {
			if err := ci(jsonMap, tfMap); err != nil {
				return err
			}
		}
		if violations := check(tfMap, blocks, ""); len(violations) > 0 {
			return errors.Errorf("invalid spec.forProvider/spec.initProvider: %s", strings.Join(violations, "; "))
		}
		return nil
	}
}

// collect returns the nested blocks of a schema that are, or contain,
// singleton blocks.
func collect(s map[string]*schema.Schema) []block {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var blocks []block
	for _, k := range keys {
		v := s[k]
		if v.Type != schema.TypeList && v.Type != schema.TypeSet {
			continue
		}
		res, ok := v.Elem.(*schema.Resource)
		if !ok {
			continue
		}
		b := block{
			tfName:    k,
			crdName:   n.NewFromSnake(k).LowerCamelComputed,
			singleton: v.MaxItems == 1,
			children:  collect(res.Schema),
		}
		if b.singleton || len(b.children) > 0 {
			blocks = append(blocks, b)
		}
	}
	return blocks
}

func check(params map[string]any, blocks []block, prefix string) []string {
	var violations []string
	for _, b := range blocks {
		items, ok := params[b.tfName].([]any)
		if !ok {
			continue
		}
		p := b.crdName
		if prefix != "" {
			p = prefix + "." + b.crdName
		}
		if b.singleton && len(items) > 1 {
			violations = append(violations, fmt.Sprintf("%s has %d entries but allows at most 1; merge them into a single entry", p, len(items)))
		}
		for i, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			violations = append(violations, check(m, b.children, fmt.Sprintf("%s[%d]", p, i))...)
		}
	}
	return violations
}
