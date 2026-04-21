# Assessment: `Client.spec.forProvider` drift on `authenticationFlowBindingOverrides` (ArgoCD reconciliation loop)

Tracking issue: *Client forProvider specs added dynamically, breaking deployment
with ArgoCD* — reported against `crossplane-contrib/provider-keycloak` v2.1.0.
Update the link below to the actual issue URL once known:
`https://github.com/crossplane-contrib/provider-keycloak/issues/<NN>`.

This document is **research / assessment only** — no code changes are proposed in
this commit. The intent is to give the maintainers a concrete starting point
for choosing the fix, since the right answer has API-compatibility implications.

---

## 1. What the user observes

Applying:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
spec:
  forProvider:
    authenticationFlowBindingOverrides:
      - browserIdSelector:
          matchLabels:
            my-label: "something"
```

…ends up, after one reconcile, as:

```yaml
spec:
  forProvider:
    authenticationFlowBindingOverrides:
      - browserId: <uuid>           # <-- written by the controller
        browserIdRef:
          name: my-something-flow   # <-- written by the controller
        browserIdSelector:
          matchLabels:
            my-label: "something"
```

ArgoCD sees the live object diverging from the stored manifest on every sync
and enters an endless OutOfSync loop.

## 2. Why this happens (upjet / Crossplane behaviour)

Two independent mechanisms in upjet/crossplane-runtime cause writes back into
`spec.forProvider`. Both apply to this resource.

### 2.1 Reference resolution writes the resolved value back to `forProvider`

`config/openidclient/config.go` declares a cross-resource reference for the
nested `authentication_flow_binding_overrides.browser_id` field:

```go
r.References["authentication_flow_binding_overrides.browser_id"] = config.Reference{
    TerraformName: "keycloak_authentication_flow",
}
r.References["authentication_flow_binding_overrides.direct_grant_id"] = config.Reference{
    TerraformName: "keycloak_authentication_flow",
}
```

This causes upjet to generate `BrowserIDRef` / `BrowserIDSelector` (and the
`DirectGrant*` equivalents) in **both** the InitParameters and the Parameters
struct of the nested type
(`apis/{cluster,namespaced}/openidclient/v1alpha1/zz_client_types.go`):

```go
// AuthenticationFlowBindingOverridesParameters – this lives under spec.forProvider
BrowserID         *string      `json:"browserId,omitempty" tf:"browser_id,omitempty"`
BrowserIDRef      *v1.Reference`json:"browserIdRef,omitempty" tf:"-"`
BrowserIDSelector *v1.Selector `json:"browserIdSelector,omitempty" tf:"-"`
```

Then the generated resolver (`zz_generated.resolvers.go`) writes the resolved
ID and the canonical `Ref` back into the same `forProvider` slice element on
every reconcile:

```go
mg.Spec.ForProvider.AuthenticationFlowBindingOverrides[i3].BrowserID    = reference.ToPtrValue(rsp.ResolvedValue)
mg.Spec.ForProvider.AuthenticationFlowBindingOverrides[i3].BrowserIDRef = rsp.ResolvedReference
```

Because these fields are part of `spec`, the writes are persisted to the API
server as part of the managed resource — and ArgoCD diffs against them.

### 2.2 Late-initialization copies observed values into `forProvider`

Independently of references, upjet's default reconciler runs
`LateInitialize`, which copies any value present in the observed Terraform
state into the corresponding `forProvider` field if that field is currently
`nil`. The Terraform schema marks `browser_id` / `direct_grant_id` as
`Optional+Computed`, so they always come back populated and get late-init'd
even when the user never set a selector/ref.

The same pattern is the reason the SAML `Client`
(`apis/.../samlclient/v1alpha1`) shows drift on these fields too, even though
no `References` are configured for it — the LateInit alone is enough.

So the bug, from the user's point of view, has **two contributors**:

1. Reference resolution writing `BrowserID` and `BrowserIDRef` into
   `forProvider` (only when a selector/ref is used).
2. LateInit writing `BrowserID` / `DirectGrantID` into `forProvider` after the
   first successful apply (always).

Any complete fix must address both, otherwise users will still see drift on
the plain string fields even if we silence the reference write-back.

## 3. How other upjet-based providers handle this

Surveying provider-aws, provider-azure, provider-gcp, and the various
upjet-based community providers:

* All of them have **the same upstream behaviour** for cross-resource
  references — the resolved value lands in `spec.forProvider`. This is by
  design in `crossplane-runtime`'s `reference` package; see
  `crossplane-runtime/pkg/reference/reference.go` (the `Resolve` method
  returns `ResolvedValue` and `ResolvedReference`, which the *generated*
  resolver assigns into `forProvider`).
* The Crossplane recommendation for ArgoCD users is to add the upjet-managed
  fields to ArgoCD's `ignoreDifferences`. The minimal practical recipe is:

  ```yaml
  ignoreDifferences:
    - group: openidclient.keycloak.crossplane.io
      kind: Client
      jqPathExpressions:
        - '.spec.forProvider.authenticationFlowBindingOverrides[].browserId'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].browserIdRef'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].directGrantId'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].directGrantIdRef'
  ```

* The forward-looking solution adopted across upjet providers is the
  **management-policies / `initProvider` split** (alpha in crossplane 1.13,
  beta in 1.15, stable in 1.17). With this enabled, users put values into
  `spec.initProvider` (which is *not* part of the desired-state diff after
  the first reconcile) and set
  `spec.managementPolicies: ["Observe","Create","Update"]` — i.e. drop
  `LateInitialize`. Then both the reference write-backs that target
  `initProvider` *and* LateInit are silenced. provider-aws and provider-azure
  have shipped docs recommending this pattern explicitly.

  Note from the generated code: upjet already generates `InitParameters`
  variants of every field for this provider (see e.g.
  `AuthenticationFlowBindingOverridesInitParameters` in
  `zz_client_types.go`), and the resolver also iterates
  `mg.Spec.InitProvider.AuthenticationFlowBindingOverrides`. So the
  infrastructure for the "modern" workaround is already in place — it just
  isn't documented for end users.

## 4. Options for fixing this in `provider-keycloak`

Ordered from least to most invasive.

### Option A — Documentation only (recommended short-term)

Add a section to the README / a docs page explaining the two write-back
mechanisms and the two ways to neutralise them:

1. **ArgoCD users**: use `ignoreDifferences` on the affected fields (sample
   snippets above; one snippet for `Client`, one for `SamlClient`, and a
   generic one for any reference).
2. **All users**: opt in to the management-policies split — use
   `spec.initProvider` for the reference selectors and set
   `spec.managementPolicies: ["Observe","Create","Update"]`. Because LateInit
   is dropped, no further writes hit `spec.forProvider`.

Pros: zero API change, zero risk; works today.
Cons: doesn't make the out-of-the-box experience better.

### Option B — Stop writing references into `forProvider`, keep `forProvider` only as the user authored it

Concretely: in `config/openidclient/config.go`, reshape the schema so that
`browser_id` / `direct_grant_id` are no longer Optional fields on
`forProvider`. The mechanical change is:

```go
authBlock := r.TerraformResource.Schema["authentication_flow_binding_overrides"].
    Elem.(*schema.Resource).Schema
authBlock["browser_id"].Optional = false
authBlock["browser_id"].Computed = true
authBlock["direct_grant_id"].Optional = false
authBlock["direct_grant_id"].Computed = true
```

…then re-run `make generate`. The user would have to switch to the
management-policies split (set the value via `spec.initProvider`) since the
field would no longer exist on `forProvider`.

Pros: removes the drift at its source for new users.
Cons: **breaking API change**. Existing CRs that already store
`forProvider.authenticationFlowBindingOverrides[].browserId` would fail
schema validation after upgrade until they migrate. We would have to ship a
v1beta1 type and a conversion webhook (provider-keycloak does not have one
today).

### Option C — Stop LateInit from writing the leaf fields back (non-breaking)

Add a `LateInitializer` config that explicitly skips `browser_id` and
`direct_grant_id`. Upjet exposes
`config.Resource.LateInitializer.IgnoredFields` exactly for this use case:

```go
r.LateInitializer = config.LateInitializer{
    IgnoredFields: []string{
        "authentication_flow_binding_overrides.browser_id",
        "authentication_flow_binding_overrides.direct_grant_id",
    },
}
```

This addresses contributor #2 (LateInit) without changing the CRD schema.
Existing CRs continue to validate. Users who use selectors (contributor #1)
still see the resolved `browserId`/`browserIdRef` written into `forProvider`
on every reconcile — that one is unavoidable with current upjet — but
they're the smaller subset of users and can use the targeted ArgoCD
`ignoreDifferences` snippet from §3.

Pros: non-breaking; trivial to implement; addresses the most common drift
source for both `Client` and `SamlClient`; pattern is reusable for any
other Optional+Computed field that bites users similarly.
Cons: doesn't fully eliminate drift when selectors are used.

### Effect of "moving to status only" for *existing* resources

Per the agent instructions — "Research the effect of moving to Status only
for existing resources":

* "Status only" here means removing `browserId`/`directGrantId` (and the
  `*Ref`/`*Selector` siblings) from the `Parameters` (forProvider) struct
  entirely, leaving them only on the `Observation` (atProvider) struct.
* Concrete consequences for existing in-cluster `Client` objects:
  1. **CRD schema validation breaks on read**. CR objects whose
     `spec.forProvider.authenticationFlowBindingOverrides[].browserId` is
     populated (which, per this very bug, is *all* of them) would fail the
     OpenAPI validation embedded in the new CRD. kube-apiserver rejects the
     stored object on next write, and Crossplane's reconciler can't update
     status either.
  2. **No automatic migration**. Crossplane managed-resource APIs do not
     ship a conversion webhook by default; provider-keycloak has none. The
     operator would have to either bump to a new API version (`v1beta1`)
     and run both versions in parallel with a conversion webhook, or force
     users to delete + re-create with `deletionPolicy: Orphan` and
     re-import — disruptive and unattractive.
  3. **Terraform state churn**. Even with the schema change, the underlying
     Terraform schema still treats `browser_id` as Optional+Computed, so
     upjet would keep round-tripping the value through `atProvider` (which
     is fine) but would also, on every Apply, send an empty
     `authentication_flow_binding_overrides[].browser_id` to Terraform
     because nothing in `forProvider` populates it any more. That removes
     the binding override server-side. **This is a behaviour-breaking
     change unless `initProvider` is also wired up for these fields.**

Conclusion: a "status-only" migration is not feasible without (a) a new API
version + conversion webhook, and (b) requiring users to switch to the
management-policies split for actually configuring the override. It is **not
a drop-in fix**.

## 5. Recommendation

For **v2.1.x patch**: ship Option A (docs) plus Option C (LateInit ignore for
the affected nested fields on `openidclient.Client` and
`samlclient.Client`). This:

* Removes the always-on drift caused by LateInit for users who *don't* use
  selectors (the majority case).
* Is non-breaking — no CRD schema change, no migration.
* Documents the ArgoCD `ignoreDifferences` recipe for users who *do* use
  selectors (the minority case where reference write-back is unavoidable
  with current upjet).

For **v3 / next major**: pursue Option B together with a conversion
webhook, aligned with the wider Crossplane move to management-policies +
`initProvider` as the default.

## 6. Files / locations to change for Option C (when the maintainer green-lights it)

* `config/openidclient/config.go`
  Inside the `keycloak_openid_client` configurator, set:
  ```go
  r.LateInitializer = config.LateInitializer{
      IgnoredFields: []string{
          "authentication_flow_binding_overrides.browser_id",
          "authentication_flow_binding_overrides.direct_grant_id",
      },
  }
  ```
* `config/samlclient/config.go`
  Same change for `keycloak_saml_client`.
* Re-run `make generate` to refresh `apis/.../zz_*` and `examples-generated/`.
* Add a manual verification example under `examples/openidclient/` (this repo
  currently has no automated e2e harness — only `internal/clients/keycloak_test.go`
  exists as a Go unit test, and `.github/workflows/e2e.yaml.disabled` is
  disabled). The example should:
  1. Apply a `Client` with `authenticationFlowBindingOverrides.browserIdSelector`.
  2. Wait for `Synced=True` & `Ready=True`.
  3. Re-read the object and confirm that `spec.forProvider.authenticationFlowBindingOverrides[].browserId`
     and `directGrantId` are *not* repopulated by LateInit (the only acceptable
     remaining write-back is the resolver setting `browserIdRef` /
     `directGrantIdRef`, which can be silenced with the ArgoCD
     `ignoreDifferences` snippet from §3).

## 7. References

* Crossplane docs — *Managed resources / Management policies*:
  https://docs.crossplane.io/latest/concepts/managed-resources/#management-policies
* upjet docs — *Configuring a resource* (References, LateInit):
  https://github.com/crossplane/upjet/blob/main/docs/configuring-a-resource.md
* crossplane-runtime — `reference` package source:
  https://github.com/crossplane/crossplane-runtime/blob/main/pkg/reference/reference.go
