package config

import (
	"context"
	"regexp"
	"testing"
)

var deterministicExternalNameRe = regexp.MustCompile(`eq \.ResourceType "([^"]+)"`)

// TestDeterministicExternalNameResources guards the resources for which the
// controller template sets managed.WithDeterministicExternalName(true) (issue
// #815). That option lets the reconciler proceed after a create whose result
// could not be recorded, which is only safe if the Terraform ID can be derived
// from spec alone. Otherwise a second Create could leak a Keycloak object.
func TestDeterministicExternalNameResources(t *testing.T) {
	matches := deterministicExternalNameRe.FindAllStringSubmatch(controllerTemplate, -1)
	if len(matches) == 0 {
		t.Fatal("controller template opts no resource into WithDeterministicExternalName")
	}

	ctx := context.Background()
	for _, m := range matches {
		tfName := m[1]
		t.Run(tfName, func(t *testing.T) {
			e, ok := ExternalNameConfigs[tfName]
			if !ok {
				t.Fatalf("%s: no external name config registered", tfName)
			}
			params := map[string]any{"realm_id": "my-realm"}
			// Before creation the external name is empty; afterwards it is
			// whatever was persisted. Both must yield the same Terraform ID.
			before, err := e.GetIDFn(ctx, "", params, nil)
			if err != nil {
				t.Fatalf("GetIDFn with empty external name: %v", err)
			}
			after, err := e.GetIDFn(ctx, "something-else", params, nil)
			if err != nil {
				t.Fatalf("GetIDFn with external name: %v", err)
			}
			if before == "" || before != after {
				t.Fatalf("Terraform ID is not derived from spec: %q (no external name) vs %q (with external name)", before, after)
			}
			name, err := e.GetExternalNameFn(map[string]any{"id": before})
			if err != nil {
				t.Fatalf("GetExternalNameFn: %v", err)
			}
			if name != before {
				t.Errorf("external name %q does not round-trip Terraform ID %q", name, before)
			}
		})
	}
}

func TestDefaultGroupsExternalName(t *testing.T) {
	getID := ExternalNameConfigs["keycloak_default_groups"].GetIDFn
	cases := map[string]struct {
		externalName string
		params       map[string]any
		want         string
	}{
		"FromRealmID":         {params: map[string]any{"realm_id": "my-realm"}, want: "my-realm/default-groups"},
		"RealmIDWins":         {externalName: "old/default-groups", params: map[string]any{"realm_id": "my-realm"}, want: "my-realm/default-groups"},
		"RealmInternalIDOnly": {externalName: "my-realm/default-groups", params: map[string]any{"realm_internal_id": "uuid"}, want: "my-realm/default-groups"},
		"NothingKnownYet":     {params: map[string]any{"realm_internal_id": "uuid"}, want: ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := getID(context.Background(), tc.externalName, tc.params, nil)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
