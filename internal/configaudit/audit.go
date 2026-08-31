// Package configaudit audits the provider configuration for reference wiring
// that is inconsistent, missing or incomplete.
//
// The audit is a cross-check of the Terraform schema against the configuration
// the generator actually builds (config.GetProvider). The schema alone cannot
// tell whether an attribute is a cross-resource reference, but schema plus
// configuration can mechanically tell that an attribute is treated in more than
// one way, or that a reference points at only one member of a set of resource
// types that is elsewhere modelled with config/multitypes.
//
// Three detectors are implemented, see
// design/0001-schema-driven-resource-onboarding.md:
//
//   - DetectorDrift ("drift"): the same attribute, with the same type and the
//     same optionality, is wired differently across resources.
//   - DetectorMissingMultitype ("missing-multitype"): a reference points at one
//     member of a type family while the resource wires none of the other
//     members.
//   - DetectorUnclassified ("unclassified"): a reference-shaped attribute
//     (*_id, *_ids, *_alias) that is neither wired nor recorded as a
//     non-reference.
//
// The audit reports; it never decides. Every finding is a question for a human
// (or for upstream research), not a bug by itself.
package configaudit

import (
	"fmt"
	"sort"
	"strings"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Detector names.
const (
	DetectorDrift            = "drift"
	DetectorMissingMultitype = "missing-multitype"
	DetectorUnclassified     = "unclassified"
)

// Finding classes.
const (
	// ClassGap is an attribute wired to a single target type on some resources
	// and left unwired on others whose schema is identical.
	ClassGap = "gap"
	// ClassMultitype is an attribute name resolving to more than one target
	// type, i.e. the situation config/multitypes exists for.
	ClassMultitype = "multitype"
	// ClassMissingMultitype is a reference to one member of a type family while
	// the other members are not wired on the same resource.
	ClassMissingMultitype = "missing-multitype"
	// ClassUnclassified is a reference-shaped attribute with no reference and
	// no recorded reason for not having one.
	ClassUnclassified = "unclassified"
)

// Finding statuses.
const (
	// StatusOpen marks a finding that still needs a decision.
	StatusOpen = "open"
	// StatusSatisfied marks a finding that is already handled - e.g. an
	// attribute that resolves to several target types because it is configured
	// with config/multitypes on every resource that wires it.
	StatusSatisfied = "satisfied"
)

// Finding is a single audit result. It is deliberately flat so that the JSON
// output stays easy to consume from scripts.
type Finding struct {
	Detector  string `json:"detector"`
	Class     string `json:"class"`
	Status    string `json:"status"`
	Attribute string `json:"attribute"`
	// Resource is set for per-resource findings (missing-multitype,
	// unclassified). Drift findings are per attribute and leave it empty.
	Resource string `json:"resource,omitempty"`
	// Shape is the schema shape drift findings are grouped by: attributes
	// sharing a name but not a type or an optionality are never compared with
	// each other, and are reported separately.
	Shape string `json:"shape,omitempty"`
	// WiredTo maps a reference target to the resources wiring the attribute to
	// it. Only set for drift findings.
	WiredTo map[string][]string `json:"wiredTo,omitempty"`
	// UnwiredOn lists the resources declaring the attribute without wiring it.
	// Only set for drift findings.
	UnwiredOn []string `json:"unwiredOn,omitempty"`
	// Family is a set of reference targets that is elsewhere modelled as a
	// multi-type field. Only set for missing-multitype findings.
	Family []string `json:"family,omitempty"`
	// Missing lists the family members the resource does not wire. Only set for
	// missing-multitype findings.
	Missing []string `json:"missing,omitempty"`
	// ProtocolSpecific marks a missing-multitype finding on a resource whose
	// name is prefixed keycloak_openid_ / keycloak_saml_, i.e. a resource that
	// only ever applies to one protocol. Those are reported but not counted as
	// actionable.
	ProtocolSpecific bool `json:"protocolSpecific,omitempty"`
	// Detail is a human-readable summary of the finding.
	Detail string `json:"detail"`
}

// Key returns a stable identifier for a finding, suitable for deduplicating
// issues filed from repeated runs.
func (f Finding) Key() string {
	if f.Resource != "" {
		return fmt.Sprintf("%s/%s/%s", f.Detector, f.Resource, f.Attribute)
	}
	if f.Shape != "" {
		return fmt.Sprintf("%s/%s/%s", f.Detector, f.Attribute, f.Shape)
	}
	return fmt.Sprintf("%s/%s", f.Detector, f.Attribute)
}

// Report is the result of an audit run.
type Report struct {
	// Resources is the number of configured resources that were audited.
	Resources int `json:"resources"`
	// Findings is sorted by detector, then attribute, then resource.
	Findings []Finding `json:"findings"`
}

// Actionable returns the findings of a detector that still need a decision,
// ignoring satisfied findings and protocol-specific candidates. Multi-type
// classifications are excluded too: they name no resource to change, and the
// resource that is missing a family member is reported per resource by the
// missing-multitype detector - counting both would report the same work twice.
func (r Report) Actionable(detector string) []Finding {
	out := make([]Finding, 0, len(r.Findings))
	for _, f := range r.Findings {
		if f.Detector != detector || f.Status != StatusOpen || f.ProtocolSpecific {
			continue
		}
		if f.Class == ClassMultitype {
			continue
		}
		out = append(out, f)
	}
	return out
}

// attribute is one attribute of one resource, after the configuration has been
// built - so synthetic config/multitypes fields are included.
type attribute struct {
	name     string
	typeKey  string
	required bool
	// target is the Terraform name of the referenced resource, empty when the
	// attribute is not wired as a reference.
	target string
}

// resourceView caches the attributes and the reference targets of a resource.
type resourceView struct {
	name       string
	attributes []attribute
	targets    map[string]bool
}

func (v *resourceView) wires(attr string) bool {
	for _, a := range v.attributes {
		if a.name == attr && a.target != "" {
			return true
		}
	}
	return false
}

// Audit runs all detectors against the built provider configuration.
func Audit(p *ujconfig.Provider) Report {
	views := buildViews(p)
	report := Report{Resources: len(views)}
	report.Findings = append(report.Findings, detectDrift(views)...)
	report.Findings = append(report.Findings, detectMissingMultitype(views)...)
	report.Findings = append(report.Findings, detectUnclassified(views)...)
	sortFindings(report.Findings)
	return report
}

func buildViews(p *ujconfig.Provider) []*resourceView {
	views := make([]*resourceView, 0, len(p.Resources))
	for name, r := range p.Resources {
		if r == nil || r.TerraformResource == nil {
			continue
		}
		v := &resourceView{name: name, targets: map[string]bool{}}
		for k, s := range r.TerraformResource.Schema {
			if s == nil || !isSettable(s) {
				continue
			}
			target := r.References[k].TerraformName
			v.attributes = append(v.attributes, attribute{
				name:     k,
				typeKey:  typeKey(s),
				required: s.Required,
				target:   target,
			})
			if target != "" {
				v.targets[target] = true
			}
		}
		sort.Slice(v.attributes, func(i, j int) bool { return v.attributes[i].name < v.attributes[j].name })
		views = append(views, v)
	}
	sort.Slice(views, func(i, j int) bool { return views[i].name < views[j].name })
	return views
}

// isSettable mirrors the filter config.KnownReferencers applies: status-only
// and sensitive fields never carry a cross-resource reference.
func isSettable(s *schema.Schema) bool {
	if s.Sensitive {
		return false
	}
	return s.Optional || !s.Computed
}

// typeKey renders the attribute type, so that attributes sharing a name but not
// a type are never compared with each other.
func typeKey(s *schema.Schema) string {
	switch e := s.Elem.(type) {
	case *schema.Schema:
		return fmt.Sprintf("%s<%s>", s.Type, e.Type)
	case *schema.Resource:
		return fmt.Sprintf("%s<block>", s.Type)
	default:
		return s.Type.String()
	}
}

// detectDrift reports attributes that are treated in more than one way while
// their schema is identical: same name, same type, same optionality.
func detectDrift(views []*resourceView) []Finding {
	type group struct {
		attribute string
		shape     string
		wiredTo   map[string][]string
		unwiredOn []string
	}
	groups := map[string]*group{}
	for _, v := range views {
		for _, a := range v.attributes {
			shape := shapeOf(a)
			key := a.name + "|" + shape
			g, ok := groups[key]
			if !ok {
				g = &group{attribute: a.name, shape: shape, wiredTo: map[string][]string{}}
				groups[key] = g
			}
			if a.target == "" {
				g.unwiredOn = append(g.unwiredOn, v.name)
				continue
			}
			g.wiredTo[a.target] = append(g.wiredTo[a.target], v.name)
		}
	}

	findings := make([]Finding, 0, len(groups))
	for _, g := range groups {
		// An attribute that is never wired anywhere is not drift; that is what
		// the unclassified detector is for.
		if len(g.wiredTo) == 0 {
			continue
		}
		// A resource that is itself the reference target names the object it
		// configures rather than pointing at another one - keycloak_realm.realm
		// and keycloak_openid_client.client_id are the resource's own
		// identifier, not a dangling reference.
		unwired := make([]string, 0, len(g.unwiredOn))
		for _, name := range g.unwiredOn {
			if _, isTarget := g.wiredTo[name]; !isTarget {
				unwired = append(unwired, name)
			}
		}
		// Consistent: every resource declaring it wires it to the same target.
		if len(g.wiredTo) == 1 && len(unwired) == 0 {
			continue
		}

		f := Finding{
			Detector:  DetectorDrift,
			Attribute: g.attribute,
			Shape:     g.shape,
			Status:    StatusOpen,
			WiredTo:   map[string][]string{},
			UnwiredOn: sorted(unwired),
		}
		for target, resources := range g.wiredTo {
			f.WiredTo[target] = sorted(resources)
		}

		targets := sorted(keys(g.wiredTo))
		if len(targets) > 1 {
			// One attribute name resolving to more than one target type is the
			// multitype signal rather than a gap.
			f.Class = ClassMultitype
			if multitypeSatisfied(views, g.attribute, targets, wiringResources(g.wiredTo)) {
				f.Status = StatusSatisfied
			}
			f.Detail = fmt.Sprintf("%s resolves to %d target types (%s), which is a multi-type field - see config/multitypes",
				g.attribute, len(targets), strings.Join(targets, ", "))
		} else {
			f.Class = ClassGap
			f.Detail = fmt.Sprintf("%s is wired to %s on %d resource(s) and left unwired on %d with an identical schema",
				g.attribute, targets[0], len(g.wiredTo[targets[0]]), len(unwired))
		}
		findings = append(findings, f)
	}
	return findings
}

// shapeOf renders the schema shape an attribute is compared by.
func shapeOf(a attribute) string {
	if a.required {
		return a.typeKey + "/required"
	}
	return a.typeKey + "/optional"
}

// multitypeSatisfied reports whether the multi-type pattern is fully applied
// for an attribute: every protocol-neutral resource wiring it covers the whole
// family. Protocol-specific resources are skipped for the same reason the
// missing-multitype detector skips them - an OpenID-only resource legitimately
// wires only the OpenID member. Only the resources of the finding's own shape
// group are considered: a required and an optional attribute of the same name
// are separate findings and must not decide each other's status.
func multitypeSatisfied(views []*resourceView, attr string, family, group []string) bool {
	inGroup := map[string]bool{}
	for _, name := range group {
		inGroup[name] = true
	}
	for _, v := range views {
		if !inGroup[v.name] || isProtocolSpecific(v.name) || !v.wires(attr) {
			continue
		}
		for _, member := range family {
			if !v.targets[member] {
				return false
			}
		}
	}
	return true
}

// wiringResources flattens a target -> resources map into the list of resources
// wiring the attribute.
func wiringResources(wiredTo map[string][]string) []string {
	var out []string
	for _, resources := range wiredTo {
		out = append(out, resources...)
	}
	return out
}

// detectMissingMultitype reports references pointing at a single member of a
// type family. Families are derived from the configuration itself: any
// attribute already wired to more than one target type defines one.
func detectMissingMultitype(views []*resourceView) []Finding {
	families := deriveFamilies(views)

	var findings []Finding
	for _, v := range views {
		for _, a := range v.attributes {
			if a.target == "" {
				continue
			}
			for _, family := range families {
				if !contains(family, a.target) {
					continue
				}
				missing := make([]string, 0, len(family))
				for _, member := range family {
					if !v.targets[member] {
						missing = append(missing, member)
					}
				}
				if len(missing) == 0 {
					continue
				}
				findings = append(findings, Finding{
					Detector:         DetectorMissingMultitype,
					Class:            ClassMissingMultitype,
					Status:           StatusOpen,
					Resource:         v.name,
					Attribute:        a.name,
					Family:           family,
					Missing:          missing,
					ProtocolSpecific: isProtocolSpecific(v.name),
					Detail: fmt.Sprintf("%s.%s references %s but not %s",
						v.name, a.name, a.target, strings.Join(missing, ", ")),
				})
			}
		}
	}
	return findings
}

// deriveFamilies collects the sets of target types that are already modelled as
// alternatives of each other. There are two ways that shows up in the
// configuration, and both are used:
//
//   - one attribute name resolving to different targets on different resources
//     (keycloak_saml_user_attribute_protocol_mapper wires client_id to
//     keycloak_saml_client, the OpenID mappers wire it to
//     keycloak_openid_client);
//   - a config/multitypes field, whose synthetic instances live on the same
//     resource and are named after the original attribute
//     (client_id + saml_client_id).
//
// Today that yields the OpenID/SAML client pair and the OpenID/SAML client
// scope pair.
func deriveFamilies(views []*resourceView) [][]string {
	byAttribute := map[string]map[string]bool{}
	add := func(key, target string) {
		if byAttribute[key] == nil {
			byAttribute[key] = map[string]bool{}
		}
		byAttribute[key][target] = true
	}

	for _, v := range views {
		for _, a := range v.attributes {
			if a.target == "" {
				continue
			}
			add(a.name, a.target)
			// A synthetic multi-type instance is named "<prefix>_<original>",
			// so it belongs to the family of the attribute it varies.
			for _, other := range v.attributes {
				if other.target == "" || other.name == a.name {
					continue
				}
				if strings.HasSuffix(other.name, "_"+a.name) {
					add(a.name, other.target)
				}
			}
		}
	}

	seen := map[string]bool{}
	var families [][]string
	for _, targets := range byAttribute {
		if len(targets) < 2 {
			continue
		}
		family := sorted(keys(targets))
		key := strings.Join(family, ",")
		if seen[key] {
			continue
		}
		seen[key] = true
		families = append(families, family)
	}
	sort.Slice(families, func(i, j int) bool { return strings.Join(families[i], ",") < strings.Join(families[j], ",") })
	return families
}

// isProtocolSpecific reports whether a resource only ever applies to one client
// protocol, which its name states: an OpenID protocol mapper cannot attach to a
// SAML client. Such resources are reported but not counted as actionable
// missing-multitype candidates.
func isProtocolSpecific(resource string) bool {
	return strings.HasPrefix(resource, "keycloak_openid_") || strings.HasPrefix(resource, "keycloak_saml_")
}

// referenceShapedSuffixes are the attribute name shapes that usually denote a
// reference to another Keycloak object.
var referenceShapedSuffixes = []string{"_id", "_ids", "_alias"}

// detectUnclassified reports reference-shaped attributes that are not wired.
// Roughly a third of these are correct omissions (provider_id, tenant_id,
// entity_id, ...), which is why this detector reports rather than gates until
// there is a place to record the reason.
func detectUnclassified(views []*resourceView) []Finding {
	var findings []Finding
	for _, v := range views {
		for _, a := range v.attributes {
			if a.target != "" || !isReferenceShaped(a.name) {
				continue
			}
			findings = append(findings, Finding{
				Detector:  DetectorUnclassified,
				Class:     ClassUnclassified,
				Status:    StatusOpen,
				Resource:  v.name,
				Attribute: a.name,
				Detail: fmt.Sprintf("%s.%s looks like a reference (%s) but is not wired and has no recorded reason",
					v.name, a.name, a.typeKey),
			})
		}
	}
	return findings
}

func isReferenceShaped(name string) bool {
	for _, suffix := range referenceShapedSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Detector != b.Detector {
			return a.Detector < b.Detector
		}
		if a.Attribute != b.Attribute {
			return a.Attribute < b.Attribute
		}
		return a.Resource < b.Resource
	})
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
