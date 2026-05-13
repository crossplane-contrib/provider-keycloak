---
sidebar_position: 10
title: OpenID Client Scopes
description: Manage OpenID Connect client scopes and scope mappings
---

# OpenID Client Scopes

Client scopes define sets of protocol mappers and role scope mappings that can be shared across multiple clients. Scopes can be assigned as default (always included) or optional (included on request).

## API Reference

- **API Group**: `openidclient.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kinds**: `ClientScope`, `ClientDefaultScopes`, `ClientOptionalScopes`

## ClientScope

Define a reusable scope with protocol mappers and configuration.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientScope
metadata:
  name: custom-scope
spec:
  forProvider:
    name: "custom-scope"
    description: "Custom scope for application-specific claims"
    realmId: "my-realm"
    includeInTokenScope: true
    consentScreenText: "Access your custom data"
    guiOrder: 1
  providerConfigRef:
    name: keycloak-provider-config
```

### ClientScope with Extra Config

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientScope
metadata:
  name: api-scope
spec:
  forProvider:
    name: "api-access"
    description: "API access scope"
    realmId: "my-realm"
    includeInTokenScope: true
    extraConfig:
      "display.on.consent.screen": "true"
      "consent.screen.text": "Access the API"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientDefaultScopes

Assign scopes that are always included in tokens for a client.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientDefaultScopes
metadata:
  name: web-app-default-scopes
spec:
  forProvider:
    clientId: "client-uuid"
    realmId: "my-realm"
    defaultScopes:
      - "profile"
      - "email"
      - "custom-scope"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientOptionalScopes

Assign scopes that clients can request at authentication time.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientOptionalScopes
metadata:
  name: web-app-optional-scopes
spec:
  forProvider:
    clientId: "client-uuid"
    realmId: "my-realm"
    optionalScopes:
      - "address"
      - "phone"
      - "offline_access"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

### ClientScope

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name of the scope |
| `description` | string | Description shown in the UI |
| `realmId` | string | Realm this scope belongs to |
| `includeInTokenScope` | bool | Include scope name in access token `scope` claim (default `true`) |
| `consentScreenText` | string | Text displayed on consent screen |
| `guiOrder` | number | Order in the GUI |
| `extraConfig` | map | Additional configuration attributes |

### ClientDefaultScopes

| Field | Type | Description |
|-------|------|-------------|
| `clientId` | string | UUID of the client |
| `realmId` | string | Realm the client and scopes belong to |
| `defaultScopes` | []string | List of scope names to assign as default |

### ClientOptionalScopes

| Field | Type | Description |
|-------|------|-------------|
| `clientId` | string | UUID of the client |
| `realmId` | string | Realm the client and scopes belong to |
| `optionalScopes` | []string | List of scope names to assign as optional |
