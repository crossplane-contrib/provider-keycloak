/*
Copyright 2021 Upbound Inc.
*/

// Command generatedlist writes config/generated.lst, the list of Terraform
// resources exposed as managed resources by this provider. The source of truth
// is config.ExternalNameConfigs: a resource is only generated when it has an
// external name configuration.
//
// Usage:
//
//	generatedlist <path>            # write the list to <path>
//	generatedlist --check <path>    # exit non-zero when <path> is stale
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/crossplane-contrib/provider-keycloak/config"
)

const header = `# Terraform resources exposed as managed resources by this provider.
#
# GENERATED FILE - DO NOT EDIT. Regenerate with 'make generated-lst'
# (also run as part of 'make generate'). The source of truth is
# config.ExternalNameConfigs in config/external_name.go.
`

func generatedList() string {
	names := make([]string, 0, len(config.ExternalNameConfigs))
	for name := range config.ExternalNameConfigs {
		names = append(names, name)
	}
	sort.Strings(names)
	return header + strings.Join(names, "\n") + "\n"
}

func main() {
	args := os.Args[1:]
	check := false
	if len(args) > 0 && args[0] == "--check" {
		check = true
		args = args[1:]
	}
	if len(args) != 1 || args[0] == "" {
		fmt.Fprintln(os.Stderr, "usage: generatedlist [--check] <path to generated.lst>")
		os.Exit(2)
	}
	path := args[0]
	want := generatedList()

	if check {
		got, err := os.ReadFile(path) //nolint:gosec // path is provided by the build, not by user input.
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot read %s: %s\n", path, err)
			os.Exit(2)
		}
		if string(got) != want {
			fmt.Fprintf(os.Stderr, "%s is out of date with config.ExternalNameConfigs. Run 'make generated-lst' and commit the result.\n", path)
			os.Exit(1)
		}
		return
	}

	if err := os.WriteFile(path, []byte(want), 0o644); err != nil { //nolint:gosec // the list is not sensitive.
		fmt.Fprintf(os.Stderr, "cannot write %s: %s\n", path, err)
		os.Exit(2)
	}
}
