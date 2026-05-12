---
sidebar_position: 1
title: SSO with ArgoCD
description: Configure single sign-on for ArgoCD using Keycloak
---

# SSO with ArgoCD

This guide demonstrates how to configure Keycloak as an OIDC provider for ArgoCD, enabling single sign-on with role-based access control.

## Overview

```
┌──────────┐     OIDC     ┌──────────────┐     Manages     ┌──────────┐
│  ArgoCD  │◄────────────►│   Keycloak   │◄────────────────│ Provider │
│   User   │              │              │                  │ Keycloak │
└──────────┘              └──────────────┘                  └──────────┘
```

## Step 1: Create the Realm

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: platform-realm
spec:
  forProvider:
    realm: "platform"
    enabled: true
    displayName: "Platform"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 2: Create the ArgoCD Client

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: argocd
spec:
  forProvider:
    name: argocd
    clientId: argocd
    realmId: platform
    accessType: CONFIDENTIAL
    standardFlowEnabled: true
    directAccessGrantsEnabled: true
    rootUrl: "https://argocd.example.com"
    adminUrl: "https://argocd.example.com"
    webOrigins:
      - "https://argocd.example.com"
    validRedirectUris:
      - "https://argocd.example.com/auth/callback"
    validPostLogoutRedirectUris:
      - "https://argocd.example.com"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 3: Create Roles

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: argocd-admin
spec:
  forProvider:
    realmId: "platform"
    name: "argocd-admin"
    description: "ArgoCD administrator"
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: argocd-readonly
spec:
  forProvider:
    realmId: "platform"
    name: "argocd-readonly"
    description: "ArgoCD read-only access"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 4: Create a Groups Claim Mapper

Map user roles to a `groups` claim that ArgoCD can consume:

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: argocd-roles-mapper
spec:
  forProvider:
    clientIdRef:
      name: argocd
    realmId: platform
    protocol: openid-connect
    protocolMapper: oidc-usermodel-realm-role-mapper
    name: roles
    config:
      "claim.name": "groups"
      "multivalued": "true"
      "id.token.claim": "true"
      "access.token.claim": "true"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 5: Create a Group and Assign Roles

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: platform-admins
spec:
  forProvider:
    name: "Platform Admins"
    realmId: "platform"
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: platform-admin-roles
spec:
  forProvider:
    groupIdRef:
      name: platform-admins
    roleIdsRefs:
      - name: argocd-admin
    realmId: platform
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 6: Configure ArgoCD

In your ArgoCD configuration (e.g., `argocd-cm` ConfigMap):

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
data:
  url: "https://argocd.example.com"
  oidc.config: |
    name: Keycloak
    issuer: https://keycloak.example.com/realms/platform
    clientID: argocd
    clientSecret: $oidc.keycloak.clientSecret
    requestedScopes:
      - openid
      - profile
      - email
      - roles
```

And in the ArgoCD RBAC policy (`argocd-rbac-cm`):

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
  namespace: argocd
data:
  policy.csv: |
    g, argocd-admin, role:admin
    g, argocd-readonly, role:readonly
```

## Result

Users who are members of the "Platform Admins" group in Keycloak will automatically get the `argocd-admin` role when they log in to ArgoCD via SSO.
