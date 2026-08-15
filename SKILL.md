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

1. Add entry to `config/external_name.go` (`config.IdentifierFromProvider`
   for a plain provider-assigned ID, or a
   `<group>.<Resource>IdentifierFromIdentifyingProperties` helper when the ID
   must be derived from other fields).
2. Create/update `config/<group>/config.go` with references and optional
   lookup config. If an attribute can reference more than one resource type,
   use `config/multitypes` (see "Multi-Type References" below).
3. Run `make generate`.
4. Add example to `examples/<group>/<resource>.yaml`.
5. Optionally add docs page to `docs/content/docs/using/resources/<resource>.md`.
6. Optionally add the resource to `cluster/test/cases.txt` plus an e2e
   manifest under `cluster/test/` or `dev/demos/`.

See "Adding a New Resource" in `AGENTS.md` for the full playbook (external
name selection, references, import-by-properties, testing).

## Automated Schema Diff Issues

`.github/workflows/schema-diff-issues.yml` runs
`scripts/schema_diff_issues.py` to diff `config/schema.json` against
`config/generated.lst` and file one GitHub issue per missing resource that
isn't already tracked. It dry-runs only on pull requests that change the
automation itself (script or workflow file), and creates real issues
on schedule, on `Makefile` changes on `main`, and on manual dispatch.

`config/generated.lst` is generated from `config.ExternalNameConfigs` by
`make generated-lst` (part of `make generate`); never edit it by hand.
`make generated-lst-check` gates its freshness in CI.

## Cross-Resource References

```go
// In config/<group>/config.go
r.References["realm_id"] = config.Reference{
    TerraformName: "keycloak_realm",
}
```

## Multi-Type References

`r.References` only supports one `TerraformName`. When a Terraform attribute
may hold the ID of several different resource types, use `config/multitypes`
to generate one strongly-typed field per type; the values are consolidated
back into the original Terraform field at runtime. Never expose such an
attribute as a raw ID field.

```go
// Scalar field: role client_id may be an OpenID or a SAML client
multitypes.ApplyToWithOptions(r, "client_id",
    &multitypes.Options{KeepOriginalField: true},
    multitypes.Instance{Name: "client_id", Reference: config.Reference{
        TerraformName: "keycloak_openid_client", Extractor: common.PathUUIDExtractor}},
    multitypes.Instance{Name: "saml_client_id", Reference: config.Reference{
        TerraformName: "keycloak_saml_client", Extractor: common.PathUUIDExtractor}},
)

// List/set field: aggregate policy `policies` holds IDs of any policy type
multitypes.ApplyToAsList(r, "policies",
    multitypes.Instance{Name: "time_policies", Reference: config.Reference{
        TerraformName: "keycloak_openid_client_time_policy", Extractor: common.PathUUIDExtractor}},
    multitypes.Instance{Name: "role_policies", Reference: config.Reference{
        TerraformName: "keycloak_openid_client_role_policy", Extractor: common.PathUUIDExtractor}},
)
```

- `Options.KeepOriginalField: true` is required exactly when one `Instance`
  reuses the original field name (backward compatibility); otherwise the
  original field becomes computed-only and its required-parameter CEL rule is
  no longer emitted.
- An `Instance` reusing the original field name may omit its `Reference` to
  keep that field settable for raw, untyped IDs (no Ref/Selector generated).
- Every referenced `TerraformName` must exist in `config/external_name.go`, or
  `make generate` panics.
- Examples: `config/role/config.go`, `config/mapper/config.go`,
  `config/authentication/config.go`, `config/openidclient/config.go`.

## Import / Identify by Properties

To avoid 409 errors on create when a resource already exists in Keycloak,
wire the resource to `lookup.BuildIdentifyingPropertiesLookup` (see
`config/openidclient/config.go` for a full example).

## Documentation

- [Agents page](https://crossplane-contrib.github.io/provider-keycloak/docs/ai-usage/agents/)
- [llms.txt](https://crossplane-contrib.github.io/provider-keycloak/llms.txt)
- [llms-full.txt](https://crossplane-contrib.github.io/provider-keycloak/llms-full.txt)
