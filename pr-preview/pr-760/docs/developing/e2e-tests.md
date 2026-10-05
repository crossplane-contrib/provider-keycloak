# End-to-End Tests

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


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
| `cluster/test/cases-fgapv2.txt` | Ordered list of demo files of the FGAPv2 suite |
| `cluster/test/setup.sh` | Per-case setup: CRD readiness wait, timeout rewriting, ordered deletion |
| `dev/demos/basic/` | Cluster-scoped demo manifests |
| `dev/demos/namespaced/` | Namespace-scoped demo manifests |
| `dev/demos/fgapv2/` | Demos requiring fine-grained admin permissions v2 |
| `dev/demos/basic/000-init.yaml` | Prerequisite resources for cluster demos (realm, secrets, etc.) |
| `dev/demos/namespaced/000-init.yaml` | Prerequisite resources for namespaced demos |
| `cluster/test/conversion/` | Standalone chainsaw suite (not driven by uptest/`cases.txt`) for CRD conversion webhook regressions, run via `make uptest-conversion` |
| `cluster/test/restrictedrealmprovider/` | Standalone chainsaw suite for the single-realm `ProviderConfig` limitation ([#742](https://github.com/crossplane-contrib/provider-keycloak/issues/742)), run via `make uptest-restrictedrealmprovider`; see &#34;Standalone Chainsaw Suites&#34; below |

## Test Suites

The demo directory has exactly **one subdirectory per e2e suite**, and each
suite runs in its own cluster with its own Keycloak configuration:

| Suite | Demos | Case list | Keycloak features | CI job |
|-------|-------|-----------|-------------------|--------|
| regular | `dev/demos/basic/`, `dev/demos/namespaced/`, `dev/demos/orgs/` | `cluster/test/cases.txt`, `cases-kc-26.4.txt`, `cases-kc-26.5.txt`, `cases-orgs.txt` | `admin-fine-grained-authz:v1` (&#43; `organization` from 26.6) | `e2e-tests` |
| FGAPv2 | `dev/demos/fgapv2/` | `cluster/test/cases-fgapv2.txt` | `admin-fine-grained-authz:v2` | `e2e-tests-fgapv2` |

`dev/demos/orgs/` and the version-specific case lists belong to the regular
suite: they run in the same cluster and are gated by Keycloak version in the
`Makefile` (organizations need 26.6&#43;). Targeted runs always use the latest
Keycloak, so those demos can be selected freely.

The split is a hard Keycloak constraint, not a convenience: the
`admin-fine-grained-authz` feature can be enabled as **either** v1 **or** v2,
never both. Demos of one suite therefore must never end up in the other
suite&#39;s run — a v2 demo would fail on a v1 cluster and vice versa.

This is enforced structurally rather than by special-casing individual files:
`scripts/e2e_dag.py` builds one demo graph per suite (`REGULAR_VARIANTS` /
`FGAPV2_VARIANTS`) and exposes one selection command per suite (`select` and
`select-fgapv2`). Adding a suite means adding a variant tuple and a graph, not
adding exceptions.

Run the FGAPv2 suite locally:

```bash
./dev/setup_dev_environment.sh --fgap-version v2
make uptest FGAP_VERSION=v2
```

## Standalone Chainsaw Suites

Not every e2e scenario fits the demo/`cases.txt` model, which expects every
applied resource to reach Ready/Synced and then be deleted cleanly. Two kinds
of scenarios don&#39;t:

- Tests that assert an **expected failure** (e.g. a `ProviderConfig` that
  must never successfully authenticate).
- Tests that need imperative setup against the live Keycloak API before any
  Crossplane resource is applied (e.g. provisioning a service account with a
  specific, deliberately restricted set of roles).

These live as standalone `chainsaw.kyverno.io/v1alpha1 Test` resources under
`cluster/test/&lt;suite-name&gt;/`, each with its own Makefile target
(`uptest-&lt;suite-name&gt;`) and its own CI step, run directly via `chainsaw test
--test-dir cluster/test/&lt;suite-name&gt;` instead of through `uptest`. They
still require a cluster with the provider deployed and the
`keycloak-provider-config` admin `ProviderConfig` already applied — the same
prerequisites as the `uptest` target.

| Suite | Directory | Makefile target | What it covers |
|-------|-----------|-----------------|----------------|
| CRD conversion webhook | `cluster/test/conversion/` | `make uptest-conversion` | Historic stored CRD encodings converted through the live `/convert` endpoint |
| Restricted single-realm `ProviderConfig` | `cluster/test/restrictedrealmprovider/` | `make uptest-restrictedrealmprovider` | [#742](https://github.com/crossplane-contrib/provider-keycloak/issues/742): a `ProviderConfig` scoped entirely to one realm (no master-realm/realm-management roles) must fail login with a 403, a Keycloak-side constraint this provider cannot work around |

The single-realm suite&#39;s `bootstrap.sh` provisions a realm and a fully
realm-scoped service account client directly against the Keycloak API (there
is no declarative way to create a *deliberately under-privileged* service
account through the managed resources themselves), then a chainsaw `assert`
checks that a probe `Group` resource ends up `Synced=False` with the expected
initial-login `/admin/serverinfo` `403 Forbidden` message. `cleanup.sh` runs
as the assertion step&#39;s `finally` so the bootstrap-created credentials remain
available until that check completes, and the realm is still removed whether
the assertion passes or fails.

The `keycloak_version` forwarding behavior behind Configuration B from #742
is regression-tested in two ways: unit tests in `internal/clients/keycloak_test.go`
verify that the credential key is passed through to the Terraform provider,
and the existing `dev/demos/basic/087-nonmaster-provider.yaml` /
`dev/demos/namespaced/087-nonmaster-provider.yaml` demos (gated to Keycloak
&gt;= 26.4 via `cluster/test/cases-kc-26.4.txt`) exercise a non-master-realm
service account path that also depends on `keycloak_version` when Keycloak
returns an empty `systemInfo.version`.

## Adding a New Test

1. Create `dev/demos/basic/&lt;NNN&gt;-&lt;name&gt;.yaml` and/or
   `dev/demos/namespaced/&lt;NNN&gt;-&lt;name&gt;.yaml` — or, for resources requiring
   fine-grained admin permissions v2, `dev/demos/fgapv2/&lt;NNN&gt;-&lt;name&gt;.yaml`.
2. Add both paths to `cluster/test/cases.txt` (or `cases-fgapv2.txt`) in the
   correct position
   (higher numbers run first in the list; deletion is in reverse order, so
   resources that depend on earlier resources must have a higher number).
3. Make sure all prerequisites exist in earlier numbered demo files or in
   `000-init.yaml`.
4. Reference prerequisites via `*Ref` blocks, never by literal name — the
   selection DAG derives its edges from those refs.

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
`&lt;group&gt;.keycloak.crossplane.io`; for namespaced resources it is
`&lt;group&gt;.keycloak.m.crossplane.io`.

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
400 Bad Request: {&#34;error&#34;:&#34;Unable not parse a date using format []&#34;}
```

Mitigation: set both fields (format `yyyy-MM-dd HH:mm:ss`), as the upstream
Terraform provider&#39;s own acceptance test does.

### 8. Lookup Helper Not Recognising &#34;Not Found&#34;

Resources wired to `lookup.BuildIdentifyingPropertiesLookup` call an upstream
`Get...ByName` client function. Several of those return a plain
`fmt.Errorf(&#34;no ... with name %s found&#34;, ...)` rather than a typed not-found
error. If the helper in `config/&lt;group&gt;/config.go` does not treat that message
as &#34;not found&#34;, the very first reconcile fails with

```text
connect failed: cannot initialize the Terraform plugin SDK async external client:
failed to get the extended parameters for resource &#34;/&lt;name&gt;&#34;: cannot get ID: ...
```

and the resource is never created, blocking everything that references it.

Mitigation: check the upstream function&#39;s not-found error string and return
`(&#34;&#34;, nil)` for it — see `getAuthzScopeIDByIdentifyingProperties` in
`config/openidclient/config.go`.

## Methodology: Writing a Demo File

Follow these steps when writing a new demo file.

### Step 1 – Identify the correct API kind

```bash
# Find the generated types file
ls apis/cluster/openidclient/v1alpha1/ | grep &lt;resource&gt;

# Confirm the Kind constant
grep &#39;Kind\s*=&#39; apis/cluster/openidclient/v1alpha1/zz_&lt;resource&gt;_types.go
```

Use the generated `examples-generated/` file as a reference for fields and
structure.

### Step 2 – Check required fields

```bash
grep &#39;kubebuilder:validation:XValidation&#39; \
  apis/cluster/openidclient/v1alpha1/zz_&lt;resource&gt;_types.go
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
`config/&lt;group&gt;/config.go`. When a single Terraform field accepts IDs of
several different resource types (e.g. `keycloak_openid_client_aggregate_policy`&#39;s
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
kubectl apply -f dev/demos/basic/&lt;NNN&gt;-&lt;name&gt;.yaml
kubectl get -f dev/demos/basic/&lt;NNN&gt;-&lt;name&gt;.yaml -w
```

Check for stuck resources:

```bash
kubectl describe &lt;kind&gt; &lt;name&gt; | grep -A5 &#39;Status\|Message\|Reason&#39;
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

&gt; **Note:** Do **not** use `providerConfigRef.kind: ClusterProviderConfig` in
&gt; namespaced demos; the namespaced provider expects a `ProviderConfig`
&gt; (namespace-scoped).

## Known Limitations

- E2E tests only cover resources listed in `cluster/test/cases.txt` (regular
  suite) or `cluster/test/cases-fgapv2.txt` (FGAPv2 suite). Every managed
  resource must be covered by a demo — see &#34;Resource coverage gate&#34; below —
  unless it is declared in `cluster/test/uncovered-resources.txt` because
  Keycloak rejects it in a test environment.
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

Generated per-resource controller code (`internal/controller/&lt;scope&gt;/&lt;group&gt;/&lt;resource&gt;/zz_controller.go`
and `zz_setup.go`) is deliberately **not** full-tier: it belongs to a single API
group and is treated like `apis/&lt;group&gt;/`. Only hand-written provider runtime
code forces a full run.

The `detect-noop` job determines the tier (and therefore the Keycloak version
matrix). The `e2e-tests` job then has a **Calculate E2E test selection** step
that resolves the concrete demo list for the run.

Both steps compute the changed files against `main` (`git merge-base
origin/main HEAD`), never against the pull request&#39;s recorded base SHA, so the
selection stays correct even when the branch is behind or the base ref moves.

### Per-suite selection

Each suite is selected independently, so one suite can never silently disable
the other:

| Command | Output | Gates |
|---------|--------|-------|
| `select` | `full`, `skip`, or a comma-separated demo list (`basic/` &#43; `namespaced/` only) | `e2e-tests` (runs unless the tier is `skip`) |
| `select-fgapv2` | `run` or `skip` | `e2e-tests-fgapv2` |

`select-fgapv2` answers a boolean because the FGAPv2 suite is small and always
runs its full case list. It returns `run` when the change is `full`-tier, when
an FGAPv2 demo or `cases-fgapv2.txt` changed, or when a touched resource or
API group is used by an FGAPv2 demo.

A change that only touches resources covered by one suite therefore runs
exactly that suite — for example a new FGAPv2-only resource runs the FGAPv2
suite while the regular suite is skipped, and neither result is derived from
the other.

### Never zero demos

A `targeted` change set never results in an empty run. If no demo covers the
touched resources, the selection broadens to all demos of their API group; if
even that is empty it falls back to `full`. Only a `skip`-tier change set
(documentation, images, helper scripts) runs no demos at all.

### Resource coverage gate

`make e2e-cases-check` runs two gates:

1. `cluster/test/check_cases_coverage.sh` — every demo file is listed in a case
   file, and every case entry has a demo file.
2. `python3 scripts/e2e_dag.py coverage` — every managed resource CRD in
   `package/crds/` is used by at least one demo.

Adding a managed resource without an e2e demo therefore fails CI. The only
accepted exception is a resource Keycloak itself rejects in a test environment
(missing server-side artifact, removed feature, custom SPI deployment); such a
resource must be declared with its reason in
`cluster/test/uncovered-resources.txt`:

```text
Kind (group): reason
```

The gate also fails on stale entries, so an exception disappears automatically
once the resource becomes testable.

### Group-scoped configuration

`config/&lt;group&gt;/` holds the Upjet configuration of a whole API group (external
names, references, lookups). A change there cannot be attributed to a single
resource, so **all** demos of that group are selected, on top of the demos of
the concretely touched resources. Generated per-resource files
(`apis/`, `package/crds/`, `internal/controller/`) stay resource-scoped.

### Proof

Every selection is accompanied by a proof written to the job summary
(`--proof-file`), so a full run is never unexplained. It lists:

1. each changed file → the rule it matched → the tier that rule implies, with
   the files that determined the final tier marked `*`;
2. the touched resources (and touched API groups as fallback context), plus
   which changed paths implied them;
3. every selected demo with the reason it is in the set — `defines Kind (group)`,
   `uses Kind (group)`, `uses API group &#39;x&#39;`, `changed directly`,
   `prerequisite of &lt;demo&gt;` or `depends on &lt;demo&gt;`.

Reproduce it locally:

```bash
git diff --name-only $(git merge-base origin/main HEAD) HEAD | \
  python3 scripts/e2e_dag.py select --changed-files -

git diff --name-only $(git merge-base origin/main HEAD) HEAD | \
  python3 scripts/e2e_dag.py select-fgapv2 --changed-files -
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
&#34;which e2e test uses resource X?&#34;:

```bash
jq &#39;.resources[&#34;ClientTimePolicy (openidclient)&#34;]&#39; cluster/test/e2e-index.json
```

The same file holds the demo DAG under `.demos`. It is committed, so
`make check-diff` fails if it is stale.


