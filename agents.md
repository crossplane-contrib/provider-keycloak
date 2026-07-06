# Agent Instructions for provider-keycloak

> This file is intended for AI coding agents (GitHub Copilot, Cursor, Claude,
> etc.) working on this repository. For rendered documentation see
> https://crossplane-contrib.github.io/provider-keycloak/docs/ai-usage/agents/

## What This Repository Is

`provider-keycloak` is a [Crossplane](https://crossplane.io/) provider that
manages [Keycloak](https://www.keycloak.org/) resources as Kubernetes custom
resources. It is generated with [Upjet](https://github.com/crossplane/upjet)
on top of the [Keycloak Terraform Provider](https://github.com/keycloak/terraform-provider-keycloak).

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

## Code Generation

Always run `make generate` after changing `config/` to regenerate CRDs and
Go types. **Never** edit files in `apis/` or `package/crds/` by hand — they
are generated outputs.

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

## Important Constraints

- Do **not** edit `examples-generated/` by hand.
- Do **not** edit generated files in `apis/` or `package/crds/` by hand.
- Do **not** update `github.com/keycloak/terraform-provider-keycloak` via
  Renovate — it is explicitly excluded from automated dependency updates
  because upgrading it requires deliberate schema migration.
- E2E tests only cover resources listed in `cluster/test/cases.txt`.

## LLM Files

- https://crossplane-contrib.github.io/provider-keycloak/llms.txt
- https://crossplane-contrib.github.io/provider-keycloak/llms-full.txt
