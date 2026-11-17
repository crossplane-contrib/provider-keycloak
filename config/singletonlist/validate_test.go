package singletonlist

import (
	"errors"
	"strings"
	"testing"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testResource() *config.Resource {
	permissions := &schema.Resource{Schema: map[string]*schema.Schema{
		"view": {Type: schema.TypeSet, Elem: &schema.Schema{Type: schema.TypeString}},
	}}
	return &config.Resource{TerraformResource: &schema.Resource{Schema: map[string]*schema.Schema{
		"attribute": {Type: schema.TypeList, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"permissions": {Type: schema.TypeList, MaxItems: 1, Elem: permissions},
		}}},
		"smtp_server": {Type: schema.TypeList, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"token_auth": {Type: schema.TypeList, MaxItems: 1, Elem: permissions},
		}}},
		"view_scope": {Type: schema.TypeSet, MaxItems: 1, Elem: permissions},
	}}}
}

func TestValidate(t *testing.T) {
	cases := map[string]struct {
		params map[string]any
		want   []string
	}{
		"Omitted": {},
		"Empty": {
			params: map[string]any{
				"attribute":  []any{map[string]any{"permissions": []any{}}},
				"view_scope": []any{},
			},
		},
		"Null": {
			params: map[string]any{
				"attribute":  []any{map[string]any{"permissions": nil}},
				"view_scope": nil,
			},
		},
		"Valid": {
			params: map[string]any{
				"attribute":   []any{map[string]any{"permissions": []any{map[string]any{}}}, map[string]any{}},
				"smtp_server": []any{map[string]any{"token_auth": []any{map[string]any{}}}},
				"view_scope":  []any{map[string]any{}},
			},
		},
		"NestedInListBlock": {
			params: map[string]any{"attribute": []any{map[string]any{}, map[string]any{"permissions": []any{map[string]any{}, map[string]any{}}}}},
			want:   []string{"attribute[1].permissions has 2 entries"},
		},
		"NestedSingletons": {
			params: map[string]any{"smtp_server": []any{
				map[string]any{"token_auth": []any{map[string]any{}, map[string]any{}}},
				map[string]any{},
			}},
			want: []string{"smtpServer has 2 entries", "smtpServer[0].tokenAuth has 2 entries"},
		},
		"SetBlock": {
			params: map[string]any{"view_scope": []any{map[string]any{}, map[string]any{}}},
			want:   []string{"viewScope has 2 entries"},
		},
	}

	r := testResource()
	Validate()(r)
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := r.TerraformConfigurationInjector(nil, tc.params)
			if len(tc.want) == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
		})
	}
}

func TestValidateComposesWithInjector(t *testing.T) {
	t.Run("ValidatesInjectedParameters", func(t *testing.T) {
		r := testResource()
		r.TerraformConfigurationInjector = func(jsonMap, tfMap map[string]any) error {
			tfMap["view_scope"] = jsonMap["scopes"]
			return nil
		}
		Validate()(r)
		err := r.TerraformConfigurationInjector(
			map[string]any{"scopes": []any{map[string]any{}, map[string]any{}}},
			map[string]any{},
		)
		if err == nil || !strings.Contains(err.Error(), "viewScope has 2 entries") {
			t.Fatalf("expected injected parameters to be validated, got %v", err)
		}
	})

	t.Run("PreservesInjectorError", func(t *testing.T) {
		want := errors.New("injection failed")
		r := testResource()
		r.TerraformConfigurationInjector = func(_, _ map[string]any) error {
			return want
		}
		Validate()(r)
		if err := r.TerraformConfigurationInjector(nil, nil); !errors.Is(err, want) {
			t.Fatalf("expected the injector error, got %v", err)
		}
	})
}
