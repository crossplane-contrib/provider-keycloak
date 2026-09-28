// Command crdconversion enables the CRD conversion webhook for the generated
// CRDs that serve more than one API version.
//
// controller-gen does not emit a `spec.conversion` stanza, and Crossplane's
// package manager only fills in the caBundle and the service coordinates of a
// conversion webhook that is already declared in the CRD (see
// APIEstablisher.enrichControlledResource in crossplane/crossplane). Without
// this stanza the API server falls back to the `None` conversion strategy,
// which merely rewrites the apiVersion field and would hand out objects whose
// fields do not match the requested schema.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

const conversionStanza = `  conversion:
    strategy: Webhook
    webhook:
      clientConfig:
        service:
          path: /convert
      conversionReviewVersions:
      - v1
`

type crd struct {
	Kind string `json:"kind"`
	Spec struct {
		Conversion map[string]any `json:"conversion"`
		Versions   []struct {
			Name string `json:"name"`
		} `json:"versions"`
	} `json:"spec"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <crd-directory>\n", os.Args[0])
		os.Exit(1)
	}
	entries, err := os.ReadDir(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read the CRD directory: %v\n", err)
		os.Exit(1)
	}
	patched := 0
	permissionsPatched := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		p := filepath.Join(os.Args[1], e.Name())
		ok, err := patch(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot patch %s: %v\n", p, err)
			os.Exit(1)
		}
		if ok {
			patched++
		}
		if e.Name() == "realm.keycloak.crossplane.io_userprofiles.yaml" || e.Name() == "realm.keycloak.m.crossplane.io_userprofiles.yaml" {
			changed, err := patchUserProfilePermissions(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot patch %s: %v\n", p, err)
				os.Exit(1)
			}
			if changed {
				permissionsPatched++
			}
		}
	}
	fmt.Printf("Enabled the conversion webhook for %d multi-version CRD(s)\n", patched)
	fmt.Printf("Added UserProfile permissions limits to %d CRD(s)\n", permissionsPatched)
}

func patch(path string) (bool, error) {
	b, err := os.ReadFile(path) //nolint:gosec // the path is a generated CRD file
	if err != nil {
		return false, err
	}
	c := &crd{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return false, err
	}
	if c.Kind != "CustomResourceDefinition" || len(c.Spec.Versions) < 2 || c.Spec.Conversion != nil {
		return false, nil
	}

	// Insert the stanza right after the top-level "spec:" key instead of
	// re-marshaling the document, so that the rest of the generated file stays
	// byte-identical.
	marker := []byte("\nspec:\n")
	i := bytes.Index(b, marker)
	if i < 0 {
		return false, fmt.Errorf("cannot find the spec key in %s", path)
	}
	var sb strings.Builder
	sb.Write(b[:i+len(marker)])
	sb.WriteString(conversionStanza)
	sb.Write(b[i+len(marker):])
	return true, os.WriteFile(path, []byte(sb.String()), 0o644) //nolint:gosec // generated CRD files are world-readable
}

func patchUserProfilePermissions(path string) (bool, error) {
	b, err := os.ReadFile(path) //nolint:gosec // the path is a generated CRD file
	if err != nil {
		return false, err
	}
	lines := strings.SplitAfter(string(b), "\n")
	changed := false
	for i := 0; i < len(lines); i++ {
		if !strings.HasSuffix(lines[i], "permissions:\n") {
			continue
		}
		indent := len(lines[i]) - len(strings.TrimLeft(lines[i], " "))
		for j := i + 1; j < len(lines); j++ {
			lineIndent := len(lines[j]) - len(strings.TrimLeft(lines[j], " "))
			if lineIndent <= indent && strings.TrimSpace(lines[j]) != "" {
				break
			}
			if strings.TrimSpace(lines[j]) == "type: array" && lineIndent == indent+2 {
				if j > i+1 && strings.TrimSpace(lines[j-1]) == "maxItems: 1" {
					break
				}
				lines[j] = strings.Repeat(" ", lineIndent) + "maxItems: 1\n" + lines[j]
				changed = true
				break
			}
		}
	}
	if !changed {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(strings.Join(lines, "")), 0o644) //nolint:gosec // generated CRD files are world-readable
}
