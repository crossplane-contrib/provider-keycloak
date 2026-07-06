---
name: provider-keycloak
description: Build, test, and extend provider-keycloak — a Crossplane provider that manages Keycloak (IAM/SSO) resources as Kubernetes custom resources.
---

# provider-keycloak Skill

provider-keycloak is a [Crossplane](https://crossplane.io/) provider generated with
[Upjet](https://github.com/crossplane/upjet) from the
[Keycloak Terraform Provider](https://github.com/keycloak/terraform-provider-keycloak).
It reconciles Kubernetes CRDs against a live Keycloak instance.

## Quick Reference

```
# Regenerate CRDs + Go types after changing config/
make generate

# Run unit tests
make test

# Preview docs locally
cd docs && hugo server --buildDrafts

# Regenerate llms.txt / llms-full.txt
make docs-gen

# Verify docs freshness (CI gate)
make docs-freshness-check
```

## Build and Test

```bash
make generate   # regenerate after config/ changes
make test       # unit tests
make e2e        # end-to-end tests (requires live cluster + Keycloak)
```

## Key Files

| File | Purpose |
|------|---------|
| `config/external_name.go` | Maps Terraform resource names to Keycloak-side identifiers |
| `config/<group>/config.go` | Cross-resource references, import config, custom behaviors |
| `generate/main.go` | Entry point for Upjet code generation |
| `cluster/test/cases.txt` | Resources covered by E2E tests |
| `docs/scripts/gen-llms.sh` | Generates llms.txt and llms-full.txt |

## Conventions

- Never edit files in `apis/` or `package/crds/` by hand — they are generated outputs.
- Never edit `examples-generated/` by hand.
- Always run `make generate` after changing `config/`.
- Do not add `+nullable` markers to generated types (Upjet does not support them).
- Do not let both a `Memberships` resource and a `Groups` resource with `exhaustive=true`
  manage the same group's membership simultaneously — they will conflict.
- `github.com/keycloak/terraform-provider-keycloak` must not be updated via Renovate;
  it requires deliberate schema migration.

## Adding a New Resource

1. Add entry to `config/external_name.go`.
2. Create/update `config/<group>/config.go` with references and optional lookup config.
3. Run `make generate`.
4. Add example to `examples/<group>/<resource>.yaml`.
5. Optionally add docs page to `docs/content/docs/using/resources/<resource>.md`.

## Cross-Resource References

```go
// In config/<group>/config.go
r.References["realm_id"] = config.Reference{
    TerraformName: "keycloak_realm",
}
```

## Import / Identify by Properties

To avoid 409 errors on create when a resource already exists in Keycloak,
wire the resource to `lookup.BuildIdentifyingPropertiesLookup` (see
`config/openidclient/config.go` for a full example).

## Documentation

- [Agents page](https://crossplane-contrib.github.io/provider-keycloak/docs/ai-usage/agents/)
- [llms.txt](https://crossplane-contrib.github.io/provider-keycloak/llms.txt)
- [llms-full.txt](https://crossplane-contrib.github.io/provider-keycloak/llms-full.txt)
