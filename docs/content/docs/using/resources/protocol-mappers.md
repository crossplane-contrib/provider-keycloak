---
sidebar_position: 6
title: Protocol Mappers
description: Configure token claims and SAML assertions
---

# Protocol Mappers

Protocol mappers control what information is included in tokens (OIDC) or assertions (SAML).

## API Reference

> **Schema source:** This page highlights common fields and examples. For the complete OpenAPI schema, including references, selectors, status fields, and connection details, see the generated CRDs in `package/crds/`.

- **API Group**: `client.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kind**: `ProtocolMapper`

## OpenID Connect: User Attribute Mapper

Include a user attribute as a token claim:

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: department-mapper
spec:
  forProvider:
    clientId: my-client
    realmId: my-realm
    protocol: openid-connect
    protocolMapper: oidc-usermodel-attribute-mapper
    name: department-mapper
    config:
      "user.attribute": "department"
      "claim.name": "department"
      "id.token.claim": "true"
      "access.token.claim": "true"
  providerConfigRef:
    name: keycloak-provider-config
```

## OpenID Connect: Role Mapper

Include user roles in the token:

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: roles-mapper
spec:
  forProvider:
    clientId: my-client
    realmId: my-realm
    protocol: openid-connect
    protocolMapper: oidc-usermodel-realm-role-mapper
    name: roles
    config:
      "claim.name": "roles"
      "multivalued": "true"
      "id.token.claim": "true"
      "access.token.claim": "true"
  providerConfigRef:
    name: keycloak-provider-config
```

## OpenID Connect: Client Role Mapper

Include client-specific roles in the token:

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: client-roles-mapper
spec:
  forProvider:
    clientIdRef:
      name: kubernetes
    realmId: my-realm
    protocol: openid-connect
    protocolMapper: oidc-usermodel-client-role-mapper
    name: roles
    config:
      "usermodel.clientRoleMapping.clientId": "kubernetes"
      "claim.name": "roles"
      "multivalued": "true"
      "id.token.claim": "true"
      "access.token.claim": "true"
  providerConfigRef:
    name: keycloak-provider-config
```

## OpenID Connect: Audience Mapper

Add an audience claim to the token:

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: audience-mapper
spec:
  forProvider:
    clientId: my-client
    realmId: my-realm
    protocol: openid-connect
    protocolMapper: oidc-audience-mapper
    name: audience-mapper
    config:
      "included.client.audience": "target-client"
      "id.token.claim": "true"
  providerConfigRef:
    name: keycloak-provider-config
```

## SAML: User Property Mapper

Map a user property to a SAML assertion attribute:

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: saml-email-mapper
spec:
  forProvider:
    clientId: my-saml-client
    realmId: my-realm
    protocol: saml
    protocolMapper: saml-user-property-mapper
    name: email-mapper
    config:
      "property": "email"
      "friendly.name": "email"
      "attribute.name": "email"
      "attribute.nameformat": "Basic"
  providerConfigRef:
    name: keycloak-provider-config
```

## Common Mapper Types

| Protocol | Mapper Type | Purpose |
|----------|------------|---------|
| OIDC | `oidc-usermodel-attribute-mapper` | Map user attribute to claim |
| OIDC | `oidc-usermodel-realm-role-mapper` | Map realm roles to claim |
| OIDC | `oidc-usermodel-client-role-mapper` | Map client roles to claim |
| OIDC | `oidc-audience-mapper` | Add audience claim |
| OIDC | `oidc-full-name-mapper` | Map full name to claim |
| SAML | `saml-user-property-mapper` | Map user property to assertion |
| SAML | `saml-hardcoded-attribute-mapper` | Add static assertion attribute |
| SAML | `saml-x509-subject-name-mapper` | Map X.509 subject to assertion |
