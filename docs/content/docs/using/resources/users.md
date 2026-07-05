---
sidebar_position: 3
title: Users
description: Manage users, user role mappings, group memberships, permissions, and custom user federation
---

# Users

Use `User` to declaratively manage people who can authenticate to Keycloak. Use `Groups`, `Roles`, and `Permissions` to manage access around those users, and `UserFederation` when you need a custom external user store integration.

## API Reference

| Kind | API Group | Terraform Resource |
|------|-----------|-------------------|
| User | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_user`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/user) |
| Groups | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_user_groups`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/user_groups) |
| Roles | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_user_roles`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/user_roles) |
| Permissions | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_users_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/users_permissions) |
| UserFederation | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_custom_user_federation`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/custom_user_federation) |

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
| `UserFederation` | `providerId` | Selects the custom federation provider implementation. |
| `UserFederation` | `config` | Provider-specific configuration passed to the federation plugin. |
| `UserFederation` | `enabled` | Enables or disables the federation provider. |

## Related Resources

- [Groups](./groups.md)
- [Roles](./roles.md)
- [Default Configuration](./default-config.md)
- [User Federation](./user-federation.md)
