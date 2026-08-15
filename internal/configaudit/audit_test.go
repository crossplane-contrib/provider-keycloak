package configaudit

import (
	"strings"
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resource builds a Terraform resource with optional string attributes, where
// attrs maps an attribute name to its reference target ("" when unwired).
func resource(attrs map[string]string) *ujconfig.Resource {
	r := &ujconfig.Resource{
		TerraformResource: &schema.Resource{Schema: map[string]*schema.Schema{}},
		References:        map[string]ujconfig.Reference{},
	}
	for name, target := range attrs {
		r.TerraformResource.Schema[name] = &schema.Schema{Type: schema.TypeString, Optional: true}
		if target != "" {
			r.References[name] = ujconfig.Reference{TerraformName: target}
		}
	}
	return r
}

func provider(resources map[string]*ujconfig.Resource) *ujconfig.Provider {
	return &ujconfig.Provider{Resources: resources}
}

func findingsOf(r Report, detector string) []Finding {
	var out []Finding
	for _, f := range r.Findings {
		if f.Detector == detector {
			out = append(out, f)
		}
	}
	return out
}

func TestDriftReportsGap(t *testing.T) {
	report := Audit(provider(map[string]*ujconfig.Resource{
		"keycloak_a_identity_provider": resource(map[string]string{"first_broker_login_flow_alias": "keycloak_authentication_flow"}),
		"keycloak_b_identity_provider": resource(map[string]string{"first_broker_login_flow_alias": ""}),
		"keycloak_authentication_flow": resource(map[string]string{"alias": ""}),
	}))

	drift := findingsOf(report, DetectorDrift)
	if len(drift) != 1 {
		t.Fatalf("got %d drift findings, want 1: %+v", len(drift), drift)
	}
	f := drift[0]
	if f.Class != ClassGap || f.Status != StatusOpen {
		t.Errorf("got class %q status %q, want %q/%q", f.Class, f.Status, ClassGap, StatusOpen)
	}
	if len(f.UnwiredOn) != 1 || f.UnwiredOn[0] != "keycloak_b_identity_provider" {
		t.Errorf("got unwiredOn %v, want [keycloak_b_identity_provider]", f.UnwiredOn)
	}
}

func TestDriftIgnoresConsistentWiring(t *testing.T) {
	report := Audit(provider(map[string]*ujconfig.Resource{
		"keycloak_a": resource(map[string]string{"realm_id": "keycloak_realm"}),
		"keycloak_b": resource(map[string]string{"realm_id": "keycloak_realm"}),
	}))
	if drift := findingsOf(report, DetectorDrift); len(drift) != 0 {
		t.Errorf("got %d drift findings, want 0: %+v", len(drift), drift)
	}
}

func TestDriftIgnoresDifferingSchemaShapes(t *testing.T) {
	required := resource(map[string]string{"client_id": ""})
	required.TerraformResource.Schema["client_id"] = &schema.Schema{Type: schema.TypeString, Required: true}

	report := Audit(provider(map[string]*ujconfig.Resource{
		"keycloak_a":             resource(map[string]string{"client_id": "keycloak_openid_client"}),
		"keycloak_b":             required,
		"keycloak_openid_client": resource(map[string]string{"name": ""}),
	}))
	// The unwired attribute has a different optionality, so it is not compared
	// with the wired one and only the unclassified detector reports it.
	if drift := findingsOf(report, DetectorDrift); len(drift) != 0 {
		t.Errorf("got %d drift findings, want 0: %+v", len(drift), drift)
	}
}

func TestDriftIgnoresSelfNamingAttribute(t *testing.T) {
	report := Audit(provider(map[string]*ujconfig.Resource{
		"keycloak_a":     resource(map[string]string{"realm": "keycloak_realm"}),
		"keycloak_realm": resource(map[string]string{"realm": ""}),
	}))
	if drift := findingsOf(report, DetectorDrift); len(drift) != 0 {
		t.Errorf("got %d drift findings, want 0: %+v", len(drift), drift)
	}
}

func TestDriftClassifiesTwoTargetsAsMultitype(t *testing.T) {
	report := Audit(provider(map[string]*ujconfig.Resource{
		// Wires the whole family, so the multi-type pattern is applied.
		"keycloak_generic_mapper": resource(map[string]string{
			"client_id":      "keycloak_openid_client",
			"saml_client_id": "keycloak_saml_client",
		}),
		"keycloak_other_mapper": resource(map[string]string{"client_id": "keycloak_saml_client"}),
	}))

	drift := findingsOf(report, DetectorDrift)
	if len(drift) != 1 {
		t.Fatalf("got %d drift findings, want 1: %+v", len(drift), drift)
	}
	if drift[0].Class != ClassMultitype {
		t.Errorf("got class %q, want %q", drift[0].Class, ClassMultitype)
	}
}

func TestMissingMultitypeReportsIncompleteFamily(t *testing.T) {
	report := Audit(provider(map[string]*ujconfig.Resource{
		// Defines the family: client_id resolves to two target types.
		"keycloak_generic_mapper": resource(map[string]string{
			"client_id":      "keycloak_openid_client",
			"saml_client_id": "keycloak_saml_client",
		}),
		// Protocol-neutral and incomplete: the actionable finding.
		"keycloak_ldap_role_mapper": resource(map[string]string{"client_id": "keycloak_openid_client"}),
		// Protocol-specific: reported, but not actionable.
		"keycloak_openid_audience_protocol_mapper": resource(map[string]string{"client_id": "keycloak_openid_client"}),
	}))

	actionable := report.Actionable(DetectorMissingMultitype)
	if len(actionable) != 1 {
		t.Fatalf("got %d actionable missing-multitype findings, want 1: %+v", len(actionable), actionable)
	}
	f := actionable[0]
	if f.Resource != "keycloak_ldap_role_mapper" || f.Attribute != "client_id" {
		t.Errorf("got finding for %s.%s, want keycloak_ldap_role_mapper.client_id", f.Resource, f.Attribute)
	}
	if len(f.Missing) != 1 || f.Missing[0] != "keycloak_saml_client" {
		t.Errorf("got missing %v, want [keycloak_saml_client]", f.Missing)
	}

	all := findingsOf(report, DetectorMissingMultitype)
	if len(all) != 2 {
		t.Errorf("got %d missing-multitype findings, want 2 (one of them protocol-specific): %+v", len(all), all)
	}
}

func TestMissingMultitypeNeedsAPrecedent(t *testing.T) {
	// No attribute resolves to more than one target, so no family exists and
	// nothing can be reported - the detector generalises from a precedent only.
	report := Audit(provider(map[string]*ujconfig.Resource{
		"keycloak_ldap_role_mapper": resource(map[string]string{"client_id": "keycloak_openid_client"}),
		"keycloak_saml_client":      resource(map[string]string{"name": ""}),
	}))
	if all := findingsOf(report, DetectorMissingMultitype); len(all) != 0 {
		t.Errorf("got %d missing-multitype findings, want 0: %+v", len(all), all)
	}
}

func TestUnclassifiedOnlyReportsReferenceShapedAttributes(t *testing.T) {
	report := Audit(provider(map[string]*ujconfig.Resource{
		"keycloak_a": resource(map[string]string{
			"provider_id": "",
			"role_ids":    "",
			"some_alias":  "",
			"description": "",
			"realm_id":    "keycloak_realm",
		}),
		"keycloak_realm": resource(map[string]string{"realm": ""}),
	}))

	got := map[string]bool{}
	for _, f := range findingsOf(report, DetectorUnclassified) {
		got[f.Attribute] = true
	}
	for _, want := range []string{"provider_id", "role_ids", "some_alias"} {
		if !got[want] {
			t.Errorf("missing unclassified finding for %q", want)
		}
	}
	for _, unwanted := range []string{"description", "realm_id"} {
		if got[unwanted] {
			t.Errorf("unexpected unclassified finding for %q", unwanted)
		}
	}
}

func TestAuditSkipsStatusAndSensitiveAttributes(t *testing.T) {
	r := resource(map[string]string{})
	r.TerraformResource.Schema["computed_id"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	r.TerraformResource.Schema["secret_id"] = &schema.Schema{Type: schema.TypeString, Optional: true, Sensitive: true}

	report := Audit(provider(map[string]*ujconfig.Resource{"keycloak_a": r}))
	if len(report.Findings) != 0 {
		t.Errorf("got %d findings, want 0: %+v", len(report.Findings), report.Findings)
	}
}

func TestFindingKeysAreUnique(t *testing.T) {
	optional := resource(map[string]string{"client_id": "keycloak_openid_client"})
	required := resource(map[string]string{"client_id": ""})
	required.TerraformResource.Schema["client_id"] = &schema.Schema{Type: schema.TypeString, Required: true}
	requiredWired := resource(map[string]string{"client_id": ""})
	requiredWired.TerraformResource.Schema["client_id"] = &schema.Schema{Type: schema.TypeString, Required: true}
	requiredWired.References["client_id"] = ujconfig.Reference{TerraformName: "keycloak_openid_client"}

	report := Audit(provider(map[string]*ujconfig.Resource{
		"keycloak_a": optional,
		"keycloak_b": resource(map[string]string{"client_id": ""}),
		"keycloak_c": required,
		"keycloak_d": requiredWired,
	}))

	seen := map[string]bool{}
	for _, f := range report.Findings {
		if seen[f.Key()] {
			t.Errorf("duplicate finding key %q", f.Key())
		}
		seen[f.Key()] = true
	}
	if len(findingsOf(report, DetectorDrift)) != 2 {
		t.Errorf("want one drift finding per schema shape, got %+v", findingsOf(report, DetectorDrift))
	}
}

func TestWriteTableHidesSatisfiedFindings(t *testing.T) {
	report := Report{
		Resources: 2,
		Findings: []Finding{
			{Detector: DetectorDrift, Class: ClassGap, Status: StatusOpen, Attribute: "a", Detail: "open finding"},
			{Detector: DetectorDrift, Class: ClassMultitype, Status: StatusSatisfied, Attribute: "b", Detail: "satisfied finding"},
		},
	}

	var out strings.Builder
	if err := WriteTable(&out, report, false); err != nil {
		t.Fatalf("WriteTable: %v", err)
	}
	if !strings.Contains(out.String(), "open finding") {
		t.Errorf("open finding missing from table output:\n%s", out.String())
	}
	if strings.Contains(out.String(), "satisfied finding") {
		t.Errorf("satisfied finding should be hidden by default:\n%s", out.String())
	}

	out.Reset()
	if err := WriteTable(&out, report, true); err != nil {
		t.Fatalf("WriteTable: %v", err)
	}
	if !strings.Contains(out.String(), "satisfied finding") {
		t.Errorf("satisfied finding missing from --show-all output:\n%s", out.String())
	}
}
