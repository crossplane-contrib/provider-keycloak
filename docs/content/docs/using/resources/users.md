---
sidebar_position: 3
title: Users
description: Manage users, user role mappings, group memberships, permissions, and custom user federation
---

Use `User` to declaratively manage people who can authenticate to Keycloak. Use `Groups`, `Roles`, and `Permissions` to manage access around those users, and `UserFederation` when you need a custom external user store integration.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| User | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_user`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/user) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/user.keycloak.crossplane.io/User/v1alpha1) |
| Groups | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_user_groups`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/user_groups) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/user.keycloak.crossplane.io/Groups/v1alpha1) |
| Roles | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_user_roles`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/user_roles) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/user.keycloak.crossplane.io/Roles/v1alpha1) |
| Permissions | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_users_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/users_permissions) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/user.keycloak.crossplane.io/Permissions/v1alpha1) |
| AdminPermissions | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_users_admin_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/users_admin_permissions) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/user.keycloak.crossplane.io/AdminPermissions/v1alpha1) |
| UserFederation | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_custom_user_federation`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/custom_user_federation) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/user.keycloak.crossplane.io/UserFederation/v1alpha1) |

## Examples

### Basic user

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: bree
spec:
  deletionPolicy: Delete
  forProvider:
    realmId: "dev"
    username: "bree"
  providerConfigRef:
    name: "keycloak-provider-config"
```

### User roles

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: user-roles
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    roleIdsRefs:
      - name: test
        policy:
          resolve: Always
    userIdRef:
      name: "tim-tester"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### User groups

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: Groups
metadata:
  name: user-groups
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    groupIdsRefs:
      - name: test
        policy:
          resolve: Always
    userIdRef:
      name: "tim-tester"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### User permissions

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: Permissions
metadata:
  name: my-user-permission
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Fine-grained admin permissions (v2)

`AdminPermissions` manages a single fine-grained admin permission for the users
of a realm. It requires Keycloak 26.2 or newer started with the
`admin-fine-grained-authz:v2` feature and a realm with
`adminPermissionsEnabled: true`. Keycloak then creates an `admin-permissions`
client for the realm that acts as the resource server for all of its admin
permissions.

`admin-fine-grained-authz:v2` replaces `admin-fine-grained-authz:v1`, so
`AdminPermissions` and the v1 `Permissions` resources cannot be used against the
same Keycloak instance.

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: AdminPermissions
metadata:
  name: admins-manage-users
spec:
  deletionPolicy: Delete
  forProvider:
    name: admins-can-manage-users
    description: Admins can view and manage all users
    decisionStrategy: UNANIMOUS
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    scopes:
      - view
      - manage
      - map-roles
      - manage-group-membership
      - impersonate
  providerConfigRef:
    name: "keycloak-provider-config"
```

Valid `scopes` for user permissions are `view`, `manage`, `map-roles`,
`manage-group-membership` and `impersonate`. A permission without policies is
evaluated as "deny", so attach policies once they exist on the realm's
`admin-permissions` client, either by ID via `policies` or through the typed
reference fields (`groupPolicies`, `rolePolicies`, `userPolicies`, ...).

### Custom user federation

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: UserFederation
metadata:
  name: custom-user-federation
spec:
  forProvider:
    config:
      dummyBool: true
      dummyString: foobar
      multivalue: value1##value2
    enabled: true
    name: custom
    providerId: custom
    realmIdSelector:
      matchLabels:
        testing.upbound.io/example-name: realm
```

## Key Fields

| Resource | Field | Description |
|----------|-------|-------------|
| `User` | `realmId` | Realm where the user account exists. |
| `User` | `username` | Unique username in the realm. |
| `User` | `enabled` | Enables or disables login for the user. |
| `Groups` | `userIdRef` | Targets the user whose group memberships are managed. |
| `Groups` | `groupIdsRefs` | References groups to assign to the user. |
| `Roles` | `userIdRef` | Targets the user whose direct roles are managed. |
| `Roles` | `roleIdsRefs` | References roles to assign to the user. |
| `Permissions` | `realmIdRef` | Enables fine-grained admin permissions for user management in a realm. |
| `AdminPermissions` | `realmIdRef` | Realm whose users the permission applies to. |
| `AdminPermissions` | `scopes` | Admin operations the permission covers, e.g. `view`, `manage`. |
| `AdminPermissions` | `decisionStrategy` | How the attached policies are combined: `UNANIMOUS`, `AFFIRMATIVE` or `CONSENSUS`. |
| `AdminPermissions` | `policies` | IDs of authorization policies granting the permission. Typed `*Policies` reference fields are also available. |
| `UserFederation` | `providerId` | Selects the custom federation provider implementation. |
| `UserFederation` | `config` | Provider-specific configuration passed to the federation plugin. |
| `UserFederation` | `enabled` | Enables or disables the federation provider. |

## Related Resources

- [Groups](./groups.md)
- [Roles](./roles.md)
- [Default Configuration](./default-config.md)
- [User Federation](./user-federation.md)
