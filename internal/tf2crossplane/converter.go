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
	"sort"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/crossplane/upjet/v2/pkg/types/name"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"

	kcconfig "github.com/crossplane-contrib/provider-keycloak/config"
)

// Options controls how the converter renders Crossplane manifests.
type Options struct {
	// Namespaced selects the namespaced provider variant
	// (keycloak.m.crossplane.io) instead of the cluster-scoped one
	// (keycloak.crossplane.io).
	Namespaced bool

	// ProviderConfigRef is the name written into spec.providerConfigRef.name
	// for every generated resource. When empty, no providerConfigRef is
	// emitted.
	ProviderConfigRef string

	// DeletionPolicy, when set, is written into spec.deletionPolicy.
	DeletionPolicy string

	// ManagementPolicies, when non-empty, is written into
	// spec.managementPolicies.
	ManagementPolicies []string
}

// Converter turns Terraform HCL into Crossplane manifests. It is safe for
// concurrent use once constructed.
type Converter struct {
	provider *config.Provider
	opts     Options
}

// New builds a Converter using the same provider configuration the code
// generator relies on, so the Terraform<->CRD mapping is always in lockstep
// with the generated CRDs.
func New(opts Options) (*Converter, error) {
	var (
		p   *config.Provider
		err error
	)
	if opts.Namespaced {
		p, err = kcconfig.GetProviderNamespaced(true)
	} else {
		p, err = kcconfig.GetProvider(true)
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot load provider configuration")
	}
	return &Converter{provider: p, opts: opts}, nil
}

// Result is the outcome of converting one or more HCL sources.
type Result struct {
	// Documents holds the rendered YAML documents in input order.
	Documents []Document
	// Warnings collects non-fatal issues encountered during conversion
	// (unresolved expressions, unknown fields, etc.).
	Warnings []string
	// Unsupported lists Terraform resource types that have no corresponding
	// CRD in provider-keycloak.
	Unsupported []string
}

// Document is a single rendered Crossplane manifest along with the source it
// originated from.
type Document struct {
	// TerraformType is the Terraform resource type (e.g. keycloak_group) or a
	// synthetic label such as "provider" for scaffolding.
	TerraformType string
	// TerraformName is the Terraform local name (e.g. "this").
	TerraformName string
	// Manifest is the rendered YAML (without the leading document separator).
	Manifest string
}

// Convert parses the given HCL source and renders the corresponding Crossplane
// manifests. filename is only used for diagnostics.
func (c *Converter) Convert(src []byte, filename string) (*Result, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL(src, filename)
	if diags.HasErrors() {
		return nil, errors.Errorf("cannot parse HCL: %s", diags.Error())
	}
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("unexpected HCL body type (only native HCL syntax is supported)")
	}

	res := &Result{}
	for _, block := range body.Blocks {
		switch block.Type {
		case "resource":
			c.convertResourceBlock(block, src, res)
		case "provider":
			c.convertProviderBlock(block, res)
		case "data":
			res.Warnings = append(res.Warnings, fmt.Sprintf("skipping data source %q: data sources are not managed resources; import the existing object instead", blockLabels(block)))
		case "module":
			res.Warnings = append(res.Warnings, fmt.Sprintf("skipping module %q: modules must be converted by expanding them into their resources", blockLabels(block)))
		case "variable", "locals", "output", "terraform":
			// These do not map to Crossplane resources; silently ignore.
		default:
			res.Warnings = append(res.Warnings, fmt.Sprintf("skipping unsupported top-level block %q", block.Type))
		}
	}

	sort.Strings(res.Unsupported)
	res.Unsupported = dedupe(res.Unsupported)
	return res, nil
}

func (c *Converter) convertResourceBlock(block *hclsyntax.Block, src []byte, res *Result) {
	if len(block.Labels) != 2 {
		res.Warnings = append(res.Warnings, "skipping malformed resource block (expected type and name labels)")
		return
	}
	tfType, localName := block.Labels[0], block.Labels[1]

	r, found := c.provider.Resources[tfType]
	if !found {
		res.Unsupported = append(res.Unsupported, tfType)
		res.Warnings = append(res.Warnings, fmt.Sprintf("no CRD found for Terraform resource %q; skipping %s.%s", tfType, tfType, localName))
		return
	}

	forProvider, warns := c.bodyToMap(block.Body, r.TerraformResource.Schema, r, "", src)
	res.Warnings = append(res.Warnings, warns...)

	doc := crd{
		APIVersion: c.apiVersion(r),
		Kind:       r.Kind,
		Metadata:   objectMeta{Name: sanitizeName(localName)},
		Spec: crdSpec{
			ForProvider:        forProvider,
			DeletionPolicy:     c.opts.DeletionPolicy,
			ManagementPolicies: c.opts.ManagementPolicies,
		},
	}
	if c.opts.ProviderConfigRef != "" {
		doc.Spec.ProviderConfigRef = &localRef{Name: c.opts.ProviderConfigRef}
	}

	manifest, err := marshalYAML(doc)
	if err != nil {
		res.Warnings = append(res.Warnings, fmt.Sprintf("cannot render %s.%s: %v", tfType, localName, err))
		return
	}
	res.Documents = append(res.Documents, Document{
		TerraformType: tfType,
		TerraformName: localName,
		Manifest:      manifest,
	})
}

// bodyToMap walks an HCL body against a Terraform SDK schema, producing the
// camelCased spec.forProvider representation. path is the dotted reference key
// prefix used to look up cross-resource references for nested fields.
func (c *Converter) bodyToMap(body *hclsyntax.Body, sch map[string]*schema.Schema, r *config.Resource, path string, src []byte) (map[string]any, []string) { //nolint:gocyclo // The branching mirrors the Terraform schema value types and is clearer inline.
	out := map[string]any{}
	var warns []string

	// Attributes: scalars, collections, maps, and references.
	attrNames := make([]string, 0, len(body.Attributes))
	for n := range body.Attributes {
		attrNames = append(attrNames, n)
	}
	sort.Strings(attrNames)
	for _, tfName := range attrNames {
		attr := body.Attributes[tfName]
		fieldSchema := sch[tfName]
		if fieldSchema == nil {
			warns = append(warns, fmt.Sprintf("%s: unknown field %q has no CRD schema; skipping", joinPath(path, tfName), tfName))
			continue
		}
		refKey := joinPath(path, tfName)

		// Reference to another Keycloak-managed resource?
		if tfType, refLocal, ok := resourceReference(attr.Expr); ok {
			if _, isRef := r.References[refKey]; isRef {
				refField := name.ReferenceFieldName(name.NewFromSnake(tfName), false, "").LowerCamelComputed
				out[refField] = map[string]any{"name": sanitizeName(refLocal)}
				continue
			}
			warns = append(warns, fmt.Sprintf("%s: value references %s but this field is not a configured reference; emitting a placeholder", refKey, tfType))
			out[camel(tfName)] = exprText(attr.Expr, src)
			continue
		}

		// Unresolved expression (var./local./data./module./functions)?
		if len(attr.Expr.Variables()) > 0 {
			warns = append(warns, fmt.Sprintf("%s: value is a dynamic expression that cannot be evaluated statically; emitting a placeholder", refKey))
			out[camel(tfName)] = exprText(attr.Expr, src)
			continue
		}

		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			warns = append(warns, fmt.Sprintf("%s: cannot evaluate value: %s; emitting a placeholder", refKey, diags.Error()))
			out[camel(tfName)] = exprText(attr.Expr, src)
			continue
		}
		gv, aw := ctyToGo(val, fieldSchema, refKey)
		warns = append(warns, aw...)
		out[camel(tfName)] = gv
	}

	// Nested blocks: always rendered as lists to match Upjet's generated CRDs,
	// which model even MaxItems==1 blocks as arrays.
	grouped := map[string][]*hclsyntax.Block{}
	order := []string{}
	for _, b := range body.Blocks {
		if _, seen := grouped[b.Type]; !seen {
			order = append(order, b.Type)
		}
		grouped[b.Type] = append(grouped[b.Type], b)
	}
	for _, tfName := range order {
		fieldSchema := sch[tfName]
		if fieldSchema == nil {
			warns = append(warns, fmt.Sprintf("%s: unknown block %q has no CRD schema; skipping", joinPath(path, tfName), tfName))
			continue
		}
		elem, ok := fieldSchema.Elem.(*schema.Resource)
		if !ok {
			warns = append(warns, fmt.Sprintf("%s: block %q does not map to a nested object; skipping", joinPath(path, tfName), tfName))
			continue
		}
		list := make([]any, 0, len(grouped[tfName]))
		for _, b := range grouped[tfName] {
			m, bw := c.bodyToMap(b.Body, elem.Schema, r, joinPath(path, tfName), src)
			warns = append(warns, bw...)
			list = append(list, m)
		}
		out[camel(tfName)] = list
	}

	return out, warns
}

func (c *Converter) convertProviderBlock(block *hclsyntax.Block, res *Result) {
	if len(block.Labels) != 1 || block.Labels[0] != "keycloak" {
		return
	}
	res.Warnings = append(res.Warnings, "translated provider \"keycloak\" block into a ProviderConfig + Secret scaffold; fill in credentials and never commit secret values")
	res.Documents = append(res.Documents,
		Document{TerraformType: "provider", TerraformName: "keycloak", Manifest: strings.TrimSpace(providerConfigScaffold)},
		Document{TerraformType: "secret", TerraformName: "keycloak", Manifest: strings.TrimSpace(credentialsSecretScaffold)},
	)
}

// apiVersion computes the CRD apiVersion for a resource, e.g.
// group.keycloak.crossplane.io/v1alpha1.
func (c *Converter) apiVersion(r *config.Resource) string {
	group := c.provider.RootGroup
	if r.ShortGroup != "" {
		group = r.ShortGroup + "." + c.provider.RootGroup
	}
	return group + "/" + r.Version
}

func blockLabels(b *hclsyntax.Block) string {
	return strings.Join(b.Labels, ".")
}

func camel(tfName string) string {
	return name.NewFromSnake(tfName).LowerCamelComputed
}

func joinPath(path, field string) string {
	if path == "" {
		return field
	}
	return path + "." + field
}

func dedupe(in []string) []string {
	if len(in) == 0 {
		return in
	}
	out := in[:1]
	for _, s := range in[1:] {
		if s != out[len(out)-1] {
			out = append(out, s)
		}
	}
	return out
}

// SupportedTypes returns the sorted list of Terraform resource types the
// converter can map to CRDs.
func (c *Converter) SupportedTypes() []string {
	types := make([]string, 0, len(c.provider.Resources))
	for t := range c.provider.Resources {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}
