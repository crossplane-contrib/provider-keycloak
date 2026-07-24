---
title: Terraform Migration
weight: 5
---

`tf2crossplane` is a command-line tool that converts Terraform HCL using the
[Keycloak Terraform provider](https://github.com/keycloak/terraform-provider-keycloak)
into Crossplane Managed Resource manifests for provider-keycloak.

Because provider-keycloak is generated with [Upjet](https://github.com/crossplane/upjet)
on top of the same Terraform provider, every CRD's `spec.forProvider` maps 1:1
onto the Terraform resource arguments. The converter loads the exact provider
configuration used by the code generator, so its Terraform&nbsp;&rarr;&nbsp;CRD
mapping always stays in lockstep with the generated CRDs.

## Building

The tool lives in `cmd/tf2crossplane`:

```bash
go build -o tf2crossplane ./cmd/tf2crossplane
```

## Usage

```bash
# Convert a single file
tf2crossplane main.tf

# Convert every *.tf file in a directory
tf2crossplane ./terraform/

# Read from stdin, write to a file
cat main.tf | tf2crossplane -o resources.yaml

# Emit namespaced (keycloak.m.crossplane.io) resources
tf2crossplane --namespaced main.tf

# List the Terraform resource types the converter can map
tf2crossplane --list-supported
```

### Flags

| Flag | Description |
|------|-------------|
| `-o, --output` | Write manifests to a file instead of stdout. |
| `--namespaced` | Emit namespaced (`keycloak.m.crossplane.io`) resources. |
| `--provider-config` | Name written into `spec.providerConfigRef.name` (default `keycloak-provider-config`). |
| `--deletion-policy` | Value written into `spec.deletionPolicy` (e.g. `Delete`, `Orphan`). |
| `--management-policies` | Comma-separated list written into `spec.managementPolicies`. |
| `--list-supported` | Print the convertible Terraform resource types and exit. |
| `-q, --quiet` | Suppress warnings on stderr. |

## What it converts

- **Resources** &mdash; each `resource "keycloak_*" "name"` block becomes a
  managed resource. Argument names are converted from snake_case to camelCase
  (`realm_id` &rarr; `realmId`), and nested blocks become nested lists to match
  the generated CRDs.
- **References** &mdash; when an argument points at another Keycloak resource in
  the same configuration (for example `realm_id = keycloak_realm.this.id`) and
  that field is a configured cross-resource reference, the tool emits the
  idiomatic Crossplane form (`realmIdRef: { name: this }`) instead of a literal
  ID.
- **Provider block** &mdash; a `provider "keycloak"` block is translated into a
  [`ProviderConfig`](/docs/using/reference/provider-config/) plus a credentials
  `Secret` scaffold. Secret values are always placeholders&mdash;never commit
  real credentials.

## Limitations

A static converter cannot evaluate everything a Terraform plan would:

- **Variables, locals, expressions, `count`/`for_each`** cannot be resolved
  statically. The tool emits the original expression text as a placeholder and
  prints a warning so you can fill in the value by hand.
- **Data sources and modules** are skipped with a warning. Data sources
  correspond to observing/importing existing objects rather than managing them.
- **Resources with no matching CRD** are reported (see `--list-supported`) and
  skipped rather than producing invalid YAML.

Always review the generated manifests and address any warnings before applying
them.
