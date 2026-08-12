/*
Copyright 2024 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tf2crossplane

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zclconf/go-cty/cty"
)

// ctyToGo converts a statically-evaluated HCL value into a plain Go value
// suitable for YAML marshalling, using the Terraform schema to decide how
// nested keys should be transformed (camelCased for typed objects, kept
// verbatim for free-form maps).
func ctyToGo(val cty.Value, sch *schema.Schema, path string) (any, []string) { //nolint:gocyclo // Mirrors the cty type switch; splitting it would obscure the mapping.
	if val.IsNull() {
		return nil, nil
	}
	t := val.Type()
	switch {
	case t == cty.String:
		return val.AsString(), nil
	case t == cty.Bool:
		return val.True(), nil
	case t == cty.Number:
		return numberToGo(val), nil
	case t.IsTupleType() || t.IsListType() || t.IsSetType():
		return ctySeqToGo(val, sch, path)
	case t.IsObjectType() || t.IsMapType():
		return ctyMapToGo(val, sch, path)
	default:
		return nil, []string{fmt.Sprintf("%s: unsupported value type %s; skipping", path, t.FriendlyName())}
	}
}

func numberToGo(val cty.Value) any {
	bf := val.AsBigFloat()
	if bf.IsInt() {
		i, _ := bf.Int64()
		return i
	}
	f, _ := bf.Float64()
	return f
}

func ctySeqToGo(val cty.Value, sch *schema.Schema, path string) (any, []string) {
	var (
		out   []any
		warns []string
	)
	elemSchema := elementSchema(sch)
	for it := val.ElementIterator(); it.Next(); {
		_, ev := it.Element()
		gv, w := ctyToGo(ev, elemSchema, path+"[]")
		warns = append(warns, w...)
		out = append(out, gv)
	}
	if out == nil {
		out = []any{}
	}
	return out, warns
}

func ctyMapToGo(val cty.Value, sch *schema.Schema, path string) (any, []string) {
	var warns []string
	out := map[string]any{}

	// A typed nested object (schema Elem is *schema.Resource) has known field
	// names that must be camelCased; a free-form map keeps its keys verbatim.
	nested, typed := nestedResourceSchema(sch)
	for it := val.ElementIterator(); it.Next(); {
		k, ev := it.Element()
		key := k.AsString()
		var childSchema *schema.Schema
		outKey := key
		if typed {
			childSchema = nested[key]
			outKey = camel(key)
		} else {
			childSchema = elementSchema(sch)
		}
		gv, w := ctyToGo(ev, childSchema, joinPath(path, key))
		warns = append(warns, w...)
		out[outKey] = gv
	}
	return out, warns
}

// elementSchema returns the schema describing the elements of a collection, if
// available.
func elementSchema(sch *schema.Schema) *schema.Schema {
	if sch == nil {
		return nil
	}
	if es, ok := sch.Elem.(*schema.Schema); ok {
		return es
	}
	return nil
}

// nestedResourceSchema returns the field schema map for a typed nested object,
// reporting whether the schema describes such an object.
func nestedResourceSchema(sch *schema.Schema) (map[string]*schema.Schema, bool) {
	if sch == nil {
		return nil, false
	}
	if r, ok := sch.Elem.(*schema.Resource); ok {
		return r.Schema, true
	}
	return nil, false
}
