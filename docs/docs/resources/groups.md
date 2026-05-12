---
sidebar_position: 5
title: Groups
description: Manage Keycloak groups for organizing users
---

# Groups

Groups provide a way to organize users and assign roles to multiple users at once.

## API Reference

- **API Group**: `group.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kind**: `Group`

## Basic Group

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: developers
spec:
  forProvider:
    name: "Developers"
    realmId: "my-realm"
  deletionPolicy: Delete
```

## Group with Attributes

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: platform-team
spec:
  forProvider:
    name: "Platform Team"
    realmId: "my-realm"
    attributes:
      department: "engineering"
      cost-center: "CC-1234"
  deletionPolicy: Delete
```

## Child Group (Nested Groups)

Create hierarchical group structures:

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: frontend-devs
spec:
  forProvider:
    name: "Frontend Developers"
    realmId: "my-realm"
    parentId: "parent-group-id"
  deletionPolicy: Delete
```

## Group Memberships

Assign users to groups:

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Memberships
metadata:
  name: dev-team-members
spec:
  forProvider:
    realmId: "my-realm"
    groupIdRef:
      name: developers
    members:
      - "user-id-1"
      - "user-id-2"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Group display name |
| `realmId` | string | Realm this group belongs to |
| `parentId` | string | Parent group ID (for nested groups) |
| `attributes` | map | Custom key-value attributes |
