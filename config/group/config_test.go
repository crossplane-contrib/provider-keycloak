package group

import (
	"testing"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestFindGroupByNameAndParent_TopLevelGroup(t *testing.T) {
	groups := []*keycloak.Group{
		{Id: "g1", Name: "group-a"},
		{Id: "g2", Name: "group-b"},
	}

	found := findGroupByNameAndParent("group-a", "", groups, "")
	if found == nil {
		t.Fatal("expected to find group-a, got nil")
	}
	if found.Id != "g1" {
		t.Errorf("expected id g1, got %s", found.Id)
	}
}

func TestFindGroupByNameAndParent_TopLevelGroupNotFound(t *testing.T) {
	groups := []*keycloak.Group{
		{Id: "g1", Name: "group-a"},
	}

	found := findGroupByNameAndParent("nonexistent", "", groups, "")
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
}

func TestFindGroupByNameAndParent_ChildGroup(t *testing.T) {
	groups := []*keycloak.Group{
		{
			Id:   "parent-1",
			Name: "parent-1",
			SubGroups: []*keycloak.Group{
				{Id: "child-1", Name: "test-child"},
			},
		},
		{
			Id:   "parent-2",
			Name: "parent-2",
			SubGroups: []*keycloak.Group{
				{Id: "child-2", Name: "test-child"},
			},
		},
	}

	// Find test-child under parent-1
	found := findGroupByNameAndParent("test-child", "parent-1", groups, "")
	if found == nil {
		t.Fatal("expected to find test-child under parent-1, got nil")
	}
	if found.Id != "child-1" {
		t.Errorf("expected id child-1, got %s", found.Id)
	}

	// Find test-child under parent-2
	found = findGroupByNameAndParent("test-child", "parent-2", groups, "")
	if found == nil {
		t.Fatal("expected to find test-child under parent-2, got nil")
	}
	if found.Id != "child-2" {
		t.Errorf("expected id child-2, got %s", found.Id)
	}
}

func TestFindGroupByNameAndParent_ChildGroupNotFoundWrongParent(t *testing.T) {
	groups := []*keycloak.Group{
		{
			Id:   "parent-1",
			Name: "parent-1",
			SubGroups: []*keycloak.Group{
				{Id: "child-1", Name: "test-child"},
			},
		},
	}

	found := findGroupByNameAndParent("test-child", "nonexistent-parent", groups, "")
	if found != nil {
		t.Errorf("expected nil when parent doesn't match, got %+v", found)
	}
}

func TestFindGroupByNameAndParent_DoesNotReturnChildAsTopLevel(t *testing.T) {
	groups := []*keycloak.Group{
		{
			Id:   "parent-1",
			Name: "parent-1",
			SubGroups: []*keycloak.Group{
				{Id: "child-1", Name: "test-child"},
			},
		},
	}

	// Looking for a top-level group named "test-child" (parentID="")
	// Should NOT find the child group under parent-1
	found := findGroupByNameAndParent("test-child", "", groups, "")
	if found != nil {
		t.Errorf("expected nil for top-level search when group only exists as child, got %+v", found)
	}
}

func TestFindGroupByNameAndParent_DeeplyNestedGroup(t *testing.T) {
	groups := []*keycloak.Group{
		{
			Id:   "grandparent",
			Name: "grandparent",
			SubGroups: []*keycloak.Group{
				{
					Id:   "parent",
					Name: "parent",
					SubGroups: []*keycloak.Group{
						{Id: "child", Name: "deep-child"},
					},
				},
			},
		},
	}

	found := findGroupByNameAndParent("deep-child", "parent", groups, "")
	if found == nil {
		t.Fatal("expected to find deep-child under parent, got nil")
	}
	if found.Id != "child" {
		t.Errorf("expected id child, got %s", found.Id)
	}
}

func TestFindGroupByNameAndParent_EmptyGroups(t *testing.T) {
	found := findGroupByNameAndParent("any", "", []*keycloak.Group{}, "")
	if found != nil {
		t.Errorf("expected nil for empty groups, got %+v", found)
	}
}

func TestFindGroupByNameAndParent_SameNameDifferentLevels(t *testing.T) {
	// Group "test" exists both as top-level and as child of parent-1
	groups := []*keycloak.Group{
		{Id: "top-test", Name: "test"},
		{
			Id:   "parent-1",
			Name: "parent-1",
			SubGroups: []*keycloak.Group{
				{Id: "child-test", Name: "test"},
			},
		},
	}

	// Looking for top-level "test"
	found := findGroupByNameAndParent("test", "", groups, "")
	if found == nil {
		t.Fatal("expected to find top-level test, got nil")
	}
	if found.Id != "top-test" {
		t.Errorf("expected id top-test, got %s", found.Id)
	}

	// Looking for "test" under parent-1
	found = findGroupByNameAndParent("test", "parent-1", groups, "")
	if found == nil {
		t.Fatal("expected to find test under parent-1, got nil")
	}
	if found.Id != "child-test" {
		t.Errorf("expected id child-test, got %s", found.Id)
	}
}
