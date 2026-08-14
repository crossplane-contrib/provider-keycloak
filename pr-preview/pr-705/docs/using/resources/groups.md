# Groups

Use groups when multiple users should share the same roles or when you need a hierarchical structure such as teams, departments, or environments. Groups let you model organization structure once and then manage access in bulk.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| Group | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/Group/v1alpha1) |
| Memberships | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group_memberships`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group_memberships) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/Memberships/v1alpha1) |
| Roles | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group_roles`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group_roles) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/Roles/v1alpha1) |
| Permissions | `group.keycloak.crossplane.io/v1alpha1` | [`keycloak_group_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/group_permissions) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/group.keycloak.crossplane.io/Permissions/v1alpha1) |

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

## Related Resources

- [Users](./users.md)
- [Roles](./roles.md)
- [Default Configuration](./default-config.md)

