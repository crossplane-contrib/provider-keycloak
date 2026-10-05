# OpenID Client Scopes

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


Use these resources when you want to group protocol mappers and role scope mappings so they can be reused across multiple clients. Default scopes are always included in tokens, while optional scopes are only added when explicitly requested.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| ClientScope | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_scope`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_scope) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientScope/v1alpha1) |
| ClientDefaultScopes | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_default_scopes`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_default_scopes) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientDefaultScopes/v1alpha1) |
| ClientOptionalScopes | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_optional_scopes`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_optional_scopes) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientOptionalScopes/v1alpha1) |

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
    name: &#34;keycloak-provider-config&#34;
  forProvider:
    description: When requested, this scope will map a user&#39;s group memberships to a claim
    guiOrder: 1
    includeInTokenScope: true
    name: my-groups
    realmIdRef:
      name: &#34;dev&#34;
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
      name: &#34;test&#34;
      policy:
        resolve: Always
    defaultScopes:
      - profile
      - email
      - roles
      - web-origins
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
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
      name: &#34;test&#34;
      policy:
        resolve: Always
    optionalScopes:
      - address
      - phone
      - offline_access
      - microprofile-jwt
      - my-groups
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
```

## Related Resources

- [Clients](./clients.md)
- [Protocol Mappers](./protocol-mappers.md)
- [Roles](./roles.md)
- [Realms](./realms.md)


