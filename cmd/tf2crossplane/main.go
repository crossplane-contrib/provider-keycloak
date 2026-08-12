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

// Command tf2crossplane converts Terraform HCL that uses the Keycloak
// Terraform provider into Crossplane Managed Resource manifests for
// provider-keycloak.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kingpin/v2"

	"github.com/crossplane-contrib/provider-keycloak/internal/tf2crossplane"
)

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "Convert Terraform HCL for the Keycloak provider into Crossplane manifests.")
	var (
		inputs             = app.Arg("inputs", "Terraform .tf files or directories to convert. Reads stdin when omitted.").Strings()
		output             = app.Flag("output", "Write the rendered manifests to this file instead of stdout.").Short('o').String()
		namespaced         = app.Flag("namespaced", "Emit namespaced (keycloak.m.crossplane.io) resources.").Bool()
		providerConfigRef  = app.Flag("provider-config", "Name written into spec.providerConfigRef.name.").Default("keycloak-provider-config").String()
		deletionPolicy     = app.Flag("deletion-policy", "Value written into spec.deletionPolicy (e.g. Delete or Orphan).").String()
		managementPolicies = app.Flag("management-policies", "Comma-separated list written into spec.managementPolicies.").String()
		listSupported      = app.Flag("list-supported", "Print the Terraform resource types the converter can map and exit.").Bool()
		quiet              = app.Flag("quiet", "Suppress warnings on stderr.").Short('q').Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	opts := tf2crossplane.Options{
		Namespaced:        *namespaced,
		ProviderConfigRef: *providerConfigRef,
		DeletionPolicy:    *deletionPolicy,
	}
	if *managementPolicies != "" {
		for _, p := range strings.Split(*managementPolicies, ",") {
			if p = strings.TrimSpace(p); p != "" {
				opts.ManagementPolicies = append(opts.ManagementPolicies, p)
			}
		}
	}

	conv, err := tf2crossplane.New(opts)
	if err != nil {
		fatalf("cannot initialise converter: %v", err)
	}

	if *listSupported {
		for _, t := range conv.SupportedTypes() {
			fmt.Println(t)
		}
		return
	}

	sources, err := gatherSources(*inputs)
	if err != nil {
		fatalf("%v", err)
	}

	var (
		docs        []string
		warnings    []string
		unsupported []string
	)
	for _, s := range sources {
		res, err := conv.Convert(s.data, s.name)
		if err != nil {
			fatalf("%s: %v", s.name, err)
		}
		for _, d := range res.Documents {
			docs = append(docs, d.Manifest)
		}
		warnings = append(warnings, prefixWarnings(s.name, res.Warnings)...)
		unsupported = append(unsupported, res.Unsupported...)
	}

	out := strings.Join(docs, "\n---\n")
	if len(docs) > 0 {
		out += "\n"
	}
	if err := writeOutput(*output, out); err != nil {
		fatalf("%v", err)
	}

	if !*quiet {
		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, "warning: "+w)
		}
		if u := dedupeSorted(unsupported); len(u) > 0 {
			fmt.Fprintf(os.Stderr, "warning: %d Terraform resource type(s) have no CRD and were skipped: %s\n", len(u), strings.Join(u, ", "))
		}
	}
}

type source struct {
	name string
	data []byte
}

func gatherSources(inputs []string) ([]source, error) {
	if len(inputs) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("cannot read stdin: %w", err)
		}
		return []source{{name: "<stdin>", data: data}}, nil
	}

	var files []string
	for _, in := range inputs {
		info, err := os.Stat(in)
		if err != nil {
			return nil, fmt.Errorf("cannot access %q: %w", in, err)
		}
		if info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(in, "*.tf"))
			if err != nil {
				return nil, err
			}
			files = append(files, matches...)
			continue
		}
		files = append(files, in)
	}
	sort.Strings(files)

	sources := make([]source, 0, len(files))
	for _, f := range files {
		data, err := os.ReadFile(f) //nolint:gosec // Reading user-specified Terraform files is the tool's purpose.
		if err != nil {
			return nil, fmt.Errorf("cannot read %q: %w", f, err)
		}
		sources = append(sources, source{name: f, data: data})
	}
	return sources, nil
}

func writeOutput(path, content string) error {
	if path == "" {
		_, err := os.Stdout.WriteString(content)
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil { //nolint:gosec // Generated manifests are not secrets.
		return fmt.Errorf("cannot write %q: %w", path, err)
	}
	return nil
}

func prefixWarnings(name string, warnings []string) []string {
	out := make([]string, len(warnings))
	for i, w := range warnings {
		out[i] = name + ": " + w
	}
	return out
}

func dedupeSorted(in []string) []string {
	sort.Strings(in)
	out := in[:0]
	var last string
	for i, s := range in {
		if i == 0 || s != last {
			out = append(out, s)
			last = s
		}
	}
	return out
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
