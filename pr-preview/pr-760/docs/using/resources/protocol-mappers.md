# Protocol Mappers

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


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
    name: &#34;keycloak-provider-config&#34;
  deletionPolicy: Delete
  forProvider:
    name: &#34;picture&#34;
    protocol: &#34;openid-connect&#34;
    clientIdRef:
      name: &#34;test&#34;
      policy:
        resolve: Always
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    protocolMapper: &#34;oidc-usermodel-attribute-mapper&#34;
    config:
      userinfo.token.claim: &#34;true&#34;
      user.attribute: &#34;picture&#34;
      id.token.claim: &#34;true&#34;
      access.token.claim: &#34;true&#34;
      claim.name: &#34;picture&#34;
      jsonType.label: &#34;String&#34;
      introspection.token.claim: &#34;true&#34;
```

### OIDC client role `ProtocolMapper` on a client scope

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: openid-client-scope-protocol-mapper
spec:
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
  deletionPolicy: Delete
  forProvider:
    name: &#34;client roles&#34;
    protocol: &#34;openid-connect&#34;
    clientScopeIdRef:
      name: &#34;openid-client-scope&#34;
      policy:
        resolve: Always
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    protocolMapper: &#34;oidc-usermodel-client-role-mapper&#34;
    config:
      multivalued: &#34;true&#34;
      user.attribute: &#34;foo&#34;
      access.token.claim: &#34;true&#34;
      claim.name: &#34;resource_access.${client_id}.roles&#34;
      jsonType.label: &#34;String&#34;
      introspection.token.claim: &#34;true&#34;
```

### SAML role list `ProtocolMapper`

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: saml-client-protocol-mapper
spec:
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
  deletionPolicy: Delete
  forProvider:
    name: &#34;user roles&#34;
    protocol: &#34;saml&#34;
    samlClientIdRef:
      name: &#34;saml-client&#34;
      policy:
        resolve: Always
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    protocolMapper: &#34;saml-role-list-mapper&#34;
    config:
      attribute.name: &#34;Role&#34;
      attribute.nameformat: &#34;Basic&#34;
      friendly.name: &#34;test&#34;
      single: &#34;true&#34;
```

### `RoleMapper` on a client

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: RoleMapper
metadata:
  name: openid-client-role-mapper
spec:
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
  deletionPolicy: Delete
  forProvider:
    clientIdRef:
      name: &#34;test&#34;
      policy:
        resolve: Always
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    roleIdRef:
      name: &#34;test-client&#34;
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
    name: &#34;keycloak-provider-config&#34;
  deletionPolicy: Delete
  forProvider:
    clientScopeIdRef:
      name: &#34;openid-client-scope&#34;
      policy:
        resolve: Always
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    roleIdRef:
      name: &#34;test-client&#34;
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
    name: &#34;keycloak-provider-config&#34;
  deletionPolicy: Delete
  forProvider:
    name: &#34;my-mapper&#34;
    clientIdRef:
      name: &#34;test&#34;
      policy:
        resolve: Always
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    claimName: &#34;test&#34;
```

### `GroupMembershipProtocolMapper` on a client scope

```yaml
apiVersion: openidgroup.keycloak.crossplane.io/v1alpha1
kind: GroupMembershipProtocolMapper
metadata:
  name: openid-client-scope-group-membership-protocol-mapper
spec:
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
  deletionPolicy: Delete
  forProvider:
    name: &#34;my-mapper&#34;
    clientScopeIdRef:
      name: &#34;openid-client-scope&#34;
      policy:
        resolve: Always
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    claimName: &#34;test&#34;
```

## Related Resources

- [Clients](./clients.md)
- [OpenID Client Scopes](./openid-client-scopes.md)
- [Groups](./groups.md)
- [Roles](./roles.md)
- [SAML Clients](./saml-clients.md)
- [Realms](./realms.md)


