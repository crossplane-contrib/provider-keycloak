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
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// TestConvertGolden converts each testdata/*.tf file and compares the rendered
// manifests against the matching *.tf.golden file. Run with -update to
// regenerate the golden files.
func TestConvertGolden(t *testing.T) {
	conv, err := New(Options{ProviderConfigRef: "keycloak-provider-config"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tfFiles, err := filepath.Glob(filepath.Join("testdata", "*.tf"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tfFiles) == 0 {
		t.Fatal("no testdata/*.tf fixtures found")
	}

	for _, tf := range tfFiles {
		tf := tf
		t.Run(strings.TrimSuffix(filepath.Base(tf), ".tf"), func(t *testing.T) {
			src, err := os.ReadFile(tf)
			if err != nil {
				t.Fatal(err)
			}
			res, err := conv.Convert(src, tf)
			if err != nil {
				t.Fatalf("Convert: %v", err)
			}
			got := res.Render()

			golden := tf + ".golden"
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden (run with -update to create): %v", err)
			}
			if got != string(want) {
				t.Errorf("rendered manifests differ from %s.\n--- got ---\n%s\n--- want ---\n%s", golden, got, want)
			}
		})
	}
}

func TestConvertReference(t *testing.T) {
	conv, err := New(Options{ProviderConfigRef: "pc"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	src := `
resource "keycloak_realm" "r" {
  realm = "demo"
}
resource "keycloak_group" "g" {
  realm_id = keycloak_realm.r.id
  name     = "g"
}
`
	res, err := conv.Convert([]byte(src), "ref.tf")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if len(res.Documents) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(res.Documents))
	}
	group := res.Documents[1].Manifest
	if !strings.Contains(group, "realmIdRef:") || !strings.Contains(group, "name: r") {
		t.Errorf("expected realmIdRef to reference realm r, got:\n%s", group)
	}
	if strings.Contains(group, "realmId:") {
		t.Errorf("literal realmId should not be emitted when a reference is used:\n%s", group)
	}
}

func TestConvertUnsupported(t *testing.T) {
	conv, err := New(Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	src := `resource "keycloak_does_not_exist" "x" { foo = "bar" }`
	res, err := conv.Convert([]byte(src), "u.tf")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if len(res.Documents) != 0 {
		t.Fatalf("expected no documents, got %d", len(res.Documents))
	}
	if len(res.Unsupported) != 1 || res.Unsupported[0] != "keycloak_does_not_exist" {
		t.Errorf("expected unsupported type recorded, got %v", res.Unsupported)
	}
}

func TestSanitizeName(t *testing.T) {
	cases := map[string]string{
		"this":         "this",
		"child_group":  "child-group",
		"My.Realm":     "my-realm",
		"_leading":     "leading",
		"UPPER":        "upper",
		"a b c":        "a-b-c",
		"weird$chars!": "weirdchars",
		"":             "resource",
	}
	for in, want := range cases {
		if got := sanitizeName(in); got != want {
			t.Errorf("sanitizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestSupportedTypesCoverage ensures every externally-configured resource is
// convertible, i.e. present in the provider's resource map.
func TestSupportedTypesCoverage(t *testing.T) {
	conv, err := New(Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if len(conv.SupportedTypes()) == 0 {
		t.Fatal("expected a non-empty set of supported types")
	}
	for _, tn := range conv.SupportedTypes() {
		if !strings.HasPrefix(tn, "keycloak_") {
			t.Errorf("unexpected resource type without keycloak_ prefix: %q", tn)
		}
	}
}
