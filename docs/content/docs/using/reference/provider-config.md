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

The provider therefore republishes the connection secret under a second name,
with the keys renamed. The original secret is never modified, so the provider
can still rebuild its Terraform state from it.

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
| `keycloak.crossplane.io/connection-secret-transform-name` | Name of the republished secret (default: `<connection secret>-transformed`) |

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
    keycloak.crossplane.io/connection-secret-transform-name: "my-app-oidc"
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

This publishes `default/my-app-oidc` next to the untouched
`default/my-app-connection`. The republished secret is owned by the connection
secret, which is in turn owned by the managed resource, so both are deleted
together with the `Client`.

Namespaced managed resources (`*.keycloak.m.crossplane.io`) are configured
with the annotations only, since the `rename`/`add` maps live on the
cluster-scoped `ProviderConfig`.

Editing an annotation (or the `ProviderConfig` map) does not change the
connection secret itself, so the transformed secret is not updated
immediately: the provider re-evaluates each connection secret once per poll
interval (`--poll-interval`, one minute by default) and applies the new
configuration then.

## Adding Status/ProviderConfig Fields

Beyond renaming existing keys, the transformed secret can also gain entirely
new keys, sourced from the owning managed resource's `status.atProvider` or
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
| `connection-secret-transform-name` is not a valid secret name, or names the connection secret itself | Nothing is written | `InvalidTransformedSecretName` |
| A secret of that name exists and was not written by this controller | Nothing is written, the existing secret is left untouched | `TransformedSecretNotOwned` |

The controller only ever writes secrets it created itself — they carry the
labels `keycloak.crossplane.io/connection-secret-transform: "true"` and
`keycloak.crossplane.io/connection-secret-source: <connection secret>` and are
controller-owned by the connection secret. It only reacts to Crossplane
connection secrets (type `connection.crossplane.io/v1alpha1`) and to its own
output; other secrets, including the provider's credentials, are ignored.

