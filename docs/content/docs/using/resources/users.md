---
sidebar_position: 3
title: Users
description: Manage Keycloak users declaratively
---

# Users

Users represent individuals who can authenticate with Keycloak.

## API Reference

> **Schema source:** This page highlights common fields and examples. For the complete OpenAPI schema, including references, selectors, status fields, and connection details, see the generated CRDs in `package/crds/`.

- **API Group**: `user.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kind**: `User`

## Basic User

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: basic-user
spec:
  forProvider:
    realmId: "my-realm"
    username: "jdoe"
    enabled: true
  providerConfigRef:
    name: keycloak-provider-config
```

## User with Full Profile

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: john-doe
spec:
  forProvider:
    realmId: "my-realm"
    username: "johndoe"
    email: "john.doe@example.com"
    firstName: "John"
    lastName: "Doe"
    enabled: true
  providerConfigRef:
    name: keycloak-provider-config
```

## User with Initial Password

Set a temporary password that must be changed on first login:

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: user-with-password
spec:
  forProvider:
    realmId: "my-realm"
    username: "newuser"
    enabled: true
    initialPassword:
      - valueSecretRef:
          key: "password"
          name: "user-password-secret"
          namespace: "crossplane-system"
        temporary: true
  providerConfigRef:
    name: keycloak-provider-config
```

## User with Federated Identity

Link a user to an external identity provider (e.g., GitHub):

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: federated-user
spec:
  forProvider:
    realmId: "my-realm"
    username: "ghuser"
    federatedIdentity:
      - identityProvider: "github"
        userId: "123456"
        userName: "ghuser"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm this user belongs to |
| `username` | string | Unique username |
| `email` | string | Email address |
| `firstName` | string | First name |
| `lastName` | string | Last name |
| `enabled` | bool | Whether the user can authenticate |
| `initialPassword` | object | Temporary password configuration |
| `federatedIdentity` | []object | External identity provider links |
