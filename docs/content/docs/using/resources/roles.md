---
sidebar_position: 4
title: Roles
description: Manage Keycloak roles and role assignments
---

# Roles

Roles define permissions and can be assigned to users or groups. Keycloak supports realm roles and client roles.

## API Reference

> **Schema source:** This page highlights common fields and examples. For the complete OpenAPI schema, including references, selectors, status fields, and connection details, see the generated CRDs in `package/crds/`.

- **API Group**: `role.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kind**: `Role`

## Realm Role

A role scoped to the entire realm:

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: admin-role
spec:
  forProvider:
    realmId: "my-realm"
    name: "admin"
    description: "Administrator role with full access"
  providerConfigRef:
    name: keycloak-provider-config
```

## Client Role

A role scoped to a specific client:

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: api-reader
spec:
  forProvider:
    realmId: "my-realm"
    name: "reader"
    clientId: "my-api-client-id"
    description: "Read-only access to the API"
  providerConfigRef:
    name: keycloak-provider-config
```

## Composite Role

A role that inherits permissions from other roles:

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: super-admin
spec:
  forProvider:
    realmId: "my-realm"
    name: "super-admin"
    compositeRoles:
      - "admin"
      - "user-manager"
  providerConfigRef:
    name: keycloak-provider-config
```

## Role with Attributes

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: team-lead
spec:
  forProvider:
    realmId: "my-realm"
    name: "team-lead"
    description: "Team lead with elevated permissions"
    attributes:
      department: "engineering"
      level: "senior"
  providerConfigRef:
    name: keycloak-provider-config
```

## Assign Roles to Groups

Use the `group.keycloak.crossplane.io/Roles` resource to map roles to groups:

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: admin-group-roles
spec:
  forProvider:
    groupIdRef:
      name: admin-group
    roleIdsRefs:
      - name: k8s-admin
      - name: argocd-admin
    realmId: my-realm
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm this role belongs to |
| `name` | string | Role name |
| `clientId` | string | Client ID (for client roles) |
| `description` | string | Human-readable description |
| `compositeRoles` | []string | Roles to inherit from |
| `attributes` | map | Custom key-value attributes |
