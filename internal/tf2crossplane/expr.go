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
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// resourceReference reports whether expr is a reference to another Keycloak
// managed resource (e.g. keycloak_realm.this.id) and, if so, returns the
// referenced Terraform type and local name.
func resourceReference(expr hclsyntax.Expression) (tfType, localName string, ok bool) {
	tr, ok := singleTraversal(expr)
	if !ok {
		return "", "", false
	}
	if len(tr) < 2 {
		return "", "", false
	}
	root := tr.RootName()
	if !strings.HasPrefix(root, "keycloak_") {
		return "", "", false
	}
	nameStep, ok := tr[1].(hcl.TraverseAttr)
	if !ok {
		return "", "", false
	}
	return root, nameStep.Name, true
}

// singleTraversal unwraps template wrappers around a single scope traversal so
// that both `keycloak_realm.this.id` and `"${keycloak_realm.this.id}"` are
// recognised.
func singleTraversal(expr hclsyntax.Expression) (hcl.Traversal, bool) {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		return e.Traversal, true
	case *hclsyntax.TemplateWrapExpr:
		return singleTraversal(e.Wrapped)
	case *hclsyntax.TemplateExpr:
		if len(e.Parts) == 1 {
			return singleTraversal(e.Parts[0])
		}
	}
	return nil, false
}

// exprText returns the original source text of an expression, used as a
// placeholder when a value cannot be evaluated statically.
func exprText(expr hclsyntax.Expression, src []byte) string {
	rng := expr.Range()
	start, end := rng.Start.Byte, rng.End.Byte
	if start < 0 || end > len(src) || start > end {
		return ""
	}
	return strings.TrimSpace(string(src[start:end]))
}
