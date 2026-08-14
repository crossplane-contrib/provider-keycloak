---
sidebar_position: 4
title: Roles
description: Manage realm roles and client roles in Keycloak
---

Use roles to define permissions in Keycloak. Create realm roles for permissions shared across a realm, and client roles when access should be scoped to a specific application or service.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| Role | `role.keycloak.crossplane.io/v1alpha1` | [`keycloak_role`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/role) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/role.keycloak.crossplane.io/Role/v1alpha1) |
| AdminPermissions | `role.keycloak.crossplane.io/v1alpha1` | [`keycloak_role_admin_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/role_admin_permissions) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/role.keycloak.crossplane.io/AdminPermissions/v1alpha1) |

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

### Fine-grained admin permissions (v2)

`AdminPermissions` manages a single fine-grained admin permission for the roles
of a realm. It requires Keycloak 26.2 or newer started with the
`admin-fine-grained-authz:v2` feature and a realm with
`adminPermissionsEnabled: true`. Keycloak then creates an `admin-permissions`
client for the realm that acts as the resource server for all of its admin
permissions.

`admin-fine-grained-authz:v2` replaces `admin-fine-grained-authz:v1`, so
`AdminPermissions` and the v1 permission resources cannot be used against the
same Keycloak instance.

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: AdminPermissions
metadata:
  name: admins-map-roles
spec:
  deletionPolicy: Delete
  forProvider:
    name: admins-can-map-roles
    description: Admins can assign any role of the realm
    decisionStrategy: UNANIMOUS
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    scopes:
      - map-role
      - map-role-client-scope
      - map-role-composite
  providerConfigRef:
    name: "keycloak-provider-config"
```

Valid `scopes` for role permissions are `map-role`, `map-role-client-scope` and
`map-role-composite`. Without `roleIds` the permission applies to every role of
the realm, otherwise only to the referenced roles. A permission without policies
is evaluated as "deny", so attach policies once they exist on the realm's
`admin-permissions` client, either by ID via `policies` or through the typed
reference fields (`groupPolicies`, `rolePolicies`, `userPolicies`, ...).

## Key Fields

| Field | Description |
|-------|-------------|
| `name` | Role name stored in Keycloak. |
| `realmId` | Realm where the role is created. |
| `clientIdRef` | Set this for a client role so the role is scoped to a specific client. Omit it for a realm role. |
| `description` | Human-readable role description. |
| `compositeRoles` | Optional list of roles that should be included in this role as composites. |
| `AdminPermissions`: `realmIdRef` | Realm whose roles the permission applies to. |
| `AdminPermissions`: `roleIdsRefs` | Restricts the permission to specific roles. Leave empty to target all roles of the realm. |
| `AdminPermissions`: `scopes` | Admin operations the permission covers: `map-role`, `map-role-client-scope`, `map-role-composite`. |
| `AdminPermissions`: `decisionStrategy` | How the attached policies are combined: `UNANIMOUS`, `AFFIRMATIVE` or `CONSENSUS`. |
| `AdminPermissions`: `policies` | IDs of authorization policies granting the permission. Typed `*Policies` reference fields are also available. |

## Related Resources

- [Groups](./groups.md)
- [Users](./users.md)
- [Default Configuration](./default-config.md)
- [Service Accounts](./service-accounts.md)
