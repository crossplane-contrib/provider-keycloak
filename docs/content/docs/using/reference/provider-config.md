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

Keycloak's `/admin/serverinfo` endpoint (used once, on initial login, to
detect the server version) is a **global**, master-realm-only endpoint — it
cannot be scoped to a single realm. This has two practical consequences when
you try to restrict the provider's service account so it can only administer
one non-master realm (say, `demo`), as described in
[crossplane-contrib/provider-keycloak#742](https://github.com/crossplane-contrib/provider-keycloak/issues/742).

### Service account confined entirely to the `demo` realm

If the client used by the `ProviderConfig` lives in, and only has roles
scoped to, the `demo` realm, login fails immediately with:

```text
failed to perform initial login to Keycloak:
error sending GET request to /admin/serverinfo: 403 Forbidden
```

This happens because `/admin/serverinfo` requires the `view-system` role (or,
since Keycloak 26.5.4, `manage-realms`) from the **master realm's**
`master-realm` client — a role that cannot be granted to a service account
that only holds `demo`-realm-scoped roles. There is currently no way to work
around this: the provider must authenticate with a client that has at least
one of those two roles in `master`.

### Service account in `master`, but only with `demo`-realm-admin roles

A client created in the `master` realm whose service account is only granted
roles from the `demo` realm's admin client (e.g. `manage-realm`,
`manage-clients`, `manage-users`, ...) can reach `/admin/serverinfo`
successfully, but the response has an empty version field, because the
account still lacks `view-system`/`manage-realms`. Older provider builds
failed with:

```text
failed to perform initial login to Keycloak: malformed version: []
```

**This case is fixed** by setting the `keycloak_version` credential key (see
[Supported Credential Keys](#supported-credential-keys)) to the version of
your Keycloak server. When the server reports an empty version, the provider
falls back to this configured value instead of failing:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  client_id: "crossplane-provider"
  client_secret: "<client-secret>"
  url: "https://keycloak.example.com"
  realm: "master"
  keycloak_version: "26.6.2"
```

With `keycloak_version` set, this client can successfully log in and manage
only the resources its service account roles permit (e.g. everything under
the `demo` realm), without holding broader administrative access to other
realms.

### Known limitation

A service account with **no** roles in `master` at all (the fully
realm-scoped setup above) cannot be supported without changes to the
underlying [`terraform-provider-keycloak`](https://github.com/keycloak/terraform-provider-keycloak)
client, which currently treats any non-2xx response from
`/admin/serverinfo` — including on the initial login handshake — as fatal.
Until that is addressed upstream, a single-realm-only `ProviderConfig` must
use a client credential located in `master`, granted only the target
realm's admin-client roles (Configuration B above) plus `keycloak_version`
set explicitly.

Both configurations are covered by e2e tests: Configuration B by
`dev/demos/basic/087-nonmaster-provider.yaml` (and its namespaced
equivalent), whose setup creates the client in `master` but grants only
`provider-e2e-realm` admin-client roles, and Configuration A's 403 failure
mode by the standalone `cluster/test/restrictedrealmprovider/` chainsaw
suite. See [End-to-End Tests](../../developing/e2e-tests.md#standalone-chainsaw-suites)
for details.
