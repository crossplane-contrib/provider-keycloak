---
sidebar_position: 14
title: SAML Clients
description: Manage SAML clients and client scopes
---

# SAML Clients

SAML (Security Assertion Markup Language) clients enable single sign-on for applications that use the SAML 2.0 protocol. This page covers SAML-specific client types and their scopes.

## API Reference

> **Schema source:** This page highlights common fields and examples. For the complete OpenAPI schema, including references, selectors, status fields, and connection details, see the generated CRDs in `package/crds/`.

- **API Group**: `samlclient.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kinds**: `Client`, `ClientScope`, `ClientDefaultScopes`

## SAML Client

```yaml
apiVersion: samlclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: saml-app
spec:
  forProvider:
    clientId: "https://saml-app.example.com/saml/metadata"
    name: "SAML Application"
    realmId: "my-realm"
    enabled: true
    signDocuments: true
    signAssertions: true
    clientSignatureRequired: true
    includeAuthnStatement: true
    nameIdFormat: "username"
    assertionConsumerPostUrl: "https://saml-app.example.com/saml/acs"
    logoutServicePostBindingUrl: "https://saml-app.example.com/saml/slo"
    validRedirectUris:
      - "https://saml-app.example.com/*"
  providerConfigRef:
    name: keycloak-provider-config
```

### SAML Client with Encryption

```yaml
apiVersion: samlclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: encrypted-saml-app
spec:
  forProvider:
    clientId: "https://secure-app.example.com/saml/metadata"
    name: "Encrypted SAML Application"
    realmId: "my-realm"
    enabled: true
    signDocuments: true
    signAssertions: true
    encryptAssertions: true
    encryptionAlgorithm: "AES-256"
    clientSignatureRequired: true
    canonicalizationMethod: "EXCLUSIVE"
    assertionConsumerPostUrl: "https://secure-app.example.com/saml/acs"
  providerConfigRef:
    name: keycloak-provider-config
```

## SAML ClientScope

Define a reusable SAML scope with protocol mappers.

```yaml
apiVersion: samlclient.keycloak.crossplane.io/v1alpha1
kind: ClientScope
metadata:
  name: saml-roles-scope
spec:
  forProvider:
    name: "saml-roles"
    description: "Scope for including role information in SAML assertions"
    realmId: "my-realm"
    consentScreenText: "Access your roles"
    guiOrder: 1
  providerConfigRef:
    name: keycloak-provider-config
```

## SAML ClientDefaultScopes

Assign default scopes to a SAML client.

```yaml
apiVersion: samlclient.keycloak.crossplane.io/v1alpha1
kind: ClientDefaultScopes
metadata:
  name: saml-app-default-scopes
spec:
  forProvider:
    clientId: "saml-client-uuid"
    realmId: "my-realm"
    defaultScopes:
      - "saml-roles"
      - "saml-attributes"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

### Client

| Field | Type | Description |
|-------|------|-------------|
| `clientId` | string | SAML entity ID (typically a URL) |
| `name` | string | Display name for the client |
| `realmId` | string | Realm this client belongs to |
| `enabled` | bool | Whether the client is active (default `true`) |
| `signDocuments` | bool | Sign SAML documents |
| `signAssertions` | bool | Sign SAML assertions |
| `clientSignatureRequired` | bool | Expect signed documents from client (default `true`) |
| `encryptAssertions` | bool | Encrypt SAML assertions |
| `encryptionAlgorithm` | string | Algorithm for assertion encryption |
| `assertionConsumerPostUrl` | string | SAML POST binding URL for assertions |
| `assertionConsumerRedirectUrl` | string | SAML Redirect binding URL for assertions |
| `nameIdFormat` | string | Name ID format (`username`, `email`, `transient`, `persistent`) |

### ClientScope

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name of the scope |
| `description` | string | Description shown in the UI |
| `realmId` | string | Realm this scope belongs to |
| `consentScreenText` | string | Text displayed on consent screen |
| `guiOrder` | number | Order in the GUI |

### ClientDefaultScopes

| Field | Type | Description |
|-------|------|-------------|
| `clientId` | string | UUID of the SAML client |
| `realmId` | string | Realm the client and scopes belong to |
| `defaultScopes` | []string | List of scope names to assign as default |
