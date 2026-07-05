---
sidebar_position: 12
title: Service Accounts
description: Assign realm and client roles to OpenID client service accounts
---

Use these resources when a client needs to authenticate as itself for machine-to-machine access. They assign realm or client roles to a client's service account. The client must have `serviceAccountsEnabled: true`.

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
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    realmId: "dev"
    role: "svc-realm-role"
    serviceAccountUserClientIdRef:
      name: "service-acc-1"
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
    roleRef:
      name: "svc-role"
      policy:
        resolve: Always
    serviceAccountUserClientIdRef:
      name: "service-acc-1"
      policy:
        resolve: Always
```

## Related Resources

- [Clients](./clients.md)
- [Roles](./roles.md)
- [Realms](./realms.md)
