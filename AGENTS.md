# Agent Instructions for provider-keycloak

> This file is intended for AI coding agents (GitHub Copilot, Cursor, Claude,
> etc.) working on this repository. For rendered documentation see
> https://crossplane-contrib.github.io/provider-keycloak/docs/ai-usage/agents/

## What This Repository Is

`provider-keycloak` is a [Crossplane](https://crossplane.io/) provider that
manages [Keycloak](https://www.keycloak.org/) resources as Kubernetes custom
resources. It is generated with [Upjet](https://github.com/crossplane/upjet)
on top of the [Keycloak Terraform Provider](https://github.com/keycloak/terraform-provider-keycloak).

**One-line flow:**
```
Keycloak Terraform Provider  →  Upjet (code generator)  →  Crossplane provider  →  Kubernetes CRDs
```

Users declare Keycloak resources as YAML (`spec.forProvider` maps to Terraform arguments),
and the provider reconciles them continuously against a live Keycloak instance.

## Repository Layout

```
apis/               Crossplane API types (generated + hand-authored)
cmd/                provider and generator entry points
config/             Upjet resource configuration (external names, references)
docs/               Hugo (hextra) documentation site
examples/           Hand-authored example manifests for each managed resource
examples-generated/ Auto-generated example manifests (do not edit by hand)
package/crds/       Generated CRD YAML (source of truth for field schemas)
internal/           Internal controller and reconciler logic
generate/           Generation scripts
cluster/            Uptest end-to-end test manifests and setup
dev/                Local development environment scripts
scripts/            Utility scripts
```

## Core Concepts

- **ProviderConfig** – connection details for a Keycloak instance (URL,
  client ID, credentials secret reference).
- **Managed Resources** – Kubernetes CRDs that map 1:1 to Keycloak objects.
  `spec.forProvider` maps to Terraform resource arguments.
- **Reconciliation** – the provider controller continuously ensures Keycloak
  matches the desired state expressed in `spec.forProvider`.
- **External Name** – the Keycloak-side identifier wired in
  `config/external_name.go`.
- **References** – cross-resource references (e.g., `realmIdRef`) are
  configured in `config/<group>/config.go`.

## Key Files for Common Tasks

| Task | File(s) |
|------|---------|
| Add a new resource | `config/external_name.go`, `config/<group>/config.go` |
| Change reference resolution | `config/<group>/config.go` |
| Update docs for a resource | `docs/content/docs/using/resources/<resource>.md` |
| Add/update an example manifest | `examples/<group>/<resource>.yaml` |
| Modify CRD generation | `generate/*.go`, run `make generate` |
| Run e2e tests | `make e2e`, see `cluster/test/cases.txt` for covered resources |
| Regenerate llms.txt/llms-full.txt | `make docs-gen` |

## Code Generation

Always run `make generate` after changing `config/` to regenerate CRDs and
Go types. **Never** edit files in `apis/` or `package/crds/` by hand — they
are generated outputs.

The generation pipeline:
1. `generate/main.go` calls Upjet with the Terraform provider schema.
2. Upjet writes Go types into `apis/<group>/<version>/`.
3. `make generate` runs `go generate ./...` which invokes controller-gen to write CRDs into `package/crds/`.

## Adding a New Resource

Use `make schema-diff OLD_PROVIDER_VERSION=<prev>` (or check
`config/generated.lst` against `config/schema.json`) to find Terraform
resources that aren't yet exposed as managed resources. The
`schema-diff-issues` GitHub Actions workflow automates this and files one
issue per missing resource (see "Automated Schema Diff Issues" below).

1. **External name.** Add an entry for the resource to
   `config/external_name.go`. Pick the mapping based on how Keycloak/the
   Terraform provider identifies the resource:
   - `config.IdentifierFromProvider` — the ID is a plain string returned by
     the provider (e.g. a UUID) with no need to derive it from other fields.
     Use this for resources with a simple, provider-assigned ID, including
     composite `{realm}/...` style IDs already produced by the TF provider.
   - `<group>.<Resource>IdentifierFromIdentifyingProperties` — the resource
     has no single stable ID until created, so the external name must be
     derived from a set of identifying attributes (e.g. name + realm). Add a
     small helper in `config/<group>/` following the pattern of
     `config/openidclient/*IdentifierFromIdentifyingProperties` or
     `config/group/group.go`.
   - Inspect the resource's docs page in the
     [terraform-provider-keycloak repo](https://github.com/keycloak/terraform-provider-keycloak)
     and, if unsure, check how a similar existing resource in the same group
     is mapped.
2. **References.** Identify which schema attributes point at other Keycloak
   resources (commonly `realm_id`, `client_id`, `parent_id`, etc.) and wire
   them in `config/<group>/config.go`:
   ```go
   r.References["realm_id"] = config.Reference{
       TerraformName: "keycloak_realm",
   }
   ```
   For references that require extracting an ID from a composite external
   name (e.g. `{realm}/{id}`), use `resolve.ExtractResourceID` or a custom
   extractor — see existing entries in `config/<group>/config.go` for
   patterns.

   If a single Terraform attribute can point at **more than one** resource
   type (e.g. an ID that may belong to an OpenID or a SAML client), do not
   leave it as a raw ID field — use `config/multitypes` (see
   "Multi-Type References" below).
3. **Import / identify by properties.** If the resource can already exist in
   Keycloak before Crossplane creates it (so a naive `Create` would 409),
   wire it to `lookup.BuildIdentifyingPropertiesLookup` in the group config
   (see `config/openidclient/config.go` for a full example) so it can be
   found and imported instead.
4. **Generate.** Run `make generate`. Confirm new/updated files appear under
   `apis/<group>/<version>/` and `package/crds/`, and that the CRD's
   `spec.forProvider` fields match the Terraform resource's schema.
5. **Examples.** Add a hand-authored example manifest to
   `examples/<group>/<resource>.yaml` with realistic field values (do not
   edit `examples-generated/` by hand).
6. **Docs.** Optionally add/update
   `docs/content/docs/using/resources/<resource>.md` and run `make docs-gen`
   to refresh `llms.txt`/`llms-full.txt`.
7. **Tests.** If the resource is meant to be covered by end-to-end tests,
   add it to `cluster/test/cases.txt` and provide (or extend) a
   chainsaw/uptest manifest under `cluster/test/` or `dev/demos/`. Run
   `make test` for unit tests; e2e coverage is otherwise limited to
   resources already listed in `cluster/test/cases.txt`.

## Automated Schema Diff Issues

`.github/workflows/schema-diff-issues.yml` runs
`scripts/schema_diff_issues.py`, which diffs `config/schema.json` against
`config/generated.lst` and files one GitHub issue per resource that is
missing and not already tracked by an existing open issue.

- On `pull_request`, it only runs (in `--dry-run` mode; never creates issues)
  when the automation itself changes (`scripts/schema_diff_issues.py` or the
  workflow file) — not on every PR.
- On `schedule` (weekly) and `push` to `main` that touches the `Makefile`
  (i.e. a Terraform provider version bump), and on manual
  `workflow_dispatch`, it creates real issues.
- Dedup is based on whether the exact resource name (as a whole token, not a
  substring) already appears in an open issue's title or body.

## Cross-Resource References

References are wired in `config/<group>/config.go` via `r.References`:

```go
r.References["realm_id"] = config.Reference{
    TerraformName: "keycloak_realm",
}
```

## Multi-Type References

`r.References` only accepts a single `TerraformName`, so it cannot express a
Terraform attribute whose value may be the ID of several different resource
types. For those, use the `config/multitypes` package instead of hand-rolled
synthetic fields or raw ID inputs. It creates one synthetic, strongly-typed
field per referenceable type, and consolidates the resolved values back into
the original Terraform field before the request is sent to Terraform.

Scalar field (`multitypes.ApplyTo` / `ApplyToWithOptions`) — e.g. a role's
`client_id`, which may reference an OpenID or a SAML client:

```go
multitypes.ApplyToWithOptions(r, "client_id",
    &multitypes.Options{KeepOriginalField: true}, // keep client_id settable
    multitypes.Instance{
        Name: "client_id",
        Reference: config.Reference{
            TerraformName: "keycloak_openid_client",
            Extractor:     common.PathUUIDExtractor,
        },
    },
    multitypes.Instance{
        Name: "saml_client_id",
        Reference: config.Reference{
            TerraformName: "keycloak_saml_client",
            Extractor:     common.PathUUIDExtractor,
        },
    },
)
```

List/set field (`multitypes.ApplyToAsList` / `ApplyToAsListWithOptions`) — e.g.
`keycloak_openid_client_aggregate_policy.policies`, which holds IDs of any
policy type:

```go
multitypes.ApplyToAsList(r, "policies",
    multitypes.Instance{
        Name: "time_policies",
        Reference: config.Reference{
            TerraformName: "keycloak_openid_client_time_policy",
            Extractor:     common.PathUUIDExtractor,
        },
    },
    multitypes.Instance{
        Name: "role_policies",
        Reference: config.Reference{
            TerraformName: "keycloak_openid_client_role_policy",
            Extractor:     common.PathUUIDExtractor,
        },
    },
    // ... one Instance per referenceable policy type
)
```

Rules and gotchas:

- `Options.KeepOriginalField: true` is required (and only allowed) when one
  `Instance` reuses the original field name; use it to keep the existing field
  settable for backward compatibility. The helper panics on a mismatch.
- If no `Instance` reuses the original name, the original field becomes
  computed-only (`status.atProvider`). A useful side effect is that a
  *required* Terraform field no longer emits a required-parameter CEL rule,
  so no CRD post-processing is needed.
- For scalar fields, only one synthetic field may be set at a time; the
  consolidation injector returns an error otherwise. For list fields, all
  synthetic lists are unioned.
- Every `TerraformName` you reference must be present in
  `config/generated.lst`; otherwise `make generate` panics with
  `cannot find configuration for Terraform resource`.
- Examples in the codebase: `config/role/config.go` and `config/mapper/config.go`
  (`client_id`/`saml_client_id`), `config/authentication/config.go`
  (`parent_flow_alias`/`parent_subflow_alias`), `config/openidclient/config.go`
  (`clients`/`saml_clients` and aggregate-policy `policies`),
  `config/identityprovider/config.go` (`provider_alias`/`identity_provider_alias`
  wired to every identity provider type).

## Testing

- Unit tests: `make test`
- E2E tests: `make e2e` (requires a running Keycloak and Crossplane cluster)
- E2E coverage is limited to resources listed in `cluster/test/cases.txt`

## Documentation

Docs use [Hugo](https://gohugo.io/) with the
[Hextra](https://imfing.github.io/hextra/) theme.

```bash
cd docs && hugo server --buildDrafts   # local preview
make docs-gen                          # regenerate llms.txt
make docs-freshness-check             # CI: verify llms.txt is current
```

## Known Constraints and Pitfalls

- Do **not** edit `examples-generated/` by hand.
- Do **not** edit generated files in `apis/` or `package/crds/` by hand.
- `github.com/keycloak/terraform-provider-keycloak` (the go.mod pseudo-version
  and its pinned version in the Makefile) is grouped into a single weekly
  Renovate PR. Minor/patch/digest updates auto-merge once tests pass; major
  version bumps are **not** auto-merged since they require deliberate schema
  migration and must be reviewed manually.
- E2E tests only cover resources listed in `cluster/test/cases.txt`.
- **Upjet does not support `+nullable` markers.** The kubebuilder Options struct
  only supports Required, Minimum, Maximum, Default.
- **Membership ownership:** Avoid managing the same group's membership with both
  a `Memberships` resource and a `Groups` resource with `exhaustive=true` at the
  same time; this can cause reconciliation loops.
- **E2E CI versioning:** Jobs that build or deploy local xpkgs must fetch git tags
  (`git fetch --tags`) so `build/makelib/common.mk` derives the correct VERSION.

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| CRD fields not updating | `make generate` not run | Run `make generate` |
| `409 Conflict` on create | Resource already exists in Keycloak | Use `lookup.BuildIdentifyingPropertiesLookup` |
| `llms-full.txt is stale` in CI | Docs changed, `make docs-gen` not run | Run `make docs-gen` and commit |
| `no matches for kind` in e2e | CRD not established in time | See `cluster/test/setup.sh` MRD wait logic |
| `make generate` produces unexpectedly large/stale diffs | Stale local generator cache/artifacts | Remove `.work/` and `config/schema.json`, then re-run `make generate` |
| E2E provider version mismatch | Git tags not fetched before build | Add `git fetch --tags` before `make build` |
| Reconciliation loop on membership | Both `Memberships` + `Groups` (exhaustive) target same group | Use only one authoritative source |

## LLM Files

- https://crossplane-contrib.github.io/provider-keycloak/llms.txt
- https://crossplane-contrib.github.io/provider-keycloak/llms-full.txt
