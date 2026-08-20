# fix(conversion): tolerate string `clientSecretWoVersion` at v1alpha1 and isolate controller startup failures

Fixes the `v3.0.0-rc.1` crash-loop root-caused in #669 ([comment](https://github.com/crossplane-contrib/provider-keycloak/issues/669#issuecomment-5354284253)).

## Problem

Main builds between the terraform-provider v5.9.0 bump (588b321, Aug 6) and the v1alpha2 split (990f048, Aug 10) typed `clientSecretWoVersion` as a **string in v1alpha1** and persisted such objects. v3 re-freezes v1alpha1 as `number`, so the conversion webhook's strict typed decode fails on those stored objects:

```
conversion webhook for oidc.keycloak.m.crossplane.io/v1alpha1, Kind=IdentityProvider failed:
json: cannot unmarshal string into Go struct field
  IdentityProviderParameters.spec.forProvider.clientSecretWoVersion of type float64
```

Every LIST of that kind then 500s, the informer never syncs, and controller-runtime's all-or-nothing `WaitForCacheSync` kills the whole manager at the cache-sync deadline — taking the conversion webhook down with it, which makes the affected objects unreadable via the API entirely. Only clusters with objects stored during that 4-day window are affected, which is why fresh installs test green.

The failure happens in the webhook's typed decode, **before** upjet's conversion chain (`RoundTrip`/`fieldTypeConverter`) runs, so no `config/conversion` registration can fix it.

## Changes

### 1. Tolerant decode of the string encoding (primary fix)

Hand-written `UnmarshalJSON` on the frozen v1alpha1 `Parameters`/`InitParameters`/`Observation` structs of `oidc/IdentityProvider` and `openidclient/Client` (cluster + namespaced) coerces a string-encoded `clientSecretWoVersion` to the numeric encoding before default decoding.

- Honored on both relevant paths: the webhook codec (sigs.k8s.io/json) and apimachinery's `FromUnstructured` used inside upjet's `RoundTrip`.
- Non-numeric strings (e.g. `"version1"`, representable in the string-window schema but not as a number) fail with a descriptive error instead of being silently dropped.
- Hand-written files in frozen spoke packages survive `make generate` (only `zz_generated.conversion_*`/`zz_generated.resolvers.go` are regenerated).

### 2. Controller startup isolation (hardening)

`internal/resilience.WrapManager` wraps the manager passed to the controller `Setup`/`SetupGated` calls: a controller whose `Start` fails (e.g. cache-sync timeout on an un-listable kind) is logged and dropped instead of aborting the manager.

- Only runnables implementing `controller.Controller` are wrapped; the webhook server, cache, CRD gate, metrics recorders and session cleanup keep fatal semantics.
- Shutdown-path errors still propagate; `NeedLeaderElection` is forwarded so runnable-group scheduling is unchanged.
- Failed controllers cannot be restarted in controller-runtime, so the log message states the kind stays unreconciled until the provider restarts.

Worst case degrades from "provider crash-loops, webhook-backed kinds unreadable, login flow outage" to "one controller idles, everything else including the conversion webhook keeps serving".

## Tests

- `config/conversion_test.go`: string-encoded stored v1alpha1 `Client`/`IdentityProvider` objects decode and upgrade to v1alpha2 through the real webhook code path; non-numeric strings are rejected with a descriptive error.
- `internal/resilience/manager_test.go`: startup failure swallowed, shutdown error propagated, success passthrough, leader-election forwarding.
- `cluster/test/conversion/` (e2e, `make uptest-conversion`, runs in the `26.7.0` CI matrix leg): the stored encodings can no longer be created through the fixed CRD schemas, so the chainsaw suite posts `ConversionReview` payloads directly to the provider's live `/convert` endpoint (service coordinates + caBundle read from the CRD, proving the package-manager injection). Covers the pre-v3.0.0 string encoding on all 4 webhook CRDs, the v2.x number encoding, the downgrade path, and malformed objects (non-numeric string, structurally broken spec) failing gracefully per-request with the webhook surviving.

## Notes for affected clusters

Clusters already broken on the rc can't `kubectl patch` (reads go through conversion) — upgrading to a build with this fix is sufficient: the decode succeeds, objects convert and get re-persisted at v1alpha2.
