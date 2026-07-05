---
sidebar_position: 6
title: Protocol Mappers
description: Manage OIDC and SAML protocol mappers for claims, roles, and group membership
---

Use these resources when you need to control what Keycloak emits in OIDC tokens or SAML assertions. Use `ProtocolMapper` for custom claim mapping, `RoleMapper` to include roles from one client or client scope in another, and `GroupMembershipProtocolMapper` to expose group membership as a JWT claim.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| ProtocolMapper | `client.keycloak.crossplane.io/v1alpha1` | [`keycloak_generic_protocol_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/generic_protocol_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/client.keycloak.crossplane.io/ProtocolMapper/v1alpha1) |
| RoleMapper | `client.keycloak.crossplane.io/v1alpha1` | [`keycloak_generic_role_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/generic_role_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/client.keycloak.crossplane.io/RoleMapper/v1alpha1) |
| GroupMembershipProtocolMapper | `openidgroup.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_group_membership_protocol_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_group_membership_protocol_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidgroup.keycloak.crossplane.io/GroupMembershipProtocolMapper/v1alpha1) |

## Working YAML Examples

### OIDC user attribute `ProtocolMapper` on a client

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: openid-client-protocol-mapper
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: "picture"
    protocol: "openid-connect"
    clientIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    protocolMapper: "oidc-usermodel-attribute-mapper"
    config:
      userinfo.token.claim: "true"
      user.attribute: "picture"
      id.token.claim: "true"
      access.token.claim: "true"
      claim.name: "picture"
      jsonType.label: "String"
      introspection.token.claim: "true"
```

### OIDC client role `ProtocolMapper` on a client scope

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: openid-client-scope-protocol-mapper
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: "client roles"
    protocol: "openid-connect"
    clientScopeIdRef:
      name: "openid-client-scope"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    protocolMapper: "oidc-usermodel-client-role-mapper"
    config:
      multivalued: "true"
      user.attribute: "foo"
      access.token.claim: "true"
      claim.name: "resource_access.${client_id}.roles"
      jsonType.label: "String"
      introspection.token.claim: "true"
```

### SAML role list `ProtocolMapper`

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: saml-client-protocol-mapper
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: "user roles"
    protocol: "saml"
    samlClientIdRef:
      name: "saml-client"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    protocolMapper: "saml-role-list-mapper"
    config:
      attribute.name: "Role"
      attribute.nameformat: "Basic"
      friendly.name: "test"
      single: "true"
```

### `RoleMapper` on a client

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: RoleMapper
metadata:
  name: openid-client-role-mapper
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    clientIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    roleIdRef:
      name: "test-client"
      policy:
        resolve: Always
```

### `RoleMapper` on a client scope

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: RoleMapper
metadata:
  name: openid-client-scope-role-mapper
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    clientScopeIdRef:
      name: "openid-client-scope"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    roleIdRef:
      name: "test-client"
      policy:
        resolve: Always
```

### `GroupMembershipProtocolMapper` on a client

```yaml
apiVersion: openidgroup.keycloak.crossplane.io/v1alpha1
kind: GroupMembershipProtocolMapper
metadata:
  name: openid-client-group-membership-protocol-mapper
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: "my-mapper"
    clientIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    claimName: "test"
```

### `GroupMembershipProtocolMapper` on a client scope

```yaml
apiVersion: openidgroup.keycloak.crossplane.io/v1alpha1
kind: GroupMembershipProtocolMapper
metadata:
  name: openid-client-scope-group-membership-protocol-mapper
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: "my-mapper"
    clientScopeIdRef:
      name: "openid-client-scope"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    claimName: "test"
```

## Related Resources

- [Clients](./clients.md)
- [OpenID Client Scopes](./openid-client-scopes.md)
- [Groups](./groups.md)
- [Roles](./roles.md)
- [SAML Clients](./saml-clients.md)
- [Realms](./realms.md)
