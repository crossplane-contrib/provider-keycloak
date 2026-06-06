---
sidebar_position: 16
title: Default Configuration
description: Manage default groups and roles for new users
---

# Default Configuration

Default configuration resources let you define which groups and roles are automatically assigned to new users in a realm.

## API Reference

> **Schema source:** This page highlights common fields and examples. For the complete OpenAPI schema, including references, selectors, status fields, and connection details, see the generated CRDs in `package/crds/`.

- **API Group**: `defaults.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kinds**: `DefaultGroups`, `Roles`

## DefaultGroups

Assign groups that all new users are automatically added to.

```yaml
apiVersion: defaults.keycloak.crossplane.io/v1alpha1
kind: DefaultGroups
metadata:
  name: realm-default-groups
spec:
  forProvider:
    realmId: "my-realm"
    groupIds:
      - "group-uuid-1"
      - "group-uuid-2"
  providerConfigRef:
    name: keycloak-provider-config
```

### Using Group References

```yaml
apiVersion: defaults.keycloak.crossplane.io/v1alpha1
kind: DefaultGroups
metadata:
  name: realm-default-groups
spec:
  forProvider:
    realmId: "my-realm"
    groupIdsRefs:
      - name: new-users-group
      - name: basic-access-group
  providerConfigRef:
    name: keycloak-provider-config
```

## Roles

Define realm-level roles assigned to all new users by default.

```yaml
apiVersion: defaults.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: realm-default-roles
spec:
  forProvider:
    realmId: "my-realm"
    defaultRoles:
      - "basic-user"
      - "view-profile"
  providerConfigRef:
    name: keycloak-provider-config
```

### Using Role References

```yaml
apiVersion: defaults.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: realm-default-roles
spec:
  forProvider:
    realmId: "my-realm"
    defaultRolesRefs:
      - name: basic-user-role
      - name: view-profile-role
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

### DefaultGroups

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm to set default groups for |
| `groupIds` | []string | List of group UUIDs to assign as defaults |
| `groupIdsRefs` | []ref | References to Group resources |
| `groupIdsSelector` | selector | Selector to match Group resources |

### Roles

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm to set default roles for |
| `defaultRoles` | []string | List of role names to assign as defaults |
| `defaultRolesRefs` | []ref | References to Role resources |
| `defaultRolesSelector` | selector | Selector to match Role resources |
