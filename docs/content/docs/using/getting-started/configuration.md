---
sidebar_position: 2
title: Configuration
description: Configure provider-keycloak credentials to connect to Keycloak
---

# Configuration

Every Keycloak instance you manage requires a `ProviderConfig` resource that specifies how to connect to the Keycloak API.

## Create a Credentials Secret

The provider needs credentials to authenticate with Keycloak. Create a Kubernetes Secret:

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

### Credential Fields

| Field | Required | Description |
|-------|----------|-------------|
| `url` | Yes | Keycloak server URL (e.g., `https://keycloak.example.com`) |
| `client_id` | Yes | OAuth2 client ID (typically `admin-cli`) |
| `username` | Conditional | Admin username (required if not using client credentials) |
| `password` | Conditional | Admin password (required if not using client credentials) |
| `client_secret` | Conditional | Client secret (for client credential grants) |
| `realm` | No | Realm for authentication (defaults to `master`) |
| `base_path` | No | URL path prefix (e.g., `/auth` for older Keycloak versions) |
| `root_ca_certificate` | No | Custom CA certificate for TLS verification |

### Alternative: Flat Secret Format

Instead of embedded JSON, you can use plain key-value pairs:

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

## Create a ProviderConfig

Reference the secret in a `ProviderConfig`:

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
```

## URL Validation Rules

The provider validates and normalizes URL fields:

- `url` must be an absolute URL with scheme and host
- Trailing slashes are removed automatically
- `base_path` must be empty or start with `/`
- `base_path: "/"` is normalized to an empty string
- Query parameters and fragments are not allowed in URLs

## Multiple Keycloak Instances

You can manage multiple Keycloak instances by creating separate `ProviderConfig` resources, each with their own credentials secret. Reference the appropriate config in each managed resource:

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: my-realm
spec:
  forProvider:
    realm: "my-realm"
  providerConfigRef:
    name: keycloak-provider-config  # References a specific ProviderConfig
```

## Next Steps

- [Create your first realm](./first-realm.md)
