# Default Configuration

Use these resources when every new user in a realm should start with the same baseline access. `DefaultGroups` adds new users to groups automatically, and `Roles` assigns the realm roles that should always be present.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| DefaultGroups | `defaults.keycloak.crossplane.io/v1alpha1` | [`keycloak_default_groups`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/default_groups) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/defaults.keycloak.crossplane.io/DefaultGroups/v1alpha1) |
| Roles | `defaults.keycloak.crossplane.io/v1alpha1` | [`keycloak_default_roles`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/default_roles) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/defaults.keycloak.crossplane.io/Roles/v1alpha1) |

## Working YAML Examples

### `DefaultGroups`

```yaml
apiVersion: defaults.keycloak.crossplane.io/v1alpha1
kind: DefaultGroups
metadata:
  name: my-default-groups
spec:
  deletionPolicy: Delete
  forProvider:
    groupIdsRefs:
      - name: test
        policy:
          resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### `Roles`

```yaml
apiVersion: defaults.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: default-roles
spec:
  deletionPolicy: Delete
  forProvider:
    defaultRolesRefs:
      - name: test
        policy:
          resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Related Resources

- [Groups](./groups.md)
- [Roles](./roles.md)
- [Users](./users.md)
- [Realms](./realms.md)

