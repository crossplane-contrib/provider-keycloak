# provider-keycloak v3.0.0

Compared against [v2.24.1](https://github.com/crossplane-contrib/provider-keycloak/releases/tag/v2.24.1).

Every claim below links to the code or PR that proves it.

## Breaking change: `clientSecretWoVersion` is a string

Upstream changed `client_secret_wo_version` from `TypeInt` to `TypeString`
([terraform-provider-keycloak#1559](https://github.com/keycloak/terraform-provider-keycloak/pull/1559)),
released in [v5.9.0](https://github.com/keycloak/terraform-provider-keycloak/releases/tag/v5.9.0)
and pinned in [`Makefile:17`](Makefile#L17). A CRD field type cannot change in
place, so the affected kinds got a new API version and a conversion webhook
([#679](https://github.com/crossplane-contrib/provider-keycloak/pull/679)).

| Kind | `v1alpha1` (served) | `v1alpha2` (storage) |
|------|--------------------|----------------------|
| `openidclient.keycloak.crossplane.io/Client` | `type: number` ([`clients.yaml:381-385`](package/crds/openidclient.keycloak.crossplane.io_clients.yaml#L381-L385)) | `type: string` ([`clients.yaml:2009-2013`](package/crds/openidclient.keycloak.crossplane.io_clients.yaml#L2009-L2013)) |
| `oidc.keycloak.crossplane.io/IdentityProvider` | `type: number` | `type: string` |

The same applies to the namespaced variants (`*.keycloak.m.crossplane.io`).
Conversion is registered by
[`config/conversion/conversion.go`](config/conversion/conversion.go)
(`BumpVersionForIntToStringChange`) and wired into both kinds in
[`config/openidclient/config.go:50`](config/openidclient/config.go#L50) and
[`config/oidc/config.go:24`](config/oidc/config.go#L24).

### What you have to do

Nothing at upgrade time — `v1alpha1` stays served and existing objects convert
in both directions. Migrate at your own pace:

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

Verify the webhook is wired up after upgrading:

```shell
kubectl get crd clients.openidclient.keycloak.crossplane.io \
  -o jsonpath='{.spec.conversion.strategy}{"\n"}{range .spec.versions[*]}{.name}{" storage="}{.storage}{"\n"}{end}'
# Webhook
# v1alpha1 storage=false
# v1alpha2 storage=true
```

Requirements and constraints:

- The webhook is served by the provider pod and only starts when Crossplane
  mounts TLS certificates, i.e. `webhooks.enabled=true` (Crossplane default).
  With webhooks disabled the provider revision stays unhealthy, because the
  CRDs declare `spec.conversion.strategy: Webhook`
  ([`clients.yaml:9-15`](package/crds/openidclient.keycloak.crossplane.io_clients.yaml#L9-L15)).
- GitOps tooling must not prune `spec.conversion` from the CRDs; the
  `caBundle` is injected by the Crossplane package manager at runtime.
- `v1alpha1` will be removed in a future major release.

## New managed resources

19 new kinds, each available cluster-scoped (`*.keycloak.crossplane.io`) and
namespaced (`*.keycloak.m.crossplane.io`), all at `v1alpha1`.

### Fine-grained admin permissions (FGAPv2)

| Kind | API group | Terraform resource | PR |
|------|-----------|--------------------|----|
| `AdminPermissions` | `user` | `keycloak_users_admin_permissions` | [#711](https://github.com/crossplane-contrib/provider-keycloak/pull/711) |
| `AdminPermissions` | `role` | `keycloak_role_admin_permissions` | [#717](https://github.com/crossplane-contrib/provider-keycloak/pull/717) |
| `AdminPermissions` | `group` | `keycloak_group_admin_permissions` | [#718](https://github.com/crossplane-contrib/provider-keycloak/pull/718) |
| `ClientAdminPermissions` | `openidclient` | `keycloak_openid_client_admin_permissions` | [#720](https://github.com/crossplane-contrib/provider-keycloak/pull/720) |

These require Keycloak 26.2+ with the `admin-fine-grained-authz:v2` feature and
a realm with `adminPermissionsEnabled: true`; Keycloak then creates the realm's
`admin-permissions` client that acts as resource server
([`examples/user/adminpermissions.yaml`](examples/user/adminpermissions.yaml)).

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: example-adminpermissions-realm
spec:
  forProvider:
    realm: example-adminpermissions
    enabled: true
    adminPermissionsEnabled: true
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: AdminPermissions
metadata:
  name: example-auditors-view-users
spec:
  forProvider:
    name: auditors-can-view-users
    description: Auditors can view all users
    decisionStrategy: UNANIMOUS
    realmIdRef:
      name: example-adminpermissions-realm
    scopes:
      - view
  providerConfigRef:
    name: keycloak-provider-config
```

Full examples, including permissions scoped to specific clients:
[`examples/role/adminpermissions.yaml`](examples/role/adminpermissions.yaml),
[`examples/group/adminpermissions.yaml`](examples/group/adminpermissions.yaml),
[`examples/openidclient/clientadminpermissions.yaml`](examples/openidclient/clientadminpermissions.yaml).

### Client authorization

| Kind | Terraform resource | PR |
|------|--------------------|----|
| `ClientAggregatePolicy` | `keycloak_openid_client_aggregate_policy` | [#645](https://github.com/crossplane-contrib/provider-keycloak/pull/645) |
| `ClientAuthorizationClientScopePolicy` | `keycloak_openid_client_authorization_client_scope_policy` | [#645](https://github.com/crossplane-contrib/provider-keycloak/pull/645) |
| `ClientAuthorizationScope` | `keycloak_openid_client_authorization_scope` | [#645](https://github.com/crossplane-contrib/provider-keycloak/pull/645) |
| `ClientTimePolicy` | `keycloak_openid_client_time_policy` | [#645](https://github.com/crossplane-contrib/provider-keycloak/pull/645) |
| `ClientJsPolicy` | `keycloak_openid_client_js_policy` | [#705](https://github.com/crossplane-contrib/provider-keycloak/pull/705) |
| `ClientAuthorizationPolicy` | `keycloak_generic_client_authorization_policy` | [#704](https://github.com/crossplane-contrib/provider-keycloak/pull/704) |

All in group `openidclient.keycloak.crossplane.io`.
`ClientAuthorizationPolicy` is the escape hatch for policy types that have no
dedicated kind — custom SPI policy providers, referenced by the `type` returned
from `PolicyProviderFactory.getId()`
([`examples/openidclient/clientauthorizationpolicy.yaml`](examples/openidclient/clientauthorizationpolicy.yaml)).

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientJsPolicy
metadata:
  name: example-js-policy
spec:
  forProvider:
    name: example-js-policy
    code: script-example-js-policy.js
    decisionStrategy: UNANIMOUS
    logic: POSITIVE
    realmIdRef:
      name: example-realm
    resourceServerIdRef:
      name: example-client
  providerConfigRef:
    name: keycloak-provider-config
```

### Realm keystores

`keycloak_realm_keystore_rsa` was the only keystore resource in v2.24.1
([`package/crds` at v2.24.1](https://github.com/crossplane-contrib/provider-keycloak/tree/v2.24.1/package/crds)).
[#653](https://github.com/crossplane-contrib/provider-keycloak/pull/653) adds
the remaining five, all in group `realm.keycloak.crossplane.io`:

| Kind | Terraform resource |
|------|--------------------|
| `KeystoreAesGenerated` | `keycloak_realm_keystore_aes_generated` |
| `KeystoreEcdsaGenerated` | `keycloak_realm_keystore_ecdsa_generated` |
| `KeystoreHMACGenerated` | `keycloak_realm_keystore_hmac_generated` |
| `KeystoreJavaKeystore` | `keycloak_realm_keystore_java_keystore` |
| `KeystoreRsaGenerated` | `keycloak_realm_keystore_rsa_generated` |

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: KeystoreRsaGenerated
metadata:
  name: rsa-generated-keystore
spec:
  forProvider:
    active: true
    algorithm: RS256
    enabled: true
    keySize: 2048
    name: crossplane-rsa-generated-key
    priority: 100
    realmId: dev
  providerConfigRef:
    name: keycloak-provider-config
```

Covered end-to-end by
[`dev/demos/basic/005-realm-keystores-comprehensive.yaml`](dev/demos/basic/005-realm-keystores-comprehensive.yaml).

### Identity providers

Group `oidc.keycloak.crossplane.io`:

| Kind | Terraform resource | PR |
|------|--------------------|----|
| `FacebookIdentityProvider` | `keycloak_oidc_facebook_identity_provider` | [#648](https://github.com/crossplane-contrib/provider-keycloak/pull/648) |
| `GithubIdentityProvider` | `keycloak_oidc_github_identity_provider` | [#648](https://github.com/crossplane-contrib/provider-keycloak/pull/648) |
| `MicrosoftIdentityProvider` | `keycloak_oidc_microsoft_identity_provider` | [#706](https://github.com/crossplane-contrib/provider-keycloak/pull/706) |

### Realm client registration

`realm.keycloak.crossplane.io/ClientRegistrationPolicy` from
`keycloak_realm_client_registration_policy`
([#708](https://github.com/crossplane-contrib/provider-keycloak/pull/708),
[`examples/realmclientregistrationpolicy.yaml`](examples/realmclientregistrationpolicy.yaml)):

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: ClientRegistrationPolicy
metadata:
  name: trusted-hosts
spec:
  forProvider:
    realmIdRef:
      name: basic-realm
    name: "Trusted Hosts"
    providerId: trusted-hosts
    subType: anonymous
    config:
      host-sending-registration-request-must-match: "true"
      client-uris-must-match: "true"
      trusted-hosts: example.com
  providerConfigRef:
    name: keycloak-provider-config
```

## Typed policy references

Admin-permission and permission resources take a flat list of policy IDs in
Terraform. Those lists are now exposed as one typed reference field per policy
type via `config/multitypes`, consolidated back into the original field before
the Terraform call
([#709](https://github.com/crossplane-contrib/provider-keycloak/pull/709),
[`config/openidclient/config.go:311-378`](config/openidclient/config.go#L311-L378),
[`config/user/config.go:61-128`](config/user/config.go#L61-L128)).

Added fields: `aggregatePolicies`, `clientPolicies`, `clientScopePolicies`,
`groupPolicies`, `jsPolicies`, `regexPolicies`, `rolePolicies`, `timePolicies`,
`userPolicies` (each with `*Refs` / `*Selector`).

Backwards compatible: `KeepOriginalField: true` keeps the raw `policies` list
settable for policy types without a managed resource
([`config/openidclient/config.go:311-315`](config/openidclient/config.go#L311-L315)).

```yaml
spec:
  forProvider:
    # before: raw IDs only
    policies:
      - 4b0b7b4c-4a3f-4d2e-9f24-8b7b1a5d2c11
    # now also: resolved from managed resources
    groupPoliciesRefs:
      - name: example-group-policy
```

`ClientAdminPermissions` additionally resolves `clientIds` against both OpenID
and SAML clients through `clientIdsRefs` / `samlClientIdsRefs`
([`config/openidclient/config.go:287-302`](config/openidclient/config.go#L287-L302)).

## Tooling and CI

- `make generate` now also runs doc generation, so `llms.txt` / `llms-full.txt`
  can no longer go stale ([#721](https://github.com/crossplane-contrib/provider-keycloak/pull/721)).
- `make crddiff` fails on breaking CRD schema changes;
  `CRDDIFF_ALLOW_BREAKING=true` opts out on a major release branch
  ([`Makefile:353-382`](Makefile#L353-L382)).
- Weekly workflow files an issue per Terraform resource that is not yet exposed,
  diffing `config/schema.json` against the generated `config/generated.lst`
  ([#684](https://github.com/crossplane-contrib/provider-keycloak/pull/684),
  [#697](https://github.com/crossplane-contrib/provider-keycloak/pull/697),
  [#719](https://github.com/crossplane-contrib/provider-keycloak/pull/719)).
- Scheduled upstream provider release check with issue and PR automation
  ([#702](https://github.com/crossplane-contrib/provider-keycloak/pull/702)).
- Renovate: weekly grouped updates with a 3-day minimum release age
  ([#696](https://github.com/crossplane-contrib/provider-keycloak/pull/696),
  [#693](https://github.com/crossplane-contrib/provider-keycloak/pull/693)), and
  `hugo-version` is no longer misdetected as a Go version
  ([#707](https://github.com/crossplane-contrib/provider-keycloak/pull/707)).
- Release and CI Go toolchains aligned to prevent stdlib CVE regressions
  ([#671](https://github.com/crossplane-contrib/provider-keycloak/pull/671)).
- CRD conversion chainsaw e2e tests removed
  ([#680](https://github.com/crossplane-contrib/provider-keycloak/pull/680)).

## Dependencies

| Dependency | Version | Proof |
|------------|---------|-------|
| `terraform-provider-keycloak` (schema) | 5.9.0 | [`Makefile:17`](Makefile#L17) |
| `github.com/keycloak/terraform-provider-keycloak` (Go) | `v0.0.0-20260810123218-3c42a703d62e` | [`go.mod:13`](go.mod#L13), [#699](https://github.com/crossplane-contrib/provider-keycloak/pull/699) |
| Hugo (docs) | 0.164.0 | [#674](https://github.com/crossplane-contrib/provider-keycloak/pull/674), [#682](https://github.com/crossplane-contrib/provider-keycloak/pull/682) |

## Upgrade

1. Bump the package to `v3.0.0`:

   ```yaml
   apiVersion: pkg.crossplane.io/v1
   kind: Provider
   metadata:
     name: provider-keycloak
   spec:
     package: xpkg.upbound.io/crossplane-contrib/provider-keycloak:v3.0.0
   ```

2. Confirm the conversion webhook is active (command above). If the provider
   revision does not become healthy, check that Crossplane runs with webhooks
   enabled.
3. Optionally migrate `Client` and `IdentityProvider` manifests to `v1alpha2`
   and quote `clientSecretWoVersion`.
