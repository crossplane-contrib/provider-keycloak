---
sidebar_position: 10
title: OpenID Client Scopes
description: Manage reusable OpenID Connect client scopes and assign them as default or optional scopes
---

Use these resources when you want to group protocol mappers and role scope mappings so they can be reused across multiple clients. Default scopes are always included in tokens, while optional scopes are only added when explicitly requested.

## API Reference

| Kind | API Group | Terraform Resource |
|------|-----------|-------------------|
| ClientScope | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_scope`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_scope) |
| ClientDefaultScopes | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_default_scopes`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_default_scopes) |
| ClientOptionalScopes | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_optional_scopes`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_optional_scopes) |

## Working YAML Examples

### `ClientScope`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientScope
metadata:
  name: openid-client-scope
spec:
  deletionPolicy: Delete
  providerConfigRef:
    name: "keycloak-provider-config"
  forProvider:
    description: When requested, this scope will map a user's group memberships to a claim
    guiOrder: 1
    includeInTokenScope: true
    name: my-groups
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

### `ClientDefaultScopes`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientDefaultScopes
metadata:
  name: client-default-scopes
spec:
  deletionPolicy: Delete
  forProvider:
    clientIdRef:
      name: "test"
      policy:
        resolve: Always
    defaultScopes:
      - profile
      - email
      - roles
      - web-origins
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### `ClientOptionalScopes`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientOptionalScopes
metadata:
  name: client-optional-scopes
spec:
  deletionPolicy: Delete
  forProvider:
    clientIdRef:
      name: "test"
      policy:
        resolve: Always
    optionalScopes:
      - address
      - phone
      - offline_access
      - microprofile-jwt
      - my-groups
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Related Resources

- [Clients](./clients.md)
- [Protocol Mappers](./protocol-mappers.md)
- [Roles](./roles.md)
- [Realms](./realms.md)
