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
| `keycloak.crossplane.io/connection-secret-key-rename` | Comma-separated `<oldKey>=<newKey>` pairs |
| `keycloak.crossplane.io/connection-secret-transform-name` | Name of the republished secret (default: `<connection secret>-transformed`) |

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
with the annotations only, since the `rename` map lives on the cluster-scoped
`ProviderConfig`.
