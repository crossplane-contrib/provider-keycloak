# Release Notes — provider-keycloak v3.0.0

These notes supplement the auto-generated changelog.

## ⚠️ Breaking Change: `clientSecretWoVersion` is now a string

`terraform-provider-keycloak` v5.9.0 corrected `client_secret_wo_version` from
`TypeInt` to `TypeString`. Regenerating the provider therefore changed the CRD
field type of `clientSecretWoVersion` on two kinds:

| Group | Kind | Frozen version | New storage version |
|-------|------|----------------|---------------------|
| `openidclient.keycloak.crossplane.io` / `openidclient.keycloak.m.crossplane.io` | `Client` | `v1alpha1` (`number`) | `v1alpha2` (`string`) |
| `oidc.keycloak.crossplane.io` / `oidc.keycloak.m.crossplane.io` | `IdentityProvider` | `v1alpha1` (`number`) | `v1alpha2` (`string`) |

Because a field type cannot change in place without breaking every object
already stored in etcd, `v1alpha1` has been frozen with the original `number`
type and a new `v1alpha2` was added as the storage version. A **CRD conversion
webhook** translates between the two in both directions.

### What you have to do

**Nothing, immediately.** `v1alpha1` is still served, so existing manifests and
existing objects keep working: reading a v2.x object through `v1alpha1` still
yields a number, and reading it through `v1alpha2` yields the equivalent
string.

When you are ready, migrate your manifests:

```diff
-apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
+apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
 kind: Client
 metadata:
   name: my-client
 spec:
   forProvider:
     clientId: my-client
-    clientSecretWoVersion: 1
+    clientSecretWoVersion: "1"
```

The same applies to `oidc/IdentityProvider` and to the namespaced
(`*.keycloak.m.crossplane.io`) variants.

> `v1alpha1` will be removed in a future major release. New manifests should
> use `v1alpha2` for these two kinds. All other kinds are unaffected and stay
> on `v1alpha1`.

### Requirements for the conversion webhook

The webhook is served by the provider itself and is only started when the
provider pod has TLS server certificates, which Crossplane mounts as long as
`webhooks.enabled` is `true` (the default) in your Crossplane installation. If
webhooks are disabled, the provider revision will not become healthy, because
the CRDs declare a conversion webhook without a CA bundle.

If you use GitOps tooling that prunes or normalises CRDs (Argo CD, Flux), make
sure it does not strip `spec.conversion` from the provider's CRDs — the
`caBundle` is injected by the Crossplane package manager at runtime.

## Other Changes

- The provider now calls `SetupWebhookWithManager` for both the cluster-scoped
  and the namespaced API groups; conversion webhooks were previously generated
  but never registered.
- `make generate` runs `cmd/crdconversion`, which adds the
  `spec.conversion.strategy: Webhook` stanza to every CRD that serves more than
  one version. Without the stanza the API server silently falls back to the
  `None` strategy and hands back untranslated objects.
- Generated API types of previous versions are no longer deleted by
  `make generate`; only `zz_generated.conversion_{hubs,spokes}.go` and
  `zz_generated.resolvers.go` are regenerated from scratch.
- New `make uptest-conversion` / `make e2e-conversion` targets run the
  Chainsaw upgrade tests in `cluster/test/conversion`, which create an object
  through `v1alpha1` in a real cluster and read it back through `v1alpha2`.
- `make crddiff` now fails on breaking CRD schema changes; set
  `CRDDIFF_ALLOW_BREAKING=true` on a major release branch such as this one.

## Dependency Updates

- Bumped upstream `terraform-provider-keycloak` to **v5.9.0**

---

## Upgrade Notes

1. Upgrade the provider package to `v3.0.0`.
2. Verify that the CRDs serve both versions and that the conversion webhook is
   wired up:

   ```shell
   kubectl get crd clients.openidclient.keycloak.crossplane.io \
     -o jsonpath='{.spec.conversion.strategy}{"\n"}{range .spec.versions[*]}{.name}{" storage="}{.storage}{"\n"}{end}'
   ```

   This must print `Webhook`, `v1alpha1 storage=false` and
   `v1alpha2 storage=true`.
3. Migrate `Client` and `IdentityProvider` manifests to `v1alpha2` and quote
   `clientSecretWoVersion` at your own pace.
