---
sidebar_position: 1
title: Realms
description: Manage Keycloak realms declaratively
---

# Realms

A Realm in Keycloak is a space where you manage objects like users, applications, roles, and groups. Each realm is isolated from other realms.

## API Reference

- **API Group**: `realm.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kind**: `Realm`

## Basic Realm

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: my-realm
spec:
  forProvider:
    realm: "my-realm"
    enabled: true
  providerConfigRef:
    name: keycloak-provider-config
```

## Realm with Display Settings

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: production-realm
spec:
  forProvider:
    realm: "production"
    enabled: true
    displayName: "Production Realm"
    attributes:
      environment: "production"
  providerConfigRef:
    name: keycloak-provider-config
```

## Realm with Password Policy

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: secure-realm
spec:
  forProvider:
    realm: "secure-realm"
    enabled: true
    passwordPolicy: "length(8) and digits(2) and upperCase(1)"
    otpPolicy:
      - algorithm: "HOTP"
        digits: 6
        type: "totp"
  providerConfigRef:
    name: keycloak-provider-config
```

## Realm with SMTP Configuration

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: realm-with-email
spec:
  forProvider:
    realm: "email-realm"
    enabled: true
    smtpServer:
      - host: "smtp.example.com"
        port: "587"
        from: "noreply@example.com"
  providerConfigRef:
    name: keycloak-provider-config
```

## Related Resources

- **[Realm Settings](./realm-settings.md)** — Manage `RealmEvents`, `RequiredAction`, `UserProfile`, `KeystoreRsa`, `DefaultClientScopes`, `OptionalClientScopes`, `ClientPolicyProfile`, and `ClientPolicyProfilePolicy`
- **[Default Configuration](./default-config.md)** — Configure default groups and roles for new users

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `realm` | string | The realm name (unique identifier) |
| `enabled` | bool | Whether the realm is active |
| `displayName` | string | Human-readable display name |
| `passwordPolicy` | string | Password policy expression |
| `attributes` | map | Custom key-value attributes |
| `smtpServer` | object | Email server configuration |
| `otpPolicy` | object | One-time password configuration |
