# End-to-End Tests

# End-to-End Tests

This page explains the e2e test infrastructure, common causes of **stuck
(hung) tests**, and the methodology for writing new tests.

## Overview

E2e tests are driven by [uptest](https://github.com/crossplane/uptest) and
[chainsaw](https://kyverno.github.io/chainsaw/). The test runner:

1. Reads the ordered list of demo manifests from `cluster/test/cases.txt`.
2. Calls `cluster/test/setup.sh` to prepare each test case.
3. Applies all resources in the manifest, waits for them to become Ready/Synced, imports them, then deletes them.

### File layout

| Path | Purpose |
|------|---------|
| `cluster/test/cases.txt` | Ordered list of demo files to test (deletion order is reversed) |
| `cluster/test/setup.sh` | Per-case setup: CRD readiness wait, timeout rewriting, ordered deletion |
| `dev/demos/basic/` | Cluster-scoped demo manifests |
| `dev/demos/namespaced/` | Namespace-scoped demo manifests |
| `dev/demos/basic/000-init.yaml` | Prerequisite resources for cluster demos (realm, secrets, etc.) |
| `dev/demos/namespaced/000-init.yaml` | Prerequisite resources for namespaced demos |

## Adding a New Test

1. Create `dev/demos/basic/<NNN>-<name>.yaml` and/or
   `dev/demos/namespaced/<NNN>-<name>.yaml`.
2. Add both paths to `cluster/test/cases.txt` in the correct position
   (higher numbers run first in the list; deletion is in reverse order, so
   resources that depend on earlier resources must have a higher number).
3. Make sure all prerequisites exist in earlier numbered demo files or in
   `000-init.yaml`.

## Why Tests Get Stuck

The following are the most frequent causes of stuck (never-Ready) resources
in e2e tests.

### 1. Wrong API Kind or Version

The generated kind names do **not** always match the Terraform resource name.
Use the constant defined in the generated `*_types.go` file.

```
# Wrong – kind does not exist
kind: OpenidClientTimePolicyPermission

# Correct – use the Go type constant value
kind: ClientTimePolicy
```

The generated API group for cluster-scoped resources is
`<group>.keycloak.crossplane.io`; for namespaced resources it is
`<group>.keycloak.m.crossplane.io`.

Symptoms: `no matches for kind` error in chainsaw/uptest logs, or the
resource is immediately rejected by the API server with an unknown-kind error.

### 2. Wrong Field Names

`spec.forProvider` fields come directly from the Terraform schema, not from
human-readable names. Common mismatches:

| Intended field | Actual field in CRD |
|---------------|---------------------|
| `clientRef` | `resourceServerIdRef` (for `resource_server_id`) |
| `realmRef` | `realmIdRef` |
| `clientScopesRef` | `scope[].idRef` (resolved from `ClientAuthorizationScope` name) |
| `expiresInMinutes` | does not exist; use `hour`/`hourEnd`, `notBefore`/`notOnOrAfter` |

Check the generated `zz_*_types.go` for the exact JSON field names, or look
at the matching file in `examples-generated/`.

Symptoms: strict-decoding errors (`unknown field`), or the resource is
applied but never syncs because required fields are missing.

### 3. Missing or Unresolvable Cross-Resource References

A resource stuck in `WaitingForReferencedResourceReady` means one of its
`*Ref` fields cannot be resolved because the referenced resource does not
exist or is not Ready.

Common mistakes:

- Referencing a resource by name that is not created by any earlier demo file
  or by `000-init.yaml`.
- Using `realmRef.name: dev` in a namespaced demo (should be `dev-ns` if that
  is the name of the namespaced realm resource).

Symptoms: resource remains `Synced=False` with message
`WaitingForReferencedResourceReady` or `cannot resolve references`.

### 4. Required Authorization Not Enabled on the Client

Authorization resources (`ClientAuthorizationScope`, `ClientAggregatePolicy`,
`ClientTimePolicy`, etc.) require the target Keycloak client to have
`authorization` enabled. In the demo files the `test` client (defined in
`040-oidc-clients.yaml`) has `authorization.policyEnforcementMode: PERMISSIVE`.
If you reference a different client without authorization enabled Keycloak
will return a 404/400 error and the resource will stay unsynced.

### 5. Status Conditions Not Yet Available

Immediately asserting `Ready=True` after `kubectl apply` can fail because
Crossplane may not have written the initial status conditions yet. Chainsaw
JMESPath assertions on `status.conditions` will error if the field is `nil`.

Mitigation: add a short `wait` step before asserting conditions, or check
only that the object exists.

### 6. Race: CRD Not Established Before Test Applies Resources

If a new resource type is registered just before uptest runs, the Kubernetes
API discovery cache may not yet include it. `setup.sh` already waits for all
`ManagedResourceDefinitions` to be `Established`, but newly added CRDs that
arrive very late can still hit this window.

Symptoms: `no matches for kind` even though the kind name is correct.

Mitigation: the `setup.sh` wait loop handles the common case; if a specific
CRD keeps racing, add it to the wait list explicitly.

### 7. Optional Field That Keycloak Nevertheless Requires

Some Terraform fields are optional in the schema (and therefore optional in the
CRD) but are always sent to Keycloak, which then fails to parse the empty
value. `ClientTimePolicy` is the canonical case: omitting `notBefore` /
`notOnOrAfter` yields

```text
400 Bad Request: {"error":"Unable not parse a date using format []"}
```

Mitigation: set both fields (format `yyyy-MM-dd HH:mm:ss`), as the upstream
Terraform provider's own acceptance test does.

### 8. Lookup Helper Not Recognising "Not Found"

Resources wired to `lookup.BuildIdentifyingPropertiesLookup` call an upstream
`Get...ByName` client function. Several of those return a plain
`fmt.Errorf("no ... with name %s found", ...)` rather than a typed not-found
error. If the helper in `config/<group>/config.go` does not treat that message
as "not found", the very first reconcile fails with

```text
connect failed: cannot initialize the Terraform plugin SDK async external client:
failed to get the extended parameters for resource "/<name>": cannot get ID: ...
```

and the resource is never created, blocking everything that references it.

Mitigation: check the upstream function's not-found error string and return
`("", nil)` for it — see `getAuthzScopeIDByIdentifyingProperties` in
`config/openidclient/config.go`.

## Methodology: Writing a Demo File

Follow these steps when writing a new demo file.

### Step 1 – Identify the correct API kind

```bash
# Find the generated types file
ls apis/cluster/openidclient/v1alpha1/ | grep <resource>

# Confirm the Kind constant
grep 'Kind\s*=' apis/cluster/openidclient/v1alpha1/zz_<resource>_types.go
```

Use the generated `examples-generated/` file as a reference for fields and
structure.

### Step 2 – Check required fields

```bash
grep 'kubebuilder:validation:XValidation' \
  apis/cluster/openidclient/v1alpha1/zz_<resource>_types.go
```

Every field listed in a `required parameter` message must be present in
`spec.forProvider`.

### Step 3 – Verify prerequisites

For each `*Ref` field, confirm the referenced resource:

1. Exists in a lower-numbered demo file **or** in `000-init.yaml`.
2. Is in the same namespace for namespaced demos.
3. Uses the correct API group (`*.keycloak.crossplane.io` for cluster,
   `*.keycloak.m.crossplane.io` for namespaced).

Demos must be self-contained: every referenced object is created by the demo
itself (or a lower-numbered one), never by hardcoding a Keycloak UUID. If a
field only accepts raw IDs, configure a cross-resource reference for it in
`config/<group>/config.go`. When a single Terraform field accepts IDs of
several different resource types (e.g. `keycloak_openid_client_aggregate_policy`'s
`policies`), use the `config/multitypes` helpers to expose one strongly-typed
list field per referenceable type — see `keycloak_openid_client_client_policy`
(`clients`/`saml_clients`) and `keycloak_openid_client_aggregate_policy`
(`timePolicies`, `rolePolicies`, …) for examples.

### Step 4 – Name resources to avoid collisions

Use unique, descriptive names that will not clash with other demo files. For
example, prefix with the demo number: `064-authz-scope` instead of
`manage:users`.

### Step 5 – Add to `cases.txt`

Append both the `basic` and `namespaced` paths to `cluster/test/cases.txt`
in descending order (newest at the top of the namespaced block, and at the
top of the basic block, since the file is sorted descending by number).

### Step 6 – Test locally

```bash
# Apply the demo manifest and watch for readiness
kubectl apply -f dev/demos/basic/<NNN>-<name>.yaml
kubectl get -f dev/demos/basic/<NNN>-<name>.yaml -w
```

Check for stuck resources:

```bash
kubectl describe <kind> <name> | grep -A5 'Status\|Message\|Reason'
```

Common resolution: look for `WaitingForReferencedResourceReady` or strict
decode errors, then fix the field names or references as described above.

## Namespaced vs. Cluster-Scoped Demos

| Concern | Cluster (`basic/`) | Namespaced (`namespaced/`) |
|--------|-------------------|--------------------------|
| API group | `*.keycloak.crossplane.io` | `*.keycloak.m.crossplane.io` |
| Realm ref name | `dev` | `dev-ns` |
| ProviderConfig kind | *(omit kind field)* | `kind: ProviderConfig` |
| Resource namespace | *(none)* | `namespace: dev-ns` |

> **Note:** Do **not** use `providerConfigRef.kind: ClusterProviderConfig` in
> namespaced demos; the namespaced provider expects a `ProviderConfig`
> (namespace-scoped).

## Known Limitations

- E2E tests only cover resources listed in `cluster/test/cases.txt`.
- Chainsaw JMESPath assertions on `status.conditions` can fail if conditions are `nil` immediately after apply; add a wait step or assert only object existence first.

## Test Selection and the Resource Index

CI does not always run every demo. `scripts/e2e_dag.py select` classifies every
changed file of a pull request with exactly one named rule, and the highest
resulting tier wins:

| Tier | Rules (matched in this order) | Scope |
|------|-------------------------------|-------|
| `skip` | `documentation` (`docs/`), `markdown/images` (`*.md`, `*.png`, `*.jpg`, `*.svg`), `unrelated workflow` (any `.github/` file other than `ci.yml`), `helper script` (`scripts/`) | no e2e |
| `targeted` | `generated controller` (`internal/controller/**/zz_*.go`), `API types` (`apis/`), `resource config` (`config/`), `CRD schema` (`package/crds/`), `generated example` / `example manifest`, `demo manifest` (`dev/demos/`), `e2e harness` (`cluster/test/`), plus any unclassified path as a safe fallback | resource-focused DAG slice (falling back to API groups only when the path is too broad) × latest Keycloak only |
| `full` | `go module` (`go.mod`, `go.sum`), `build system` (`Makefile`, `build/`), `provider runtime code` (`internal/`, `cmd/` — excluding generated controllers), `CI workflow` (`.github/workflows/ci.yml`), `e2e environment` (`dev/` outside `dev/demos/`), or any non-PR event | all demos × all Keycloak versions |

Generated per-resource controller code (`internal/controller/<scope>/<group>/<resource>/zz_controller.go`
and `zz_setup.go`) is deliberately **not** full-tier: it belongs to a single API
group and is treated like `apis/<group>/`. Only hand-written provider runtime
code forces a full run.

The `detect-noop` job determines the tier (and therefore the Keycloak version
matrix). The `e2e-tests` job then has a **Calculate E2E test selection** step
that resolves the concrete demo list for the run.

Both steps compute the changed files against `main` (`git merge-base
origin/main HEAD`), never against the pull request's recorded base SHA, so the
selection stays correct even when the branch is behind or the base ref moves.

### Proof

Every selection is accompanied by a proof written to the job summary
(`--proof-file`), so a full run is never unexplained. It lists:

1. each changed file → the rule it matched → the tier that rule implies, with
   the files that determined the final tier marked `*`;
2. the touched resources (and touched API groups as fallback context), plus
   which changed paths implied them;
3. every selected demo with the reason it is in the set — `defines Kind (group)`,
   `uses Kind (group)`, `uses API group 'x'`, `changed directly`,
   `prerequisite of <demo>` or `depends on <demo>`.

Reproduce it locally:

```bash
git diff --name-only $(git merge-base origin/main HEAD) HEAD | \
  python3 scripts/e2e_dag.py select --changed-files -
```

The graph is derived from the demo YAML itself: top-level `apiVersion:` and
`kind:` lines identify resources, and `*Ref:` → `name:` lookups give
cross-demo edges within the same demo variant (`basic/` or `namespaced/`). No
manual mapping file is maintained. Only genuine infrastructure names
(`keycloak-provider-config`, `crossplane-system`) are ignored when building
those edges — realm names such as `dev`/`dev-ns` are real dependencies, so a
targeted run always pulls in the realm demo that defines them.

The selected demo list is emitted in the same order as `cluster/test/cases.txt`
— dependents first, prerequisites last — because uptest deletes the examples in
the order they are listed. Deleting a prerequisite (for example the realm)
before its dependents leaves them unable to resolve their references and blocks
teardown, whereas applying in that order is safe since Crossplane retries
reference resolution until the prerequisite exists.

`make generate` refreshes `cluster/test/e2e-index.json`, which answers
"which e2e test uses resource X?":

```bash
jq '.resources["ClientTimePolicy (openidclient)"]' cluster/test/e2e-index.json
```

The same file holds the demo DAG under `.demos`. It is committed, so
`make check-diff` fails if it is stale.

