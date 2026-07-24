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

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// crd is the top-level Crossplane manifest envelope. Field ordering here is
// preserved by yaml.v3, matching the style used in examples/.
type crd struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Metadata   objectMeta `yaml:"metadata"`
	Spec       crdSpec    `yaml:"spec"`
}

type objectMeta struct {
	Name string `yaml:"name"`
}

type crdSpec struct {
	ForProvider        map[string]any `yaml:"forProvider"`
	ProviderConfigRef  *localRef      `yaml:"providerConfigRef,omitempty"`
	DeletionPolicy     string         `yaml:"deletionPolicy,omitempty"`
	ManagementPolicies []string       `yaml:"managementPolicies,omitempty"`
}

type localRef struct {
	Name string `yaml:"name"`
}

// marshalYAML renders a value as YAML with a 2-space indent, trimming the
// trailing newline so callers control document separators.
func marshalYAML(v any) (string, error) {
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return "", errors.Wrap(err, "cannot encode YAML")
	}
	if err := enc.Close(); err != nil {
		return "", errors.Wrap(err, "cannot flush YAML encoder")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

// Render joins all documents in a Result into a single multi-document YAML
// stream separated by `---`.
func (r *Result) Render() string {
	parts := make([]string, 0, len(r.Documents))
	for _, d := range r.Documents {
		parts = append(parts, d.Manifest)
	}
	return strings.Join(parts, "\n---\n") + "\n"
}
