# Organizations

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-767/llms.txt)


Use `Organization` when you need Keycloak multi-tenancy support in Keycloak 26.6 and later. Organizations let you group users under tenant-like entities and configure domain-based identity provider routing. The realm must have `organizationsEnabled: true`. Use `Memberships` to manage which existing users belong to an organization.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| Organization | `organization.keycloak.crossplane.io/v1alpha1` | [`keycloak_organization`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/organization) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/organization.keycloak.crossplane.io/Organization/v1alpha1) |
| Memberships | `organization.keycloak.crossplane.io/v1alpha1` | [`keycloak_organization_memberships`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/organization_memberships) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/organization.keycloak.crossplane.io/Memberships/v1alpha1) |

## Working YAML Examples

### `Organization`

```yaml
apiVersion: organization.keycloak.crossplane.io/v1alpha1
kind: Organization
metadata:
  name: example
spec:
  deletionPolicy: Delete
  forProvider:
    realm: "orgs"
    name: example
    enabled: true
    domain:
      - name: example.com
      - name: example.org
  providerConfigRef:
    name: "keycloak-provider-config"
```

### `Memberships`

`members` is a list of usernames. Instead of listing them directly you can use
`membersRefs`/`membersSelector` to resolve them from `User` resources (the
referenced user's `spec.forProvider.username` is used).

```yaml
apiVersion: organization.keycloak.crossplane.io/v1alpha1
kind: Memberships
metadata:
  name: example-org-memberships
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: "orgs"
    organizationIdRef:
      name: example
    membersRefs:
      - name: example-user
  providerConfigRef:
    name: "keycloak-provider-config"
```

**Note:** `Memberships` is authoritative: unmanaged members of the organization that are
not listed in `members` are removed. Managed members (created through an
organization identity provider) are never removed, because Keycloak deletes the
underlying user account when a managed membership is removed: if the
organization has managed members that are not listed in `members`, create and
update fail until they are listed or removed through Keycloak.

## Related Resources

- [Realms](./realms.md)
- [Identity Providers](./identity-providers.md)
- [Users](./users.md)


