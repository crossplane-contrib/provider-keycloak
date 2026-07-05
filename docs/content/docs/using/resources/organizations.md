---
sidebar_position: 13
title: Organizations
description: Manage Keycloak organizations for multi-tenant realms
---

Use `Organization` when you need Keycloak multi-tenancy support in Keycloak 26.6 and later. Organizations let you group users under tenant-like entities and configure domain-based identity provider routing. The realm must have `organizationsEnabled: true`.

## API Reference

| Kind | API Group | Terraform Resource |
|------|-----------|-------------------|
| Organization | `organization.keycloak.crossplane.io/v1alpha1` | [`keycloak_organization`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/organization) |

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

## Related Resources

- [Realms](./realms.md)
- [Identity Providers](./identity-providers.md)
- [Users](./users.md)
