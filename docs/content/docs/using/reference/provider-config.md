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

## Supported Credential Keys

Whichever credential source format you use (see below), the following keys
are recognized. `url` and `client_id` are required; everything else is
optional and only needs to be set when it applies to your setup.

| Key | Required | Description |
|-----|----------|-------------|
| `url` | Yes | Base URL of the Keycloak instance, before `/auth`. |
| `client_id` | Yes | Client ID used to authenticate. |
| `client_secret` | No | Client secret, for the client credentials grant. |
| `username` / `password` | No | Credentials for the resource owner password grant. |
| `access_token` | No | A pre-obtained access token, used instead of any grant. |
| `realm` | No | Realm the client/user used for authentication belongs to. Defaults to `master`. |
| `base_path` | No | Path prefix (e.g. `/auth`) if Keycloak isn't served at the URL root. |
| `admin_url` | No | Admin URL, if different from `url`. |
| `initial_login` | No | Whether to log in during provider initialization. Defaults to `true`. |
| `client_timeout` | No | Client HTTP timeout in seconds. Defaults to `15`. |
| `tls_insecure_skip_verify` | No | Skip TLS certificate verification. Defaults to `false`. |
| `root_ca_certificate` | No | Additional CA certificate to trust. |
| `tls_client_certificate` / `tls_client_private_key` | No | Client certificate/key (PEM) for mutual TLS. |
| `additional_headers` | No | Map of extra HTTP headers to send with every request. |
| `red_hat_sso` | No | Treat the server as a Red Hat SSO build when parsing its version. Defaults to `false`. |
| `jwt_signing_alg` / `jwt_signing_key` / `jwt_token` / `jwt_token_file` | No | JWT-based client authentication (`client-jwt`). |
| `keycloak_version` | No | Keycloak version to assume when the server does not report it (empty `/admin/serverinfo` version). See [Scoping the Provider to a Single Realm](#scoping-the-provider-to-a-single-realm). |

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

## Scoping the Provider to a Single Realm

To restrict the provider to one non-master realm, create its service-account
client in `master`, then grant that service account only the roles it needs
from the target realm's admin client. Do not create the client in the target
realm: Keycloak's `/admin/serverinfo` endpoint is global and requires a client
in `master` during the initial login.

The following example configures a client in `master` that can administer the
`demo` realm. Adjust the role list to the resources that you intend to manage.
The `demo-realm` client is the built-in admin client for the `demo` realm; use
the corresponding admin client ID if your realm uses a different name.

### 1. Create the service account and grant realm roles

```terraform
resource "keycloak_openid_client" "crossplane_provider" {
  realm_id  = keycloak_realm.master.id
  name      = "Crossplane Provider"
  client_id = "crossplane-provider"

  enabled                   = true
  access_type               = "CONFIDENTIAL"
  standard_flow_enabled    = false
  service_accounts_enabled = true
}

data "keycloak_openid_client" "demo_realm_admin_client" {
  realm_id  = keycloak_realm.master.id
  client_id = "demo-realm"
}

data "keycloak_role" "demo_realm_admin_roles" {
  for_each = toset([
    "create-client",
    "manage-authorization",
    "manage-clients",
    "manage-events",
    "manage-organizations",
    "manage-realm",
    "manage-users",
    "query-clients",
    "query-groups",
    "query-organizations",
    "query-realms",
    "query-users",
    "view-authorization",
    "view-clients",
    "view-events",
    "view-identity-providers",
    "view-organizations",
    "view-realm",
    "view-users",
  ])

  realm_id  = keycloak_realm.master.id
  client_id = data.keycloak_openid_client.demo_realm_admin_client.id
  name      = each.value
}

resource "keycloak_user_roles" "crossplane_provider_service_account" {
  realm_id = keycloak_realm.master.id
  user_id  = keycloak_openid_client.crossplane_provider.service_account_user_id

  role_ids = [
    for role in data.keycloak_role.demo_realm_admin_roles : role.id
  ]
}
```

Apply this configuration and obtain the generated client secret. The client
must be in the `master` realm, and the `realm_id` values for the admin roles
must refer to `master`; the roles themselves belong to the `demo-realm` admin
client and authorize operations in `demo`.

### 2. Create the credentials Secret

Set `keycloak_version` to the actual Keycloak server version. This is required
when the service account can administer `demo` but cannot read the version from
`/admin/serverinfo`.

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
      "client_id": "crossplane-provider",
      "client_secret": "<client-secret-from-keycloak>",
      "url": "https://keycloak.example.com",
      "realm": "master",
      "keycloak_version": "26.6.2"
    }
```

### 3. Create the ProviderConfig

```yaml
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-demo-provider
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-credentials
      key: credentials
      namespace: crossplane-system
```

Apply the Secret and ProviderConfig, then reference the ProviderConfig from
resources in the `demo` realm:

```yaml
spec:
  forProvider:
    realmId: demo
  providerConfigRef:
    name: keycloak-demo-provider
```

The provider authenticates in `master`, but the service account has only the
roles granted through the `demo-realm` admin client. It can therefore manage
the permitted resources in `demo` without broad administrator access to other
realms.

### Unsupported configuration

A client created in `demo` with no roles in `master` cannot be used for this
purpose. Its initial request to `/admin/serverinfo` fails with `403 Forbidden`
before `keycloak_version` can be used. This is a limitation of the underlying
[`terraform-provider-keycloak`](https://github.com/keycloak/terraform-provider-keycloak)
client, not a ProviderConfig setting. See [End-to-End Tests](../../developing/e2e-tests.md#standalone-chainsaw-suites)
for the regression test covering this limitation.
