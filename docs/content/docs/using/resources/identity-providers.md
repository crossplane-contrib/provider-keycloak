---
sidebar_position: 7
title: Identity Providers
description: Configure external identity providers for federation
---

# Identity Providers

Identity providers enable users to authenticate via external systems (e.g., GitHub, Google, SAML IdPs).

## API Reference

> **Schema source:** This page highlights common fields and examples. For the complete OpenAPI schema, including references, selectors, status fields, and connection details, see the generated CRDs in `package/crds/`.

- **API Group**: `oidc.keycloak.crossplane.io` / `saml.keycloak.crossplane.io`
- **API Version**: `v1alpha1`

## OpenID Connect Identity Provider

Federate authentication with an external OIDC provider:

```yaml
apiVersion: oidc.keycloak.crossplane.io/v1alpha1
kind: IdentityProvider
metadata:
  name: github-idp
spec:
  forProvider:
    realm: "my-realm"
    alias: "github"
    displayName: "GitHub"
    enabled: true
    authorizationUrl: "https://github.com/login/oauth/authorize"
    tokenUrl: "https://github.com/login/oauth/access_token"
    clientIdSecretRef:
      key: "client-id"
      name: "github-idp-secret"
      namespace: "crossplane-system"
    clientSecretSecretRef:
      key: "client-secret"
      name: "github-idp-secret"
      namespace: "crossplane-system"
    defaultScopes: "user:email"
  providerConfigRef:
    name: keycloak-provider-config
```

## SAML Identity Provider

Federate authentication with an external SAML Identity Provider:

```yaml
apiVersion: saml.keycloak.crossplane.io/v1alpha1
kind: IdentityProvider
metadata:
  name: corporate-saml
spec:
  forProvider:
    realm: "my-realm"
    alias: "corporate-idp"
    displayName: "Corporate SSO"
    enabled: true
    singleSignOnServiceUrl: "https://idp.corp.example.com/saml/sso"
    postBindingResponse: true
    postBindingAuthnRequest: true
    wantAssertionsSigned: true
  providerConfigRef:
    name: keycloak-provider-config
```

## Identity Provider Mappers

Map claims from the external IdP to Keycloak user attributes:

```yaml
apiVersion: identityprovider.keycloak.crossplane.io/v1alpha1
kind: IdentityProviderMapper
metadata:
  name: github-email-mapper
spec:
  forProvider:
    realm: "my-realm"
    identityProviderAlias: "github"
    identityProviderMapper: "oidc-user-attribute-idp-mapper"
    name: "email-mapper"
    extraConfig:
      claim: "email"
      userAttribute: "email"
      syncMode: "INHERIT"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `realm` | string | Realm for the identity provider |
| `alias` | string | Unique alias for the IdP |
| `enabled` | bool | Whether the IdP is active |
| `displayName` | string | Label shown on the login page |
| `authorizationUrl` | string | OIDC authorization endpoint |
| `tokenUrl` | string | OIDC token endpoint |
| `clientIdSecretRef` | object | Secret reference for OAuth2 client ID |
| `defaultScopes` | string | Scopes to request from the external IdP |
