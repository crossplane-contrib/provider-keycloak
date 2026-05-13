---
sidebar_position: 2
title: Clients
description: Manage OpenID Connect and SAML clients
---

# Clients

Clients are applications and services that can request authentication. Keycloak supports OpenID Connect (OIDC) and SAML protocols.

## API Reference

- **API Group**: `openidclient.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kind**: `Client`

## Public Client (Web Application)

For browser-based applications using the Authorization Code flow:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: web-app
spec:
  forProvider:
    clientId: web-app
    realmId: my-realm
    accessType: public
    standardFlowEnabled: true
    validRedirectUris:
      - "http://localhost:3000/callback"
      - "https://app.example.com/callback"
  providerConfigRef:
    name: keycloak-provider-config
```

## Confidential Client (Backend Service)

For server-side applications that can securely store a client secret:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: backend-service
spec:
  forProvider:
    clientId: backend-service
    realmId: my-realm
    accessType: confidential
    serviceAccountsEnabled: true
    standardFlowEnabled: false
  providerConfigRef:
    name: keycloak-provider-config
```

## Bearer-Only Client (Resource Server)

For APIs that only validate tokens and never initiate login:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: api-server
spec:
  forProvider:
    clientId: api-server
    realmId: my-realm
    accessType: bearer-only
  providerConfigRef:
    name: keycloak-provider-config
```

## Mobile Application Client (with PKCE)

For native/mobile applications using PKCE for enhanced security:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: mobile-app
spec:
  forProvider:
    clientId: mobile-app
    realmId: my-realm
    accessType: public
    standardFlowEnabled: true
    pkceCodeChallengeMethod: S256
    validRedirectUris:
      - "myapp://callback"
  providerConfigRef:
    name: keycloak-provider-config
```

## Confidential Client with Authorization

For applications that need fine-grained authorization:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: kubernetes
spec:
  forProvider:
    clientId: kubernetes
    name: kubernetes
    realmId: my-realm
    accessType: CONFIDENTIAL
    standardFlowEnabled: true
    directAccessGrantsEnabled: true
    serviceAccountsEnabled: true
    authorization:
      - policyEnforcementMode: PERMISSIVE
    validRedirectUris:
      - http://localhost:18000
      - http://localhost:8000
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `clientId` | string | Client identifier |
| `realmId` | string | Realm this client belongs to |
| `accessType` | string | `public`, `confidential`, or `bearer-only` |
| `standardFlowEnabled` | bool | Enable Authorization Code flow |
| `directAccessGrantsEnabled` | bool | Enable Resource Owner Password Credentials |
| `serviceAccountsEnabled` | bool | Enable Client Credentials grant |
| `validRedirectUris` | []string | Allowed redirect URIs after login |
| `webOrigins` | []string | Allowed CORS origins |
| `pkceCodeChallengeMethod` | string | PKCE method (`S256` recommended) |
| `authorization` | object | Fine-grained authorization settings |

## Related Resources

- **[OpenID Client Scopes](./openid-client-scopes.md)** — Manage `ClientScope`, `ClientDefaultScopes`, and `ClientOptionalScopes`
- **[Client Authorization](./client-authorization.md)** — Fine-grained authorization resources, permissions, and policies
- **[Service Accounts](./service-accounts.md)** — Assign realm and client roles to service accounts
- **[SAML Clients](./saml-clients.md)** — SAML protocol clients and scopes

## Deletion Policy

Control what happens when you delete the Kubernetes resource:

```yaml
spec:
  deletionPolicy: Delete   # Remove from Keycloak (default)
  # deletionPolicy: Orphan # Leave in Keycloak
```
