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
- But the same attribute is configured differently on resources whose schema is
  identical. `post_broker_login_flow_alias` is declared as an optional string on
  exactly the same nine identity-provider resources as
  `first_broker_login_flow_alias`, yet it is wired on two of them
  (`keycloak_spiffe_identity_provider`,
  `keycloak_oidc_openshift_v4_identity_provider`) and unwired on the other
  seven. `first_broker_login_flow_alias` is wired on eight of the nine and
  unwired on `keycloak_saml_identity_provider`.

So the real problem is not "too much typing". It is that **the configuration has
no consistency gate**, which makes drift invisible and makes onboarding a
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

## What the Schema Can and Cannot Tell Us

This was the first question raised in review, and it is worth answering
precisely, because the answer decides the whole design.

### It *can* find the drift — via a schema × config cross-check

The schema alone cannot tell us whether an attribute is a reference. But the
schema plus the current configuration can mechanically tell us that an attribute
is configured **inconsistently**: same attribute name, same type, same
optionality, different treatment. Run over the whole tree, that check produces
exactly five attributes:

| Attribute | Wired to | Unwired |
|-----------|----------|---------|
| `post_broker_login_flow_alias` | `keycloak_authentication_flow` (2) | 7 |
| `first_broker_login_flow_alias` | `keycloak_authentication_flow` (8) | 1 (`keycloak_saml_identity_provider`) |
| `parent_id` | `keycloak_group` (1) | 1 (`keycloak_custom_user_federation`) |
| `client_id` | `keycloak_openid_client` (22), `keycloak_saml_client` (3) | 3 |
| `client_scope_id` | `keycloak_openid_client_scope` (16), `keycloak_saml_client_scope` (2) | 0 |

Five findings for a reviewer to classify once, with **no false positives on the
question asked** — the check does not claim these are bugs, only that the same
attribute is treated in more than one way and that the reason should be
recorded. That is a far sharper signal than the name heuristic in the previous
section, which needs a human to dismiss roughly a third of its hits. It is also
strictly cheaper to run: no new metadata, just the provider config the generator
already builds.

The last two rows are the interesting ones, and they lead straight to the second
point.

### It *suggests* where multitypes belong — from the shape of the disagreement

`client_id` and `client_scope_id` are not gaps at all: they are wired to an
OpenID target on some resources and a SAML target on others, which is exactly
the situation `config/multitypes` exists for, and exactly how the repo already
handles them (`config/mapper/config.go` applies
`multitypes.ApplyToWithOptions` to `client_scope_id` on the four SAML-capable
mappers; `config/role/config.go` does the same for `client_id`).

So **"one attribute name resolving to more than one target type" is a mechanical
multitype signal**, and the review comment is right that the flow aliases smell
the same way: `keycloak_authentication_flow` and `keycloak_authentication_subflow`
both expose a required `alias`, so a `*_flow_alias` has two plausible targets,
and `parent_flow_alias` in `config/authentication/config.go` is already modelled
as a multitype over exactly that pair.

Whether the broker-login aliases should follow is a genuine question rather than
a foregone conclusion, and it is worth stating what was actually verified
upstream:

- Keycloak resolves both fields with `realm.getFlowByAlias()`
  (`RepresentationToModel`), which has **no `topLevel` filter** — the server
  accepts a subflow alias.
- The admin console only ever offers top-level flows: `GET /authentication/flows`
  filters on `flow.isTopLevel()`, so a subflow can never be selected in the UI.
- The Terraform provider validates nothing; the string is passed through.
- Whether Keycloak *executes* a subflow correctly in the broker-login context
  was not verified.

That is precisely a decision that must be recorded rather than guessed: wiring
`post_broker_login_flow_alias` to `keycloak_authentication_flow` on all nine
resources fixes the inconsistency and matches the admin console, while extending
it to a flow/subflow multitype matches `parent_flow_alias` and the server's
actual behaviour. The proposal below is the mechanism that forces the choice to
be made and written down; it deliberately does not make the choice.

### It *cannot* derive the target set

Deriving "which resource does this attribute point at?" from names does not
survive contact with the data. Matching an attribute's tokens against resources
that expose the corresponding bare identifier (`alias` for `*_alias`, `id` for
`*_id`) yields 17 candidates for `realm_id` (including `keycloak_realm_events`
and every `keycloak_realm_keystore_*`) and 14 for `role_id`. Worse, for
`*_flow_alias` it returns **only** `keycloak_authentication_flow` — it misses
`keycloak_authentication_subflow`, because "subflow" is not the token "flow".
The one case where a derived target set would have been most useful is the case
it gets wrong.

Conclusion: the schema is a good **detector** and a poor **decider**. The design
below leans on it for detection and keeps every decision explicit in Go.

### AI can close the gap between detector and decider — by reading upstream

This is the second question from review, and it fits exactly into the hole left
above. The cross-check produces findings; what it cannot produce is the
*reason*. The reason is almost always written down somewhere upstream — in the
Terraform provider's docs and Go code, and behind that in the Keycloak server.
Reading those three sources per finding is tedious, repetitive, requires no
repository-specific judgement, and ends in a citation. That is a good fit for an
agent, and it is precisely the `research-upstream` skill in
[0002](0002-ai-skills-single-source-of-truth.md).

The important property is that the corpus is **pinned and local**, so this is
grounded lookup rather than open-ended web browsing:

- `make pull-docs` already sparse-checks out `docs/resources` of the Terraform
  provider at `TERRAFORM_PROVIDER_VERSION` (`5.9.0`) into
  `.work/keycloak/keycloak/`, and `generate.init` depends on it — the corpus is
  on disk before generation runs, at the same version as `config/schema.json`.
- The provider's Go source is pinned in `go.mod` and addressable with
  `go list -m -f '{{.Dir}}' github.com/keycloak/terraform-provider-keycloak`,
  so `keycloak/identity_provider.go` and friends can be read at the exact
  revision this repository builds against.

The doc corpus carries the semantics the schema drops. Of the 242 non-computed,
non-sensitive top-level attributes matching `*_id`/`*_ids`/`*_alias`, **228 are
described in prose upstream** — including the distinction the name heuristic
gets wrong (`provider_id` is documented as the provider implementation name,
`post_broker_login_flow_alias` as "the authentication flow to use after users
have successfully logged in"). Where a Crossplane reference is warranted, the
upstream text says so in words.

The remaining **14 are undocumented upstream**, and they cluster tellingly:

| Undocumented attribute | Resources |
|------------------------|-----------|
| `organization_id` | 7 (all the social/OIDC IdPs, `kubernetes`, `spiffe`) |
| `first_broker_login_flow_alias`, `post_broker_login_flow_alias`, `provider_id` | `keycloak_kubernetes_identity_provider`, `keycloak_spiffe_identity_provider` |
| `client_scope_id` | `keycloak_generic_client_protocol_mapper` |

The newest resources are both the least documented upstream *and* the ones this
repository configured inconsistently — `keycloak_spiffe_identity_provider` is
one of the two resources that wired `post_broker_login_flow_alias`. Absent
upstream prose is therefore itself a useful signal: it marks the findings a
human must decide, and it is an actionable upstream contribution (a docs PR)
rather than a local workaround.

The division of labour that follows:

| Layer | Produces | Trust |
|-------|----------|-------|
| Cross-check (deterministic) | the finding: "same attribute, different treatment" | authoritative, gates CI |
| Agent research (upstream docs → provider Go → Keycloak server) | a proposal with `file:line` citations, and what remains unverified | reviewed input, never authoritative |
| Registry entry / `NotAReference` reason (human-reviewed Go) | the decision | authoritative, in the diff |

Guardrails, so this stays a research aid and not an oracle:

- An agent never edits `config/` unattended as a result of a scan. It produces a
  proposed diff plus citations; the gate still fails until a human merges it.
- Every claim cites a file and line in the pinned tree. A claim that cannot be
  cited is reported as unverified — like "does Keycloak execute a *subflow* in
  the broker-login context?", which stayed open above and belongs in an upstream
  question, not in a guess.
- The deterministic check remains the source of the *finding set*. The agent
  never decides what to look at, only what a finding means, which keeps the
  output bounded and reproducible (five findings today, not "whatever the model
  noticed").

Concretely, this is what turns step 2 of the rollout from "a maintainer reads
Keycloak source for an afternoon" into "review five annotated findings".

## Proposed Design

Four independent pieces. Each is useful alone, and each can be adopted without
the others. **A** is the tooling — the one thing that produces value on day one,
before any refactor.

### Principle: machine-generated first

The dividing line is *judgement*, not effort. Everything that can be derived
without a Keycloak-semantics decision should be generated or checked
mechanically; the small remainder should be a short, reviewable Go diff with a
recorded reason.

| Artifact | Today | Target |
|----------|-------|--------|
| CRDs, API types, examples-generated, `config/generated.lst`, e2e index, `llms.txt` | generated | unchanged |
| Detection of missing references / missing multitypes / drift | nothing | **generated report + CI gate** (A) |
| Per-resource reference expectation tests (`config/*_references_test.go`, ~735 hand-written lines) | hand-written per resource | subsumed by generic detectors; keep only genuinely special cases |
| Reference wiring for known attribute names | `KnownReferencers()` switch + per-group repetition | declarative registry (B), applied automatically |
| Scaffolding for a new resource | manual, by imitation | `make new-resource` (C) |
| Upstream semantics behind a finding | maintainer reads source | AI research pass with citations (see above) |
| External name, target choice, multitype-vs-gap | human | human, but *forced* and recorded |

### A. `make config-audit` — tooling that finds missing references and missing multitypes

A reporting tool over the already-built provider config
(`config.GetProvider(true)`) plus `config/schema.json`. No refactor, no new
metadata, no dependency on the registry below. It runs three detectors, and all
three have been measured against the current tree:

**D1 — inconsistent treatment.** Same attribute name, same type, same
optionality, treated differently across resources. Five findings today (table
above). Sub-classified automatically:

- *gap*: wired on N resources, unwired on M — e.g. `post_broker_login_flow_alias`
  (2 wired / 7 unwired);
- *multitype*: one attribute name resolving to two different target types —
  `client_id`, `client_scope_id`, which are already `multitypes` and should be
  reported as **satisfied**, not as drift.

**D2 — unclassified reference-shaped attributes.** Non-computed, non-sensitive
`*_id` / `*_ids` / `*_alias` attributes that are neither wired nor documented as
non-references: 31 today. Noisy on its own (~⅓ are correct omissions), which is
why it only becomes a gate after the registry gives it somewhere to record the
`NotAReference` reason.

**D3 — missing multitypes, without needing drift to exist first.** This is the
detector the review asked for, and it does not depend on the attribute already
being inconsistent. It works in two steps:

1. Derive *type families* from the configuration itself: any attribute already
   wired to more than one target defines a family. Today that yields exactly two
   families — `{keycloak_openid_client, keycloak_saml_client}` and
   `{keycloak_openid_client_scope, keycloak_saml_client_scope}`.
2. Report every reference that points at **one** member of a family while the
   resource has no reference to the others.

Raw, that is 51 candidates — too many, and mostly correct: an
`keycloak_openid_*_protocol_mapper` legitimately only ever attaches to an OpenID
client. Adding one purely structural filter — the *referencing* resource is
itself protocol-neutral, i.e. its name is not prefixed `keycloak_openid_` or
`keycloak_saml_` — takes it from 51 to **3**:

| Resource | Attribute | Wired to | Family member not wired |
|----------|-----------|----------|--------------------------|
| `keycloak_ldap_role_mapper` | `client_id` | `keycloak_openid_client` | `keycloak_saml_client` |
| `keycloak_identity_provider_token_exchange_scope_permission` | `clients` | `keycloak_openid_client` | `keycloak_saml_client` |
| `keycloak_generic_client_authorization_policy` | `resource_server_id` | `keycloak_openid_client` | `keycloak_saml_client` |

All three are plausible and none is obviously decidable from the schema —
whether a SAML client can be an authorization *resource server*, or the target
of a token-exchange permission, is a Keycloak-semantics question. That is the
handoff point: D3 produces the shortlist, the `research-upstream` pass annotates
each entry with upstream citations, a human writes the two-line verdict. The
same resource is also the proof that the detector is not vacuous:
`keycloak_generic_client_protocol_mapper` — equally protocol-neutral — *is*
already wired to all four family members and therefore is not reported.

Output and integration:

- Human-readable table on stdout; `--format=json` for the automation in (D).
- `--fail-on=drift` initially (D1 only), widening to `missing-multitype` and
  then `unclassified` as each class is triaged. Same shape as
  `generated-lst-check` / `e2e-cases-check` / `docs-freshness-check`.
- Findings can be waived only with a reason, in the registry (B) or an explicit
  exceptions file — the `cluster/test/uncovered-resources.txt` pattern, where
  "not yet" is not an accepted reason and stale entries fail too.

The existing hand-written expectation tests (`config/*_references_test.go`,
seven files, ~735 lines listing per-resource expected targets) are the current
substitute for exactly this. They only cover the resources somebody remembered
to enumerate; the detectors cover all 109 by construction.

### B. A declarative attribute registry (replaces "imitate the neighbour")

`config/provider.go` already has the right idea in `KnownReferencers()`: a
`switch` over attribute names that wires `realm_id`, `organization_id`,
`client_scope_id`, `role_id` and `role_ids` for *every* resource that has them.
The proposal is to turn that switch into data and to make it exhaustive:

```go
// config/references/registry.go
var Registry = map[string]Rule{
    "realm_id":        {Reference: ref("keycloak_realm")},
    "organization_id": {Reference: ref("keycloak_organization")},

    // Multi-typed: declared once here instead of repeated per resource.
    "client_scope_id": {MultiType: []multitypes.Instance{ /* openid + saml client scope */ }},

    // Deliberately not a reference — documented, not just absent.
    "provider_id": {NotAReference: "Keycloak provider implementation id (e.g. `oidc`), not an object id"},
    "tenant_id":   {NotAReference: "Microsoft tenant, external to Keycloak"},
    "entity_id":   {NotAReference: "SAML entity id, a URL chosen by the user"},

    // TODO(#712): single-type today; revisit as a flow/subflow multitype like
    // parent_flow_alias once subflow execution in the broker-login context is
    // confirmed upstream.
    "post_broker_login_flow_alias": {Reference: ref("keycloak_authentication_flow", common.PathAuthenticationFlowAliasExtractor)},
}
```

Rules:

- A rule may be scoped: `Only: []string{"keycloak_role"}` /
  `Except: []string{"keycloak_openid_client"}`, for attributes such as
  `client_id` whose meaning differs per resource.
- A rule carries either a single `Reference`, a `MultiType` instance list, or a
  `NotAReference` reason — so multi-type attributes are declared in the same
  place, and in the same vocabulary, as single-type ones. Nothing about
  `config/multitypes` changes; the registry only moves the *declaration* next to
  its siblings so that divergence between them is visible.
- Per-resource configuration in `config/<group>/config.go` keeps working and
  keeps winning: the registry is only a default, applied through the existing
  `WithDefaultResourceOptions` hook.

The registry is what gives the detectors in (A) somewhere to record a verdict:
`NotAReference` answers D2, a `MultiType` rule answers D1/D3, and a scoped rule
answers "this attribute means something different here". Without it every
finding can only be silenced, not explained — which is how the current state
came about.

Adding a Terraform resource that introduces an unknown `*_id`, or that diverges
from its siblings, then *forces* a decision, and that decision is visible in the
diff.

Expected effect on the current tree: the two flow-alias gaps and `parent_id` are
closed or documented, `client_id`/`client_scope_id` are recorded as multitypes
rather than looking like drift, and the remaining `provider_id`/`tenant_id`
attributes become documented non-references instead of silent omissions.

### C. `make new-resource` — scaffolding, not magic

```
make new-resource TF_RESOURCE=keycloak_organization_membership
```

Reads `config/schema.json` and writes exactly the files a human would write,
each with `TODO(<resource>)` markers where a decision is required:

| Generated file | Content |
|----------------|---------|
| `config/external_name.go` | one entry, defaulted to `config.IdentifierFromProvider`, annotated with the composite ID format read from the schema, plus a `TODO` to switch to an identifying-properties lookup if the object can pre-exist |
| `config/<group>/config.go` | a configurator stub, pre-filled with every reference the registry (B) can resolve, and a `TODO` per unresolved reference-shaped attribute |
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

### D. Extend the existing schema-diff automation to report *quality*, not just presence

`scripts/schema_diff_issues.py` already files an issue per Terraform resource
that is not exposed yet. The same script can, in a second pass, run `make config-audit --format=json` from (A) and report configured resources whose configuration
diverges from their siblings — e.g. "`keycloak_saml_identity_provider` has
`post_broker_login_flow_alias` unwired while eight sibling resources with an
identical schema wire it". This reuses machinery the repo already runs weekly,
and costs one function.

## Rejected Alternatives

- **Infer references purely from attribute names, with no registry.** The
  measurement above shows a false-positive rate of roughly a third
  (`provider_id`, `tenant_id`, `entity_id`, `key_alias`, client `client_id`).
  Silently wiring those would produce wrong CRDs and break users.
- **Derive the reference *target* from the schema.** Token-matching an attribute
  against resources exposing the matching bare identifier gives 17 candidates
  for `realm_id`, 14 for `role_id`, and — decisively — misses
  `keycloak_authentication_subflow` for `*_flow_alias`, which is the one case
  where the answer would have mattered most.
- **A YAML/HCL configuration DSL for resource config.** It would duplicate what
  Go already expresses type-safely, add a parser to maintain, and lose
  compile-time checking of extractors and reference targets.
- **Round-trippable generated blocks inside `config/<group>/config.go`.** Marker
  comments that a generator rewrites in place are a classic source of merge
  conflicts and accidental deletions. Scaffold-once plus the gates above gives
  the same guarantee without owning other people's files.
- **Deriving external names from the schema.** The Terraform schema does not
  describe Keycloak's ID semantics. Guessing here produces resources that import
  incorrectly; this must stay explicit until an OpenAPI spec exists.

## Rollout

1. Add `make config-audit` with **D1 and D3**, reporting only. No registry, no
   refactor: it walks the already-built provider config and prints 5 + 3
   findings. Cheapest possible first step and the one with immediate value —
   the tooling exists before anything is restructured.
2. Classify those eight, using the `research-upstream` skill so each verdict
   arrives with upstream citations: wire the two flow-alias gaps and `parent_id`
   (or record why not), record `client_id`/`client_scope_id` as intentional
   multitypes, and decide whether the three protocol-neutral resources need a
   client/SAML-client multitype. Flip `--fail-on=drift,missing-multitype`.
3. Introduce `config/references` with today's `KnownReferencers()` content moved
   into it, plus the classifications from step 2, and retire the hand-written
   per-resource expectation tests that the detectors now cover.
4. Triage the ~31 unclassified attributes (D2) and enable that gate too.
5. Add `make new-resource`, and use it for the next resource added; refine the
   templates against that experience before advertising it.
6. Extend `scripts/schema_diff_issues.py` to consume `config-audit --format=json`.
7. Optionally let the weekly automation attach agent research to each reported
   finding (upstream doc excerpt + citations), so the issue arrives
   pre-triaged. Only worth doing once the reporting itself is trusted.

Each step is independently revertible. Only step 2 changes generated CRDs, and
it needs the usual review of the `package/crds/` diff. Steps 1–2 alone already
deliver what this proposal is mostly about: tooling that finds missing
references and missing multitypes, and a short list of decisions to make.

## Open Questions

- **Should the broker-login flow aliases become a flow/subflow multitype?**
  Raised in review on this proposal. The server accepts a subflow alias and
  `parent_flow_alias` already models exactly that pair, but the admin console
  only ever offers top-level flows and subflow execution in the broker-login
  context is unverified. Worth an upstream question before changing the API
  shape; the plain single-type wiring fixes the inconsistency in the meantime
  and is forward-compatible (a later multitype can keep the original field via
  `Options.KeepOriginalField`).
- Should the consistency check treat "wired on N resources, unwired on 1" and
  "wired to two different targets" as the same finding, or as two separate
  classes? They need different fixes (gap vs multitype), so separate reporting
  may make triage faster.
- Is D3's protocol-neutrality filter (referencing resource is not named
  `keycloak_openid_*` / `keycloak_saml_*`) the right one? It is a naming
  heuristic on the *referencing* side and it does the heavy lifting: 51
  candidates down to 3. It could instead be expressed as an explicit
  "protocol-neutral resource" list, or D3 could report all 51 with the filtered
  three highlighted, so nothing is silently dropped.
- Should families be seedable by hand as well as derived? Today only two
  families exist because only two attributes are multityped. A family that
  nobody has ever wired twice (e.g. flow/subflow, were `parent_flow_alias` not
  already multityped) is invisible to D3 — the detector can only generalise
  from a precedent.
- Should the completeness gate also cover *required* attributes that reference
  nothing (e.g. free-form `parent_id` on custom user federation), or only the
  name-shaped heuristic?
- Where should the "documented non-reference" list live — inline in the registry
  (as sketched) or in a separate file, like
  `cluster/test/uncovered-resources.txt` does for e2e coverage? Inline keeps the
  reason next to the decision; a separate file is easier to review in bulk.
- Should `make new-resource` also scaffold `docs/content/docs/using/resources/`
  pages, or would that just create empty pages nobody fills in?
- Should the agent-attached upstream research be committed (e.g. as the
  `NotAReference` reason text plus a citation comment) or only live in the PR
  discussion? Committing it makes the reasoning survive; it also means the
  citation can go stale on the next provider bump, unless a check verifies the
  cited file still exists in the pinned tree.
- Should the 14 upstream-undocumented attributes be raised as a docs
  contribution to `keycloak/terraform-provider-keycloak`? It would benefit every
  consumer of the provider, and it is the "report upstream instead of working
  around it" rule applied to ourselves.
