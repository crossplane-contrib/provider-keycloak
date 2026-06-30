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
