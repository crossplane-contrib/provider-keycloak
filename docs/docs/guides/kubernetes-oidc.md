---
sidebar_position: 2
title: Kubernetes OIDC Authentication
description: Use Keycloak as an OIDC provider for Kubernetes API server authentication
---

# Kubernetes OIDC Authentication

This guide shows how to configure Keycloak as an OIDC provider for Kubernetes, enabling users to authenticate to the Kubernetes API using their Keycloak credentials.

## Overview

```
┌──────────┐    kubectl    ┌────────────────┐    OIDC verify    ┌──────────────┐
│   User   │──────────────►│  K8s API Server │◄─────────────────│   Keycloak   │
└──────────┘               └────────────────┘                   └──────────────┘
                                                                       ▲
                                                                       │ Manages
                                                                ┌──────────────┐
                                                                │   Provider   │
                                                                │   Keycloak   │
                                                                └──────────────┘
```

## Step 1: Create the Kubernetes Client

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: kubernetes
spec:
  forProvider:
    name: kubernetes
    clientId: kubernetes
    realmId: my-realm
    accessType: CONFIDENTIAL
    standardFlowEnabled: true
    directAccessGrantsEnabled: true
    serviceAccountsEnabled: true
    authorization:
      - policyEnforcementMode: PERMISSIVE
    validRedirectUris:
      - "http://localhost:18000"
      - "http://localhost:8000"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 2: Create Client Roles

```yaml
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: k8s-admin
spec:
  forProvider:
    realmId: "my-realm"
    name: "k8s-admin"
    clientId: "kubernetes"
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: k8s-viewer
spec:
  forProvider:
    realmId: "my-realm"
    name: "k8s-viewer"
    clientId: "kubernetes"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 3: Configure Role Mapper

Map client roles to a `roles` claim in the token:

```yaml
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: kubernetes-roles
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
---
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: kubernetes-name
spec:
  forProvider:
    clientIdRef:
      name: kubernetes
    realmId: my-realm
    protocol: openid-connect
    protocolMapper: oidc-usermodel-attribute-mapper
    name: name
    config:
      "claim.name": "name"
      "user.attribute": "name"
      "id.token.claim": "true"
      "access.token.claim": "true"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 4: Create Groups and Assign Roles

```yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: k8s-admins
spec:
  forProvider:
    name: "Kubernetes Admins"
    realmId: "my-realm"
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: k8s-admin-role-mapping
spec:
  forProvider:
    groupIdRef:
      name: k8s-admins
    roleIdsRefs:
      - name: k8s-admin
    realmId: my-realm
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 5: Configure the Kubernetes API Server

Add the following flags to your Kubernetes API server:

```
--oidc-issuer-url=https://keycloak.example.com/realms/my-realm
--oidc-client-id=kubernetes
--oidc-username-claim=preferred_username
--oidc-groups-claim=roles
```

## Step 6: Create Kubernetes RBAC

Bind Keycloak roles to Kubernetes RBAC:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: keycloak-k8s-admin
subjects:
  - kind: Group
    name: k8s-admin
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: keycloak-k8s-viewer
subjects:
  - kind: Group
    name: k8s-viewer
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: view
  apiGroup: rbac.authorization.k8s.io
```

## Using kubectl with OIDC

Use [kubelogin](https://github.com/int128/kubelogin) for seamless OIDC authentication:

```bash
kubectl oidc-login setup \
  --oidc-issuer-url=https://keycloak.example.com/realms/my-realm \
  --oidc-client-id=kubernetes \
  --oidc-client-secret=<client-secret>
```
