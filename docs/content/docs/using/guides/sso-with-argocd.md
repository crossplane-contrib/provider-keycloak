---
sidebar_position: 1
title: SSO with ArgoCD
description: Configure single sign-on for ArgoCD using Keycloak and provider-keycloak
---

# SSO with ArgoCD

This guide demonstrates how to configure Keycloak as an OIDC provider for ArgoCD, enabling single sign-on with role-based access control.

All manifests are available in [`examples/sso-argocd/`](https://github.com/crossplane-contrib/provider-keycloak/tree/main/examples/sso-argocd).

## Overview

```
┌──────────┐     OIDC     ┌──────────────┐     Manages     ┌──────────┐
│  ArgoCD  │◄────────────►│   Keycloak   │◄────────────────│ Provider │
│   User   │              │              │                  │ Keycloak │
└──────────┘              └──────────────┘                  └──────────┘
```

## Prerequisites

- A running Keycloak instance with provider-keycloak configured (see [Getting Started](/docs/using/getting-started/installation/))
- ArgoCD installed in your cluster (see [ArgoCD Getting Started](https://argo-cd.readthedocs.io/en/stable/getting_started/))
- A `ProviderConfig` named `keycloak-provider-config` pointing at your Keycloak

## Step 1: Create the Realm

```yaml title="examples/sso-argocd/01-realm.yaml"
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

The client must use `CONFIDENTIAL` access type with the authorization code flow. The `writeConnectionSecretToRef` field extracts the generated client secret into a Kubernetes Secret so you can feed it to ArgoCD.

```yaml title="examples/sso-argocd/02-client.yaml"
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: argocd-client
spec:
  forProvider:
    name: argocd
    clientId: argocd
    realmId: platform
    accessType: CONFIDENTIAL
    standardFlowEnabled: true
    directAccessGrantsEnabled: false
    rootUrl: "https://argocd.example.com"
    adminUrl: "https://argocd.example.com"
    baseUrl: "https://argocd.example.com"
    webOrigins:
      - "https://argocd.example.com"
    validRedirectUris:
      - "https://argocd.example.com/auth/callback"
    validPostLogoutRedirectUris:
      - "https://argocd.example.com"
  writeConnectionSecretToRef:
    name: argocd-keycloak-client-secret
    namespace: crossplane-system
  providerConfigRef:
    name: keycloak-provider-config
```

{{< callout type="info" >}}
Replace `https://argocd.example.com` with your actual ArgoCD URL throughout this guide.
{{< /callout >}}

## Step 3: Create Roles

```yaml title="examples/sso-argocd/03-roles.yaml"
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: argocd-admin
spec:
  forProvider:
    realmId: "platform"
    name: "argocd-admin"
    description: "ArgoCD administrator role"
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
    description: "ArgoCD read-only role"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 4: Create a Groups Claim Mapper

ArgoCD expects a `groups` claim in the OIDC token. Map realm roles into this claim:

```yaml title="examples/sso-argocd/04-protocol-mapper.yaml"
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: argocd-groups-mapper
spec:
  forProvider:
    clientIdRef:
      name: argocd-client
    realmId: platform
    protocol: openid-connect
    protocolMapper: oidc-usermodel-realm-role-mapper
    name: groups
    config:
      "claim.name": "groups"
      "multivalued": "true"
      "id.token.claim": "true"
      "access.token.claim": "true"
      "userinfo.token.claim": "true"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 5: Create Groups and Assign Roles

```yaml title="examples/sso-argocd/05-groups.yaml"
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: platform-admins
spec:
  forProvider:
    name: "platform-admins"
    realmId: "platform"
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: platform-viewers
spec:
  forProvider:
    name: "platform-viewers"
    realmId: "platform"
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: platform-admin-roles
spec:
  forProvider:
    realmId: "platform"
    groupIdRef:
      name: platform-admins
    roleIdsRefs:
      - name: argocd-admin
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: platform-viewer-roles
spec:
  forProvider:
    realmId: "platform"
    groupIdRef:
      name: platform-viewers
    roleIdsRefs:
      - name: argocd-readonly
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 6: Create Test Users and Assign to Groups

{{< callout type="warning" >}}
The passwords below are for demonstration only. In production, integrate with an existing identity provider (LDAP, SAML, social login) or use Keycloak's self-registration flow instead of static passwords.
{{< /callout >}}

First, create password secrets:

```bash
kubectl create secret generic argocd-admin-user-password \
  --namespace crossplane-system \
  --from-literal=******
kubectl create secret generic argocd-viewer-user-password \
  --namespace crossplane-system \
  --from-literal=******
```

```yaml title="examples/sso-argocd/06-users.yaml"
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: argocd-admin-user
spec:
  forProvider:
    realmId: "platform"
    username: "admin-user"
    email: "admin@example.com"
    firstName: "Admin"
    lastName: "User"
    enabled: true
    initialPassword:
      - valueSecretRef:
          name: argocd-admin-user-password
          namespace: crossplane-system
          key: password
        temporary: true
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: argocd-viewer-user
spec:
  forProvider:
    realmId: "platform"
    username: "viewer-user"
    email: "viewer@example.com"
    firstName: "Viewer"
    lastName: "User"
    enabled: true
    initialPassword:
      - valueSecretRef:
          name: argocd-viewer-user-password
          namespace: crossplane-system
          key: password
        temporary: true
  providerConfigRef:
    name: keycloak-provider-config
```

```yaml title="examples/sso-argocd/07-memberships.yaml"
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Memberships
metadata:
  name: platform-admin-members
spec:
  forProvider:
    realmId: "platform"
    groupIdRef:
      name: platform-admins
    members:
      - "admin-user"
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Memberships
metadata:
  name: platform-viewer-members
spec:
  forProvider:
    realmId: "platform"
    groupIdRef:
      name: platform-viewers
    members:
      - "viewer-user"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 7: Configure ArgoCD

### 7a. Store the client secret in ArgoCD

Retrieve the client secret from the Crossplane connection secret and add it to the `argocd-secret`:

```bash
CLIENT_SECRET=$(kubectl get secret argocd-keycloak-client-secret \
  -n crossplane-system \
  -o jsonpath='{.data.attribute\.client_secret}' | base64 -d)

kubectl -n argocd patch secret argocd-secret --type merge -p \
  "{\"stringData\": {\"oidc.keycloak.clientSecret\": \"${CLIENT_SECRET}\"}}"
```

### 7b. Configure ArgoCD OIDC

```yaml title="examples/sso-argocd/08-argocd-cm.yaml"
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
```

{{< callout type="info" >}}
ArgoCD resolves `$oidc.keycloak.clientSecret` from the `argocd-secret` Secret at runtime. See the [ArgoCD OIDC docs](https://argo-cd.readthedocs.io/en/stable/operator-manual/user-management/#existing-oidc-provider) for details.
{{< /callout >}}

### 7c. Configure ArgoCD RBAC

Map the Keycloak role names (from the `groups` claim) to ArgoCD roles:

```yaml title="examples/sso-argocd/09-argocd-rbac-cm.yaml"
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
  namespace: argocd
data:
  policy.default: role:readonly
  scopes: "[groups]"
  policy.csv: |
    g, argocd-admin, role:admin
    g, argocd-readonly, role:readonly
```

{{< callout type="warning" >}}
The `scopes` field must be set to `[groups]` so ArgoCD reads the `groups` claim from the OIDC token. Without this, RBAC policies won't match.
{{< /callout >}}

## Step 8: Restart ArgoCD

After updating the ConfigMaps, restart the ArgoCD server to pick up the changes:

```bash
kubectl -n argocd rollout restart deployment argocd-server
```

## Verification

1. Open ArgoCD at `https://argocd.example.com`
2. Click **"Log in via Keycloak"**
3. Log in as `admin-user` (password: `changeme`) → full admin access
4. Log in as `viewer-user` (password: `changeme`) → read-only access

You can decode the JWT token at [jwt.io](https://jwt.io) to verify the `groups` claim contains the expected roles:

```json
{
  "groups": ["argocd-admin"],
  "sub": "...",
  "email": "admin@example.com"
}
```

## Troubleshooting

| Issue | Fix |
|---|---|
| "Login failed" after redirect | Verify `validRedirectUris` includes your ArgoCD callback URL |
| No `groups` claim in token | Check the ProtocolMapper is attached to the correct client |
| RBAC not matching | Ensure `scopes: "[groups]"` is set in `argocd-rbac-cm` |
| "Invalid client secret" | Re-extract the secret from Crossplane and patch `argocd-secret` |
| Users not getting roles | Verify Group → Role assignment and User → Group membership |
