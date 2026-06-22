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

// Package main implements a post-generation tool that adds SSA list type
// markers to generated *Refs fields whose corresponding Terraform field is
// a set (has +listType=set). This works around an upstream upjet limitation
// where generateReferenceFields does not propagate list merge strategy
// markers from the base field to the generated reference fields.
//
// See: https://github.com/crossplane-contrib/provider-keycloak/issues/594
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// refsFieldPattern matches a *Refs field line (both v1.Reference and
// v1.NamespacedReference).
var refsFieldPattern = regexp.MustCompile(`Refs \[\].*Reference`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <apis-dir>\n", os.Args[0])
		os.Exit(1)
	}

	root := os.Args[1]
	count := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasPrefix(info.Name(), "zz_") || !strings.HasSuffix(info.Name(), "_types.go") {
			return nil
		}

		modified, fixErr := fixFile(path)
		if fixErr != nil {
			return fmt.Errorf("fixing %s: %w", path, fixErr)
		}
		if modified {
			count++
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("postprocess: added listType=map markers to *Refs fields in %d files\n", count)
}

// fixFile processes a single generated types file, adding +listType=map and
// +listMapKey=name markers to *Refs fields that immediately follow a field
// with +listType=set.
func fixFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	modified := false

	for i := 0; i < len(lines); i++ {
		// Look for a +listType=set marker
		if !strings.Contains(lines[i], "// +listType=set") {
			result = append(result, lines[i])
			continue
		}

		// Found +listType=set. Add it to result and scan ahead.
		result = append(result, lines[i])
		i++

		// Skip the field declaration line(s) after +listType=set
		for i < len(lines) && lines[i] != "" && !strings.HasPrefix(strings.TrimSpace(lines[i]), "//") {
			result = append(result, lines[i])
			i++
		}

		// Skip blank lines
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			result = append(result, lines[i])
			i++
		}

		// Check if we're at a "// References to" comment
		if i >= len(lines) || !strings.Contains(lines[i], "// References to") {
			result = append(result, lines[i])
			continue
		}

		// Scan ahead to find the *Refs field, checking if markers already exist
		j := i
		hasListType := false
		refsLine := -1
		for j < len(lines) && j < i+6 {
			if strings.Contains(lines[j], "+listType=map") {
				hasListType = true
				break
			}
			if refsFieldPattern.MatchString(lines[j]) {
				refsLine = j
				break
			}
			j++
		}

		// If markers are already present or no Refs field found, continue normally
		if hasListType || refsLine < 0 {
			result = append(result, lines[i])
			continue
		}

		// Add lines from comment up to (but not including) the Refs field line
		for k := i; k < refsLine; k++ {
			result = append(result, lines[k])
		}

		// Determine indentation from the Refs field line
		indent := ""
		for _, ch := range lines[refsLine] {
			if ch == '\t' || ch == ' ' {
				indent += string(ch)
			} else {
				break
			}
		}

		// Insert the markers
		result = append(result, indent+"// +listType=map")
		result = append(result, indent+"// +listMapKey=name")
		result = append(result, lines[refsLine])
		modified = true
		i = refsLine
	}

	if !modified {
		return false, nil
	}

	err = os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
	return true, err
}
