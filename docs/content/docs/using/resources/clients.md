---
sidebar_position: 2
title: Clients
description: Manage OpenID Connect clients for applications and services
---

# Clients

Use a `Client` when an application or service needs Keycloak to authenticate users with OpenID Connect. This is the resource for web apps, SPAs, backend services, service accounts, and federated workloads.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| Client | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/Client/v1alpha1) |

## Examples

### Confidential client with authorization

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: test
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    accessType: "CONFIDENTIAL"
    clientId: "test"
    fullScopeAllowed: false
    serviceAccountsEnabled: true
    authorization:
      - policyEnforcementMode: "PERMISSIVE"
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Managing a built-in client without deleting it

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: account
spec:
  managementPolicies: ["Create", "Update", "Observe"]
  forProvider:
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    accessType: "CONFIDENTIAL"
    clientId: "account"
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Service account client

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: service-acc-1
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    accessType: "CONFIDENTIAL"
    clientId: "service-acc-1"
    serviceAccountsEnabled: true
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Kubernetes federated JWT client

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
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
      name: "orgs"
      policy:
        resolve: Always
    serviceAccountsEnabled: true
    standardFlowEnabled: false
    extraConfig:
      federated.idp: k8s-federated
      federated.sub: system:serviceaccount:default:k8s-federated-test-sa
  providerConfigRef:
    name: "keycloak-provider-config"
```

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
