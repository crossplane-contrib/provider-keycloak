# 0001 – Schema-Driven Resource Onboarding

- **Status:** Draft
- **Issue:** [#712](https://github.com/crossplane-contrib/provider-keycloak/issues/712)
- **Scope:** `config/`, `generate/`, `Makefile`, `scripts/`

## Problem

Exposing a Terraform resource as a managed resource today means touching a
handful of places by hand, in the right order, with knowledge that only lives in
`AGENTS.md` and in the heads of a few maintainers:

1. an entry in `config/external_name.go`,
2. references (and sometimes `config/multitypes`) in `config/<group>/config.go`,
3. optionally an identifying-properties lookup so a pre-existing Keycloak object
   is imported instead of causing a `409`,
4. `examples/<group>/<resource>.yaml`,
5. an e2e demo plus a `cluster/test/cases*.txt` entry,
6. optionally a docs page.

Steps 4–6 are already gated by CI (`make e2e-cases-check`,
`make generated-lst-check`, `make docs-freshness-check`), so they are hard to
forget. Steps 1–3 are not gated at all: nothing notices when a reference-shaped
attribute is left as a raw ID string, and nothing notices when two sibling
resources are configured inconsistently.

That is not theoretical. Measured against the current tree (109 Terraform
resources, all 109 exposed):

- 312 non-computed, non-sensitive attributes look like references
  (`*_id`, `*_ids`, `*_alias`); 281 are wired, 31 are not.
- Most of the 31 are *correctly* unwired: `provider_id`, `tenant_id`,
  `entity_id`, `key_alias`, and the OIDC/SAML `client_id` of a client are plain
  values, not references to another managed resource.
- But `post_broker_login_flow_alias` is wired on
  `keycloak_spiffe_identity_provider` and
  `keycloak_oidc_openshift_v4_identity_provider` and silently missing on
  `keycloak_oidc_identity_provider`, the four social IdPs,
  `keycloak_saml_identity_provider` and `keycloak_kubernetes_identity_provider`
  — even though `first_broker_login_flow_alias` is wired on all of them.

So the real problem is not "too much typing". It is that **the configuration has
no completeness gate**, which makes drift invisible and makes onboarding a
resource an exercise in imitation.

## Goals

- Onboarding a resource should be one command plus a small number of decisions
  that genuinely require judgement.
- Every decision that is *not* mechanical must be explicit in the tree and
  reviewable in a diff — never inferred silently at generation time.
- Inconsistency between sibling resources must fail CI, not survive unnoticed.
- Equally usable by a human and by a cheap LLM: deterministic commands,
  deterministic output, explicit `TODO(...)` markers.
- Low maintenance: no new DSL, no framework, no second configuration language.

## Non-Goals

- Removing `config/external_name.go`. The external-name identifier depends on
  Keycloak's REST semantics, which the Terraform schema does not describe. Until
  Keycloak ships an OpenAPI spec, this stays a human decision.
- Auto-merging generated configuration into hand-written Go. Generated code must
  land in files that are clearly generated, or be scaffolded once and then owned
  by a human.
- Auto-generating e2e demos with realistic values. Scaffolding a skeleton is in
  scope; inventing semantics is not.

## Proposed Design

Three independent pieces. Each is useful alone, and each can be adopted without
the others.

### A. A declarative attribute registry (replaces "imitate the neighbour")

`config/provider.go` already has the right idea in `KnownReferencers()`: a
`switch` over attribute names that wires `realm_id`, `organization_id`,
`client_scope_id`, `role_id` and `role_ids` for *every* resource that has them.
The proposal is to turn that switch into data and to make it exhaustive:

```go
// config/references/registry.go
var Registry = map[string]Rule{
    "realm_id":                      {Reference: ref("keycloak_realm")},
    "organization_id":               {Reference: ref("keycloak_organization")},
    "first_broker_login_flow_alias": {Reference: ref("keycloak_authentication_flow", common.PathAuthenticationFlowAliasExtractor)},
    "post_broker_login_flow_alias":  {Reference: ref("keycloak_authentication_flow", common.PathAuthenticationFlowAliasExtractor)},

    // Deliberately not a reference — documented, not just absent.
    "provider_id": {NotAReference: "Keycloak provider implementation id (e.g. `oidc`), not an object id"},
    "tenant_id":   {NotAReference: "Microsoft tenant, external to Keycloak"},
    "entity_id":   {NotAReference: "SAML entity id, a URL chosen by the user"},
}
```

Rules:

- A rule may be scoped: `Only: []string{"keycloak_role"}` /
  `Except: []string{"keycloak_openid_client"}`, for attributes such as
  `client_id` whose meaning differs per resource.
- A rule may point at `multitypes` instead of a single reference, so that
  multi-type attributes are declared in the same place as single-type ones.
- Per-resource configuration in `config/<group>/config.go` keeps working and
  keeps winning: the registry is only a default, applied through the existing
  `WithDefaultResourceOptions` hook.

The important half is the **completeness gate**: a unit test walks the
generated provider and fails when a non-computed, non-sensitive attribute
matching `*_id`, `*_ids` or `*_alias` is neither wired as a reference, nor
covered by a `NotAReference` rule, nor listed with a reason in an explicit
exception file. That is the same shape as the gates the repo already trusts
(`generated-lst-check`, `e2e-cases-check`, `docs-freshness-check`), so it costs
nothing new conceptually. Adding a Terraform resource that introduces an unknown
`*_id` then *forces* a decision, and that decision is visible in the diff.

Expected effect on the current tree: the `post_broker_login_flow_alias` gap
above is closed as a side effect, and the remaining `provider_id`/`tenant_id`
attributes become documented non-references instead of silent omissions.

### B. `make new-resource` — scaffolding, not magic

```
make new-resource TF_RESOURCE=keycloak_organization_membership
```

Reads `config/schema.json` and writes exactly the files a human would write,
each with `TODO(<resource>)` markers where a decision is required:

| Generated file | Content |
|----------------|---------|
| `config/external_name.go` | one entry, defaulted to `config.IdentifierFromProvider`, annotated with the composite ID format read from the schema, plus a `TODO` to switch to an identifying-properties lookup if the object can pre-exist |
| `config/<group>/config.go` | a configurator stub, pre-filled with every reference the registry (A) can resolve, and a `TODO` per unresolved reference-shaped attribute |
| `examples/<group>/<resource>.yaml` | skeleton with all required fields, values left as `TODO` |
| `dev/demos/<suite>/NNN-<resource>.yaml` + case-list entry | skeleton demo so `make e2e-cases-check` passes only after it is filled in |

The command never rewrites a file that already exists; it prints the diff it
would apply and exits non-zero instead. Scaffolding is a one-shot operation, so
the output is owned by humans from then on — no round-trip, no merge markers,
no "generated block" fences inside hand-written files.

This is also the single command an AI agent can be told to run (see
[0002](0002-ai-skills-single-source-of-truth.md)): the model then only has to
fill in `TODO`s, which is exactly the kind of local, verifiable work cheap
models do well.

### C. Extend the existing schema-diff automation to report *quality*, not just presence

`scripts/schema_diff_issues.py` already files an issue per Terraform resource
that is not exposed yet. The same script can, in a second pass, report
configured resources whose configuration is incomplete according to (A) — e.g.
"`keycloak_saml_identity_provider` has `post_broker_login_flow_alias` unwired
while its siblings wire it". This reuses machinery the repo already runs weekly,
and costs one function.

## Rejected Alternatives

- **Infer references purely from attribute names, with no registry.** The
  measurement above shows a false-positive rate of roughly a third
  (`provider_id`, `tenant_id`, `entity_id`, `key_alias`, client `client_id`).
  Silently wiring those would produce wrong CRDs and break users.
- **A YAML/HCL configuration DSL for resource config.** It would duplicate what
  Go already expresses type-safely, add a parser to maintain, and lose
  compile-time checking of extractors and reference targets.
- **Round-trippable generated blocks inside `config/<group>/config.go`.** Marker
  comments that a generator rewrites in place are a classic source of merge
  conflicts and accidental deletions. Scaffold-once plus a completeness gate
  gives the same guarantee without owning other people's files.
- **Deriving external names from the schema.** The Terraform schema does not
  describe Keycloak's ID semantics. Guessing here produces resources that import
  incorrectly; this must stay explicit until an OpenAPI spec exists.

## Rollout

1. Introduce `config/references` with today's `KnownReferencers()` content moved
   into it, plus the completeness gate as a **reporting-only** test.
2. Triage the ~31 currently unwired attributes: wire the genuine gaps, declare
   the rest as `NotAReference` with a reason. Flip the gate to failing.
3. Add `make new-resource`, and use it for the next resource added; refine the
   templates against that experience before advertising it.
4. Extend `scripts/schema_diff_issues.py` with the quality pass.

Each step is independently revertible and none of them changes generated CRDs
except step 2, which fixes real gaps (and therefore needs the usual review of
the `package/crds/` diff).

## Open Questions

- Should the completeness gate also cover *required* attributes that reference
  nothing (e.g. free-form `parent_id` on custom user federation), or only the
  name-shaped heuristic?
- Where should the "documented non-reference" list live — inline in the registry
  (as sketched) or in a separate file, like
  `cluster/test/uncovered-resources.txt` does for e2e coverage? Inline keeps the
  reason next to the decision; a separate file is easier to review in bulk.
- Should `make new-resource` also scaffold `docs/content/docs/using/resources/`
  pages, or would that just create empty pages nobody fills in?
