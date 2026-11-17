---
sidebar_position: 2
title: Credentials
description: All supported credential fields and authentication methods
---

# Credentials Reference

This page documents all supported credential fields for connecting to a Keycloak instance.

## Supported Fields

The credential fields map directly to the [Keycloak Terraform Provider configuration](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs#argument-reference).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | **Yes** | Keycloak server URL |
| `client_id` | string | **Yes** | OAuth2 client ID for authentication |
| `username` | string | Conditional | Admin username |
| `password` | string | Conditional | Admin password |
| `client_secret` | string | Conditional | Client secret (for client credentials grant) |
| `realm` | string | No | Authentication realm (defaults to `master`) |
| `base_path` | string | No | URL path prefix (e.g., `/auth`) |
| `admin_url` | string | No | Separate admin API URL if different from `url` |
| `root_ca_certificate` | string | No | PEM-encoded CA certificate for TLS |

## Authentication Methods

### Password Grant (Admin CLI)

The most common method using username and password:

```json
{
  "client_id": "admin-cli",
  "username": "admin",
  "password": "admin",
  "url": "https://keycloak.example.com",
  "realm": "master"
}
```

### Client Credentials Grant

For automated systems using a service account:

```json
{
  "client_id": "my-service-account",
  "client_secret": "client-secret-value",
  "url": "https://keycloak.example.com",
  "realm": "master"
}
```

### Keycloak 26.8.2 and later: `kc-admin-api` client scope

Since Keycloak 26.8.2 ([CVE-2026-19688](https://www.keycloak.org/docs/latest/upgrading/index.html#admin-rest-api-now-enforces-an-audience-check)), the Admin REST API only accepts access tokens that carry the `kc-admin-api` scope and the realm's Admin API audience. Keycloak adds both through the built-in `kc-admin-api` client scope.

On upgrade and for new realms, Keycloak assigns that scope only to `admin-cli` and `security-admin-console`. Any other client in the provider credentials needs it too: service accounts (client credentials grant), your own clients for the password grant, clients in a non-master realm and clients authenticating with a JWT. Without it, every request of the provider fails with `401 Unauthorized`, and all managed resources become `Synced=False`.

Assign the scope as a default client scope of the client before you upgrade to 26.8.2, or right after the upgrade (the scope only exists from 26.8.2 on):

```bash
KC=https://keycloak.example.com
REALM=master                # realm of the client used in the provider credentials
CLIENT_ID=my-service-account
TOKEN=...                   # admin token, e.g. from admin-cli

SCOPE=$(curl -sf -H "Authorization: Bearer $TOKEN" "$KC/admin/realms/$REALM/client-scopes" | jq -r '.[] | select(.name == "kc-admin-api") | .id')
CLIENT=$(curl -sf -H "Authorization: Bearer $TOKEN" "$KC/admin/realms/$REALM/clients?clientId=$CLIENT_ID" | jq -r '.[0].id')
curl -sf -X PUT -H "Authorization: Bearer $TOKEN" "$KC/admin/realms/$REALM/clients/$CLIENT/default-client-scopes/$SCOPE"
```

If the client itself is managed by this provider (for example, a client for another provider instance), add `kc-admin-api` to its `ClientDefaultScopes`. That resource sets the complete list, so keep the scopes the client already has:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientDefaultScopes
metadata:
  name: my-service-account-default-scopes
spec:
  forProvider:
    realmId: master
    clientIdRef:
      name: my-service-account
    defaultScopes:
      - kc-admin-api
      - acr
      - basic
      - email
      - profile
      - roles
      - web-origins
  providerConfigRef:
    name: keycloak-provider-config
```

Keycloak can turn the check off with `--spi-admin--audience-check-enabled=false` (environment variable `KC_SPI_ADMIN__AUDIENCE_CHECK_ENABLED=false`). Use it only as a bridge while you assign the scope, as it removes the protection added by the fix.

## URL Validation and Normalization

The provider validates URLs before use:

| Rule | Example |
|------|---------|
| Must be absolute with scheme | ✓ `https://keycloak.example.com` |
| No query parameters | ✗ `https://keycloak.example.com?foo=bar` |
| No fragments | ✗ `https://keycloak.example.com#section` |
| Trailing slash removed | `https://kc.example.com/` → `https://kc.example.com` |
| `base_path` must start with `/` | ✓ `/auth` |
| `base_path: "/"` normalized to empty | `/` → `` |
| Trailing slash on base_path removed | `/auth/` → `/auth` |

## Custom TLS Certificate

For self-signed or internal CA certificates:

```json
{
  "client_id": "admin-cli",
  "username": "admin",
  "password": "admin",
  "url": "https://keycloak.internal.example.com",
  "root_ca_certificate": "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----"
}
```

## Base Path

Older versions of Keycloak (before v17) served the application under `/auth`. Modern versions (Quarkus-based) typically serve at the root path.

| Keycloak Version | Base Path |
|-----------------|-----------|
| < 17 (WildFly) | `/auth` |
| ≥ 17 (Quarkus) | `` (empty) |
