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

1. Add an entry to `config/external_name.go`.
2. Create or update `config/<group>/config.go` to configure references.
3. Run `make generate`.
4. Add a hand-authored example to `examples/<group>/<resource>.yaml`.

To allow import/observe by identifying properties (avoiding 409 on create), wire
to `lookup.BuildIdentifyingPropertiesLookup` in the group config (see
`config/openidclient/config.go` for an example).

## Cross-Resource References

References are wired in `config/<group>/config.go` via `r.References`:

```go
r.References["realm_id"] = config.Reference{
    TerraformName: "keycloak_realm",
}
```

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
- Do **not** update `github.com/keycloak/terraform-provider-keycloak` via
  Renovate — it is explicitly excluded from automated dependency updates
  because upgrading it requires deliberate schema migration.
- E2E tests only cover resources listed in `cluster/test/cases.txt`.
- **Upjet does not support `+nullable` markers.** The kubebuilder Options struct
  only supports Required, Minimum, Maximum, Default.
- **Membership conflicts:** Never let both a `Memberships` resource (authoritative)
  and a `Groups` resource with `exhaustive=true` manage the same group — they will
  fight and cause reconciliation loops.
- **E2E CI versioning:** Jobs that build or deploy local xpkgs must fetch git tags
  (`git fetch --tags`) so `build/makelib/common.mk` derives the correct VERSION.

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| CRD fields not updating | `make generate` not run | Run `make generate` |
| `409 Conflict` on create | Resource already exists in Keycloak | Use `lookup.BuildIdentifyingPropertiesLookup` |
| `llms-full.txt is stale` in CI | Docs changed, `make docs-gen` not run | Run `make docs-gen` and commit |
| `no matches for kind` in e2e | CRD not established in time | See `cluster/test/setup.sh` MRD wait logic |
| E2E provider version mismatch | Git tags not fetched before build | Add `git fetch --tags` before `make build` |
| Reconciliation loop on membership | Both `Memberships` + `Groups` (exhaustive) target same group | Use only one authoritative source |

## LLM Files

- https://crossplane-contrib.github.io/provider-keycloak/llms.txt
- https://crossplane-contrib.github.io/provider-keycloak/llms-full.txt
