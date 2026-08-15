---
sidebar_position: 1
title: ProviderConfig
description: Complete reference for ProviderConfig resource
---

# ProviderConfig Reference

The `ProviderConfig` resource stores connection details for a Keycloak instance.

## API Details

| Field | Value |
|-------|-------|
| API Group | `keycloak.crossplane.io` |
| API Version | `v1beta1` |
| Kind | `ProviderConfig` |
| Scope | Cluster |

## Specification

```yaml
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-provider-config
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-credentials    # Name of the Secret
      key: credentials              # Key within the Secret
      namespace: crossplane-system  # Namespace of the Secret
```

## Credential Source Options

### JSON Format (Single Key)

The most common approach — all settings in a single JSON object:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: |
    {
      "client_id": "admin-cli",
      "username": "admin",
      "password": "admin",
      "url": "https://keycloak.example.com",
      "base_path": "/auth",
      "realm": "master"
    }
```

### Flat Key Format

Individual keys for each setting:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  client_id: "admin-cli"
  username: "admin"
  password: "admin"
  url: "https://keycloak.example.com"
  base_path: "/auth"
  realm: "master"
```

### Client Credentials Grant

For service-to-service authentication without a username/password:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: |
    {
      "client_id": "my-service-account",
      "client_secret": "secret-value",
      "url": "https://keycloak.example.com",
      "realm": "master"
    }
```

## Multiple Instances

You can manage multiple Keycloak instances by creating multiple `ProviderConfig` resources:

```yaml
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-staging
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-staging-credentials
      key: credentials
      namespace: crossplane-system
---
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-production
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-production-credentials
      key: credentials
      namespace: crossplane-system
```

Then reference the appropriate config in each resource:

```yaml
spec:
  providerConfigRef:
    name: keycloak-production
```

## Connection Secret Key Renaming

Resources that publish connection details (currently `openidclient.Client`)
write them under fixed keys: `clientID`, `clientSecret` and, for service
accounts, `serviceAccountUserId`. Consumers frequently expect different key
names — Envoy Gateway's OIDC `SecurityPolicy`, for example, requires
`client-id` and `client-secret`.

The provider therefore adds the keys you configure to the resource's own
connection secret — the one named by `spec.writeConnectionSecretToRef` — so
consumers keep pointing at that single secret.

Renaming in place is *additive*: the managed resource's own controller
republishes every key it owns on each reconcile, so a key it published cannot
be removed for good. `clientID: client-id` therefore adds `client-id` next to
`clientID`, both carrying the same value. That is all a consumer such as an
Envoy Gateway OIDC `SecurityPolicy` needs, since it looks up the keys it wants
and ignores the rest. If the original names must not appear at all, switch to
the `SeparateSecret` [mode](#modes).

Configure the renaming centrally, for every resource using a `ProviderConfig`:

```yaml
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-provider-config
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-credentials
      key: credentials
      namespace: crossplane-system
  connectionSecretKeys:
    rename:
      clientID: client-id
      clientSecret: client-secret
```

…or per resource, with annotations. Annotation entries are merged on top of
the `ProviderConfig` map, so a single resource can extend or override the
central configuration without replacing it:

| Annotation | Description |
|------------|-------------|
| `keycloak.crossplane.io/connection-secret-key-rename` | `<oldKey>=<newKey>` pairs, or a YAML mapping of `<oldKey>: <newKey>` (see below) |
| `keycloak.crossplane.io/connection-secret-add-fields` | `<newKey>=<source>` pairs, or a YAML mapping of `<newKey>: <source>` (see [Adding Status/ProviderConfig Fields](#adding-statusproviderconfig-fields)) |
| `keycloak.crossplane.io/connection-secret-transform-mode` | `InPlace` (default) or `SeparateSecret`, see [Modes](#modes) |
| `keycloak.crossplane.io/connection-secret-transform-name` | `SeparateSecret` mode only: name of the republished secret (default: `<connection secret>-transformed`) |

The rename and add-fields annotations both accept two syntaxes, tried in this
order:

1. A YAML block mapping — the most natural way to write a list of pairs in a
   Kubernetes annotation:

   ```yaml
   keycloak.crossplane.io/connection-secret-key-rename: |
     clientID: client-id
     clientSecret: client-secret
   ```

2. A comma- and/or newline-separated list of `<oldKey>=<newKey>` pairs,
   used as a fallback whenever the value does not parse as a YAML mapping
   (e.g. because it uses `=` instead of `: `):

   ```yaml
   keycloak.crossplane.io/connection-secret-key-rename: "clientID=client-id,clientSecret=client-secret"
   ```

   or, one pair per line:

   ```yaml
   keycloak.crossplane.io/connection-secret-key-rename: |
     clientID=client-id
     clientSecret=client-secret
   ```

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
kind: Client
metadata:
  name: my-app
  annotations:
    keycloak.crossplane.io/connection-secret-key-rename: "clientID=client-id,clientSecret=client-secret"
spec:
  forProvider:
    clientId: my-app
    accessType: CONFIDENTIAL
    realmIdRef:
      name: my-realm
  writeConnectionSecretToRef:
    name: my-app-connection
    namespace: default
  providerConfigRef:
    name: keycloak-provider-config
```

`default/my-app-connection` then carries `client-id` and `client-secret` in
addition to the keys the provider publishes itself:

```console
$ kubectl get secret my-app-connection -o jsonpath='{.data}' | jq keys
[
  "attribute.client_secret",
  "clientID",
  "clientSecret",
  "client-id",
  "client-secret"
]
```

The keys the provider added on your behalf are recorded in the
`keycloak.crossplane.io/connection-secret-transform-keys` annotation on that
secret. Only those keys are ever written or removed again, so a key published
by the resource itself — or one of upjet's `attribute.*` keys, which carry the
Terraform state — can never be overwritten.

### Modes

| Mode | Behaviour |
|------|-----------|
| `InPlace` (default) | The configured keys are added to the resource's own connection secret. No second secret is created. Renaming is additive, i.e. the original key stays. |
| `SeparateSecret` | The connection secret is left untouched and a second, transformed secret is published next to it, containing only the renamed/added keys. Use it when the original key names must not appear. |

Pick the mode on the `ProviderConfig`
(`spec.connectionSecretKeys.mode`), per resource with the
`keycloak.crossplane.io/connection-secret-transform-mode` annotation, or on a
`ConnectionSecretTransform` (`spec.mode`) — the same precedence as the
`rename`/`add` maps:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
kind: Client
metadata:
  name: my-app
  annotations:
    keycloak.crossplane.io/connection-secret-transform-mode: SeparateSecret
    keycloak.crossplane.io/connection-secret-key-rename: "clientID=client-id,clientSecret=client-secret"
    keycloak.crossplane.io/connection-secret-transform-name: "my-app-oidc"
spec:
  writeConnectionSecretToRef:
    name: my-app-connection
    namespace: default
```

This publishes `default/my-app-oidc` next to the untouched
`default/my-app-connection`. The republished secret is owned by the connection
secret, which is in turn owned by the managed resource, so both are deleted
together with the `Client`. Switching back to `InPlace` deletes it again.

Namespaced managed resources (`*.keycloak.m.crossplane.io`) are configured
with the annotations (or a `ConnectionSecretTransform`, see below) only, since
the `rename`/`add` maps live on the cluster-scoped `ProviderConfig`.

Editing an annotation (or the `ProviderConfig` map) does not change the
connection secret itself, so the new configuration is not applied
immediately: the provider re-evaluates each connection secret once per poll
interval (`--poll-interval`, one minute by default) and applies it then.

## Configuring via a `ConnectionSecretTransform` Object

Both of the above configure the transform through the managed resource's own
manifest — the `ProviderConfig` it references, or its own annotations. A
third option, the namespaced `ConnectionSecretTransform` custom resource,
names the connection secret to transform directly instead, and is useful when
you would rather not edit the resource that owns the secret at all — for
example when it is reconciled by a separate GitOps pipeline you don't
control, or when several unrelated teams each want to republish the same
secret differently.

```yaml
apiVersion: keycloak.crossplane.io/v1beta1
kind: ConnectionSecretTransform
metadata:
  name: my-app-oidc
  namespace: default
spec:
  sourceSecretRef:
    name: my-app-connection
  # Omit mode (or set it to InPlace) to write into my-app-connection itself.
  mode: SeparateSecret
  transformedSecretName: my-app-oidc
  rename:
    clientID: client-id
    clientSecret: client-secret
  add:
    providerConfigName: "providerConfig:metadata.name"
```

`spec.sourceSecretRef.name` is the name of an existing connection secret in
the `ConnectionSecretTransform`'s own namespace — connection secrets are
always namespaced, whether the managed resource that owns them is
cluster-scoped or namespaced, so `ConnectionSecretTransform` is namespaced
too, unlike the cluster-scoped `ProviderConfig`. `spec.rename`, `spec.add` and
`spec.transformedSecretName` accept the same values as the `ProviderConfig`
map and the annotations respectively (native YAML maps here, since there is
no annotation string encoding to work around), and are merged on top of
both — a `ConnectionSecretTransform` entry always wins over the same key
configured on the `ProviderConfig` or via an annotation. This makes it the
right tool when a `ConnectionSecretTransform` should override, rather than
merely extend, another team's central configuration.

Editing, creating or deleting a `ConnectionSecretTransform` reconciles its
named secret immediately — it does not have to wait for the poll interval,
unlike editing the `ProviderConfig` or an annotation.

Its own `status.conditions` reports whether it is currently applied
(`Ready: True`, with `status.transformedSecretName` set) or refused, together
with the reason (e.g. a name collision, or another `ConnectionSecretTransform`
in the same namespace naming the same source secret, which is ambiguous and
so refuses both rather than picking one arbitrarily):

```console
$ kubectl get connectionsecrettransform my-app-oidc -o jsonpath='{.status.conditions}'
```

## Adding Status/ProviderConfig Fields

Beyond renaming existing keys, the connection secret (or, in `SeparateSecret`
mode, the transformed copy) can also gain entirely new keys, sourced from the owning managed resource's `status.atProvider` or
from the referenced `ProviderConfig` object — for example, to publish the
Keycloak client's internal ID, or which `ProviderConfig` produced the secret,
alongside the credentials. Values are never sourced from the connection
secret itself, so this cannot be used to duplicate or leak an existing
secret value under a different key.

Configure it centrally on the `ProviderConfig`:

```yaml
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-provider-config
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-credentials
      key: credentials
      namespace: crossplane-system
  connectionSecretKeys:
    rename:
      clientID: client-id
      clientSecret: client-secret
    add:
      issuerUrl: "providerConfig:metadata.name"
```

…or per resource, with the `keycloak.crossplane.io/connection-secret-add-fields`
annotation, using either of the two syntaxes described above:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
kind: Client
metadata:
  name: my-app
  annotations:
    keycloak.crossplane.io/connection-secret-add-fields: |
      internalClientId: status:atProvider.id
      providerConfigName: providerConfig:metadata.name
spec:
  forProvider:
    clientId: my-app
    accessType: CONFIDENTIAL
    realmIdRef:
      name: my-realm
  writeConnectionSecretToRef:
    name: my-app-connection
    namespace: default
  providerConfigRef:
    name: keycloak-provider-config
```

Each value is a source expression of the form `<prefix>:<dot.path>`:

| Prefix | Resolved against | Example |
|--------|-------------------|---------|
| `status:` | The owning managed resource's `status` | `status:atProvider.clientId` |
| `providerConfig:` | The referenced `ProviderConfig` object | `providerConfig:metadata.name` |

Only scalar values (string, number, boolean) are supported — a path that
resolves to a map or a list, or that does not resolve at all, is reported and
skipped rather than failing the reconcile. `providerConfig:` sources are only
available for cluster-scoped resources, the same restriction as
`ProviderConfig`-wide renaming.

### Safety rules

The transform never destroys data. A misconfiguration is skipped and reported
as a Kubernetes event on the connection secret (`kubectl describe secret
<connection secret>`), so it neither corrupts a secret nor blocks
reconciliation:

| Situation | Behaviour | Event reason |
|-----------|-----------|--------------|
| Rename target is not a valid secret key (`[-._a-zA-Z0-9]+`) | Entry ignored, the remaining renames are applied | `InvalidConnectionSecretKeyRename` |
| Two keys would end up under the same name | The colliding rename is skipped, the key keeps its original name | `ConnectionSecretKeyRenameConflict` |
| Add-field target is not a valid secret key, or its source expression does not resolve to a scalar | Entry ignored, the remaining added fields are applied | `InvalidConnectionSecretFieldAdd` |
| An added field's key collides with an existing (or renamed) key | The added field is skipped, the existing key wins | `ConnectionSecretFieldAddConflict` |
| A key would overwrite one the provider publishes itself, or an `attribute.*` key (`InPlace` mode) | The key is skipped, the existing value is kept | `ConnectionSecretKeyConflict` |
| `connection-secret-transform-mode` is neither `InPlace` nor `SeparateSecret` | The default (`InPlace`) is used | `InvalidConnectionSecretTransformMode` |
| `connection-secret-transform-name` is not a valid secret name, or names the connection secret itself (`SeparateSecret` mode) | Nothing is written | `InvalidTransformedSecretName` |
| A secret of that name exists and was not written by this controller | Nothing is written, the existing secret is left untouched | `TransformedSecretNotOwned` |
| More than one `ConnectionSecretTransform` in a namespace names the same source secret | None of them are applied; each reports the conflict on its own `status.conditions` | `ConnectionSecretTransformAmbiguous` |

In `InPlace` mode the controller only ever writes or removes the keys listed
in the connection secret's
`keycloak.crossplane.io/connection-secret-transform-keys` annotation, i.e. the
ones it added itself. In `SeparateSecret` mode it only ever writes secrets it
created itself — they carry the
labels `keycloak.crossplane.io/connection-secret-transform: "true"` and
`keycloak.crossplane.io/connection-secret-source: <connection secret>` and are
controller-owned by the connection secret. It only reacts to Crossplane
connection secrets (type `connection.crossplane.io/v1alpha1`) and to its own
output; other secrets, including the provider's credentials, are ignored.

