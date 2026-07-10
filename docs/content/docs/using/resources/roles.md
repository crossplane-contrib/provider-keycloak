---
sidebar_position: 4
title: Roles
description: Manage realm roles and client roles in Keycloak
---

# Roles

Use roles to define permissions in Keycloak. Create realm roles for permissions shared across a realm, and client roles when access should be scoped to a specific application or service.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| Role | `role.keycloak.crossplane.io/v1alpha1` | [`keycloak_role`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/role) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/role.keycloak.crossplane.io/Role/v1alpha1) |

## Examples

### Realm role

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: test
spec:
  deletionPolicy: Delete
  forProvider:
    realmId: "dev"
    name: "test"
    description: "abc"
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Client role

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: test-client
spec:
  deletionPolicy: Delete
  forProvider:
    realmId: "dev"
    name: "test-client"
    clientIdRef:
      name: "test"
      policy:
        resolve: Always
    description: "abc"
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Managing a built-in realm role without deleting it

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: offline-access
spec:
  managementPolicies: [Observe, Update]
  deletionPolicy: Orphan
  forProvider:
    realmId: "dev"
    name: "offline_access"
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Managing a built-in client role without deleting it

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: account-view-profile
spec:
  managementPolicies: [Observe, Update]
  deletionPolicy: Orphan
  forProvider:
    realmId: "dev"
    clientId: "account"
    name: "view-profile"
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Key Fields

| Field | Description |
|-------|-------------|
| `name` | Role name stored in Keycloak. |
| `realmId` | Realm where the role is created. |
| `clientIdRef` | Set this for a client role so the role is scoped to a specific client. Omit it for a realm role. |
| `description` | Human-readable role description. |
| `compositeRoles` | Optional list of roles that should be included in this role as composites. |

## Related Resources

- [Groups](./groups.md)
- [Users](./users.md)
- [Default Configuration](./default-config.md)
- [Service Accounts](./service-accounts.md)
