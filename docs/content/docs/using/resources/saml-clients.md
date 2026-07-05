---
sidebar_position: 14
title: SAML Clients
description: Configure SAML 2.0 clients, scopes, and default scopes in Keycloak
---

# SAML Clients

Use SAML clients when an application or service provider expects SAML 2.0 instead of OpenID Connect. This is common for legacy enterprise applications and commercial service providers such as Salesforce, Jira, or other platforms that rely on SAML metadata, signed assertions, and SSO endpoints.

## API Reference

| Kind | API Group | Terraform |
|------|-----------|-----------|
| `Client` | `samlclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_saml_client`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/saml_client) |
| `ClientScope` | `samlclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_saml_client_scope`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/saml_client_scope) |
| `ClientDefaultScopes` | `samlclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_saml_client_default_scopes`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/saml_client_default_scopes) |

## Working YAML examples

### SAML Client

Use a SAML client to represent the service provider that trusts Keycloak for SSO.

```yaml
apiVersion: samlclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: saml-client
spec:
  deletionPolicy: Delete
  forProvider:
    clientId: saml-client-id
    includeAuthnStatement: true
    name: saml-client
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    signAssertions: true
    signDocuments: false
    signingCertificateSecretRef:
      name: rsa-key
      namespace: dev
      key: cert
    signingPrivateKeySecretRef:
      name: saml-cliet-cert
      namespace: dev
      key: priv
  providerConfigRef:
    name: "keycloak-provider-config"
```

### SAML Client Scope

Use a SAML client scope to define reusable mapper and assertion behavior that can be shared across clients.

```yaml
apiVersion: samlclient.keycloak.crossplane.io/v1alpha1
kind: ClientScope
metadata:
  name: saml-client-scopes
spec:
  deletionPolicy: Delete
  forProvider:
    description: This scope will map a user's group memberships to SAML assertion
    guiOrder: 1
    name: groups
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### SAML Client Default Scopes

Use default scopes to attach one or more SAML client scopes automatically to a client.

```yaml
apiVersion: samlclient.keycloak.crossplane.io/v1alpha1
kind: ClientDefaultScopes
metadata:
  name: saml-client-default-scopes
spec:
  deletionPolicy: Delete
  forProvider:
    clientIdRef:
      name: saml-client
      policy:
        resolve: Always
    defaultScopes:
      - groups
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Key fields

### Client

| Field | Why it matters |
|-------|----------------|
| `clientId` | SAML entity ID for the service provider. |
| `realmIdRef` | Selects the realm that owns the client. |
| `includeAuthnStatement` | Adds an AuthnStatement to issued assertions when required by the application. |
| `signAssertions` | Signs SAML assertions sent to the service provider. |
| `signDocuments` | Signs the SAML document envelope when the integration requires it. |
| `signingCertificateSecretRef` | Supplies the public certificate used for signing. |
| `signingPrivateKeySecretRef` | Supplies the private key used for signing. |

### ClientScope

| Field | Why it matters |
|-------|----------------|
| `name` | Scope name referenced by SAML clients. |
| `description` | Explains the purpose of the scope in Keycloak. |
| `guiOrder` | Controls display order in the Keycloak admin UI. |
| `realmIdRef` | Selects the realm that owns the scope. |

### ClientDefaultScopes

| Field | Why it matters |
|-------|----------------|
| `clientIdRef` | Resolves the SAML client that should receive the scopes. |
| `defaultScopes` | Lists the scope names that are applied automatically. |
| `realmIdRef` | Ensures the client and scopes are resolved in the right realm. |

## Related Resources

- **[Clients](./clients.md)** — Compare SAML clients with OpenID Connect clients.
- **[Protocol Mappers](./protocol-mappers.md)** — Add mappers that shape SAML assertions and attributes.
- **[Realms](./realms.md)** — Create the realm that owns the client and scopes.

