# Credentials

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


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
  &#34;client_id&#34;: &#34;admin-cli&#34;,
  &#34;username&#34;: &#34;admin&#34;,
  &#34;password&#34;: &#34;admin&#34;,
  &#34;url&#34;: &#34;https://keycloak.example.com&#34;,
  &#34;realm&#34;: &#34;master&#34;
}
```

### Client Credentials Grant

For automated systems using a service account:

```json
{
  &#34;client_id&#34;: &#34;my-service-account&#34;,
  &#34;client_secret&#34;: &#34;client-secret-value&#34;,
  &#34;url&#34;: &#34;https://keycloak.example.com&#34;,
  &#34;realm&#34;: &#34;master&#34;
}
```

## URL Validation and Normalization

The provider validates URLs before use:

| Rule | Example |
|------|---------|
| Must be absolute with scheme | ✓ `https://keycloak.example.com` |
| No query parameters | ✗ `https://keycloak.example.com?foo=bar` |
| No fragments | ✗ `https://keycloak.example.com#section` |
| Trailing slash removed | `https://kc.example.com/` → `https://kc.example.com` |
| `base_path` must start with `/` | ✓ `/auth` |
| `base_path: &#34;/&#34;` normalized to empty | `/` → `` |
| Trailing slash on base_path removed | `/auth/` → `/auth` |

## Custom TLS Certificate

For self-signed or internal CA certificates:

```json
{
  &#34;client_id&#34;: &#34;admin-cli&#34;,
  &#34;username&#34;: &#34;admin&#34;,
  &#34;password&#34;: &#34;admin&#34;,
  &#34;url&#34;: &#34;https://keycloak.internal.example.com&#34;,
  &#34;root_ca_certificate&#34;: &#34;-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----&#34;
}
```

## Base Path

Older versions of Keycloak (before v17) served the application under `/auth`. Modern versions (Quarkus-based) typically serve at the root path.

| Keycloak Version | Base Path |
|-----------------|-----------|
| &lt; 17 (WildFly) | `/auth` |
| ≥ 17 (Quarkus) | `` (empty) |


