# Clients

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


Use a `Client` when an application or service needs Keycloak to authenticate users with OpenID Connect. This is the resource for web apps, SPAs, backend services, service accounts, and federated workloads.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| Client | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/Client/v1alpha1) |
| ClientAdminPermissions | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_admin_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_admin_permissions) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientAdminPermissions/v1alpha1) |

## Examples

### Confidential client with authorization

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
kind: Client
metadata:
  name: test
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    accessType: &#34;CONFIDENTIAL&#34;
    clientId: &#34;test&#34;
    fullScopeAllowed: false
    serviceAccountsEnabled: true
    authorization:
      - policyEnforcementMode: &#34;PERMISSIVE&#34;
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
```

### Managing a built-in client without deleting it

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
kind: Client
metadata:
  name: account
spec:
  managementPolicies: [&#34;Create&#34;, &#34;Update&#34;, &#34;Observe&#34;]
  forProvider:
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    accessType: &#34;CONFIDENTIAL&#34;
    clientId: &#34;account&#34;
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
```

### Managing another built-in client (account-console)

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
kind: Client
metadata:
  name: account-console
spec:
  managementPolicies: [Observe, Update]
  deletionPolicy: Orphan
  forProvider:
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    accessType: &#34;PUBLIC&#34;
    clientId: &#34;account-console&#34;
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
```

### Service account client

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
kind: Client
metadata:
  name: service-acc-1
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    accessType: &#34;CONFIDENTIAL&#34;
    clientId: &#34;service-acc-1&#34;
    serviceAccountsEnabled: true
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
```

### Kubernetes federated JWT client

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha2
kind: Client
metadata:
  name: k8s-federated-client
spec:
  deletionPolicy: Delete
  forProvider:
    accessType: CONFIDENTIAL
    clientAuthenticatorType: federated-jwt
    clientId: k8s-federated-client
    enabled: true
    name: k8s-federated-client
    realmIdRef:
      name: &#34;orgs&#34;
      policy:
        resolve: Always
    serviceAccountsEnabled: true
    standardFlowEnabled: false
    extraConfig:
      federated.idp: k8s-federated
      federated.sub: system:serviceaccount:default:k8s-federated-test-sa
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
```

### Fine-grained admin permissions (v2)

`ClientAdminPermissions` manages a single fine-grained admin permission for the
clients of a realm. It requires Keycloak 26.2 or newer started with the
`admin-fine-grained-authz:v2` feature and a realm with
`adminPermissionsEnabled: true`. Keycloak then creates an `admin-permissions`
client for the realm that acts as the resource server for all of its admin
permissions.

`admin-fine-grained-authz:v2` replaces `admin-fine-grained-authz:v1`, so
`ClientAdminPermissions` and the v1 permission resources cannot be used against
the same Keycloak instance.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientAdminPermissions
metadata:
  name: admins-manage-clients
spec:
  deletionPolicy: Delete
  forProvider:
    name: admins-can-manage-clients
    description: Admins can view and manage the referenced client
    decisionStrategy: UNANIMOUS
    realmIdRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    clientIdsRefs:
      - name: &#34;test&#34;
    scopes:
      - view
      - manage
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
```

Without `clientIds` the permission applies to every client of the realm,
otherwise only to the referenced clients. Both OpenID clients (`clientIdsRefs`)
and SAML clients (`samlClientIdsRefs`) can be referenced. A permission without
policies is evaluated as &#34;deny&#34;, so attach policies once they exist on the
realm&#39;s `admin-permissions` client, either by ID via `policies` or through the
typed reference fields (`groupPolicies`, `rolePolicies`, `userPolicies`, ...).

## Key Fields

| Field | Description |
|-------|-------------|
| `accessType` | Client type. Use `CONFIDENTIAL` for server-side apps, `PUBLIC` for browser or native apps, and `BEARER-ONLY` for APIs that only validate tokens. |
| `clientId` | Unique client identifier in the realm. |
| `serviceAccountsEnabled` | Enables a service account so the client can use client credentials flows. |
| `fullScopeAllowed` | Controls whether the client automatically receives all realm and client scopes. |
| `authorization` | Enables and configures Keycloak Authorization Services for the client. |
| `standardFlowEnabled` | Enables the authorization code flow. |
| `implicitFlowEnabled` | Enables the implicit flow for legacy browser-based integrations. |
| `directAccessGrantsEnabled` | Enables direct username/password token grants. |
| `clientAuthenticatorType` | Selects how the client authenticates, such as standard secret-based auth or `federated-jwt`. |

## Related Resources

- [OpenID Client Scopes](./openid-client-scopes.md)
- [Client Authorization](./client-authorization.md)
- [Service Accounts](./service-accounts.md)
- [SAML Clients](./saml-clients.md)
- [Protocol Mappers](./protocol-mappers.md)


