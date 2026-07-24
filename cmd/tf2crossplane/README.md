# tf2crossplane

`tf2crossplane` converts Terraform HCL that uses the
[Keycloak Terraform provider](https://github.com/keycloak/terraform-provider-keycloak)
into Crossplane Managed Resource manifests for provider-keycloak.

provider-keycloak is generated with [Upjet](https://github.com/crossplane/upjet)
on top of the same Terraform provider, so every CRD's `spec.forProvider` maps
1:1 onto the Terraform resource arguments. This tool loads the exact provider
configuration used by the code generator (`config.GetProvider`) as the single
source of truth for:

- the Terraform resource name → CRD GroupVersionKind mapping,
- the snake_case → camelCase field-name transformation, and
- cross-resource reference wiring (e.g. `realm_id` → `realmIdRef`).

## Build

```bash
go build -o tf2crossplane ./cmd/tf2crossplane
```

## Usage

```bash
tf2crossplane main.tf                 # convert a file, print to stdout
tf2crossplane ./terraform/            # convert every *.tf in a directory
cat main.tf | tf2crossplane -o out.yaml
tf2crossplane --namespaced main.tf    # emit keycloak.m.crossplane.io resources
tf2crossplane --list-supported        # list convertible Terraform types
```

| Flag | Description |
|------|-------------|
| `-o, --output` | Write manifests to a file instead of stdout. |
| `--namespaced` | Emit namespaced (`keycloak.m.crossplane.io`) resources. |
| `--provider-config` | `spec.providerConfigRef.name` value (default `keycloak-provider-config`). |
| `--deletion-policy` | `spec.deletionPolicy` value (e.g. `Delete`, `Orphan`). |
| `--management-policies` | Comma-separated `spec.managementPolicies` list. |
| `--list-supported` | Print convertible Terraform resource types and exit. |
| `-q, --quiet` | Suppress warnings on stderr. |

## Limitations

Static conversion cannot evaluate everything a Terraform plan would. Variables,
locals, expressions, `count`/`for_each`, data sources, and modules are surfaced
as placeholders or skipped with a warning. Resources with no matching CRD are
reported and skipped. Always review the generated manifests and resolve any
warnings before applying them.

See the [Terraform Migration reference](../../docs/content/docs/using/reference/terraform-migration.md)
for details.
