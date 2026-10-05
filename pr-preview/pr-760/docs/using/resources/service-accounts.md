# Service Accounts

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


Use these resources when a client needs to authenticate as itself for machine-to-machine access. They assign realm or client roles to a client&#39;s service account. The client must have `serviceAccountsEnabled: true`.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| ClientServiceAccountRealmRole | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_service_account_realm_role`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_service_account_realm_role) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientServiceAccountRealmRole/v1alpha1) |
| ClientServiceAccountRole | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_service_account_role`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_service_account_role) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientServiceAccountRole/v1alpha1) |

## Working YAML Examples

### `ClientServiceAccountRealmRole`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientServiceAccountRealmRole
metadata:
  name: service-account-realm-role
spec:
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
  deletionPolicy: Delete
  forProvider:
    realmId: &#34;dev&#34;
    role: &#34;svc-realm-role&#34;
    serviceAccountUserClientIdRef:
      name: &#34;service-acc-1&#34;
      policy:
        resolve: Always
```

### `ClientServiceAccountRole`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientServiceAccountRole
metadata:
  name: service-account-role
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
    roleRef:
      name: &#34;svc-role&#34;
      policy:
        resolve: Always
    serviceAccountUserClientIdRef:
      name: &#34;service-acc-1&#34;
      policy:
        resolve: Always
```

## Related Resources

- [Clients](./clients.md)
- [Roles](./roles.md)
- [Realms](./realms.md)


