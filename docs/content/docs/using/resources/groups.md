---
sidebar_position: 5
title: Groups
description: Manage Keycloak groups, memberships, role mappings, and permissions
---

Use groups when multiple users should share the same roles or when you need a hierarchical structure such as teams, departments, or environments. Groups let you model organization structure once and then manage access in bulk.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| Group | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/Group/v1alpha1) |
| Memberships | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group_memberships`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group_memberships) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/Memberships/v1alpha1) |
| Roles | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group_roles`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group_roles) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/Roles/v1alpha1) |
| Permissions | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group_permissions) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/Permissions/v1alpha1) |
| AdminPermissions | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group_admin_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group_admin_permissions) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/AdminPermissions/v1alpha1) |

## Examples

### Basic group

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: test
spec:
  deletionPolicy: Delete
  forProvider:
    name: test
    realmId: dev
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Child groups with the same name under different parents

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: test-parent-1
  labels:
    role: parent
    parent: test1
spec:
  deletionPolicy: Delete
  forProvider:
    name: test-parent-1
    realmId: dev
  providerConfigRef:
    name: "keycloak-provider-config"
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: test-child-1
spec:
  deletionPolicy: Delete
  forProvider:
    name: test-child
    realmId: dev
    parentIdSelector:
      matchLabels:
        role: parent
        parent: test1
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Group memberships

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Memberships
metadata:
  name: test-members
spec:
  deletionPolicy: Delete
  forProvider:
    groupIdRef:
      name: test
      policy:
        resolve: Always
    members:
      - bree
      - tim-tester
    realmId: dev
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Group roles

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: group-roles
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    groupIdRef:
      name: test
      policy:
        resolve: Always
    roleIdsRefs:
      - name: "test-client"
        policy:
          resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Group permissions

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Permissions
metadata:
  name: my-group-permission
spec:
  managementPolicies: ["Create", "Update", "Observe"]
  forProvider:
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    groupIdRef:
      name: "test"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Fine-grained admin permissions (v2)

`AdminPermissions` manages a single fine-grained admin permission for the groups
of a realm. It requires Keycloak 26.2 or newer started with the
`admin-fine-grained-authz:v2` feature and a realm with
`adminPermissionsEnabled: true`. Keycloak then creates an `admin-permissions`
client for the realm that acts as the resource server for all of its admin
permissions.

`admin-fine-grained-authz:v2` replaces `admin-fine-grained-authz:v1`, so
`AdminPermissions` and the v1 `Permissions` resource cannot be used against the
same Keycloak instance.

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: AdminPermissions
metadata:
  name: admins-manage-groups
spec:
  deletionPolicy: Delete
  forProvider:
    name: admins-can-manage-groups
    description: Admins can view and manage the members of the group
    decisionStrategy: UNANIMOUS
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    groupIdsRefs:
      - name: "test"
    scopes:
      - view
      - manage-members
  providerConfigRef:
    name: "keycloak-provider-config"
```

Valid `scopes` for group permissions are `view`, `manage`, `view-members`,
`manage-members` and `manage-membership`. Without `groupIds` the permission
applies to every group of the realm, otherwise only to the referenced groups. A
permission without policies is evaluated as "deny", so attach policies once they
exist on the realm's `admin-permissions` client, either by ID via `policies` or
through the typed reference fields (`groupPolicies`, `rolePolicies`,
`userPolicies`, ...).

## Key Fields

| Resource | Field | Description |
|----------|-------|-------------|
| `Group` | `name` | Group name shown in Keycloak. |
| `Group` | `realmId` | Realm where the group is created. |
| `Group` | `parentIdRef` / `parentIdSelector` | Places the group under a parent group for nested hierarchies. |
| `Memberships` | `groupIdRef` | Targets the group whose members you want to manage. |
| `Memberships` | `members` | List of usernames to keep in the group. |
| `Roles` | `groupIdRef` | Targets the group that should receive roles. |
| `Roles` | `roleIdsRefs` | References the roles assigned to the group. |
| `Permissions` | `realmIdRef` | Enables fine-grained admin permissions for groups in a realm. |
| `Permissions` | `groupIdRef` | Targets the group for which permissions are managed. |
| `AdminPermissions` | `realmIdRef` | Realm whose groups the permission applies to. |
| `AdminPermissions` | `groupIdsRefs` | Restricts the permission to specific groups. Leave empty to target all groups of the realm. |
| `AdminPermissions` | `scopes` | Admin operations the permission covers: `view`, `manage`, `view-members`, `manage-members`, `manage-membership`. |
| `AdminPermissions` | `decisionStrategy` | How the attached policies are combined: `UNANIMOUS`, `AFFIRMATIVE` or `CONSENSUS`. |
| `AdminPermissions` | `policies` | IDs of authorization policies granting the permission. Typed `*Policies` reference fields are also available. |

## Related Resources

- [Users](./users.md)
- [Roles](./roles.md)
- [Default Configuration](./default-config.md)
