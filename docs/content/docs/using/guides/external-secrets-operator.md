---
sidebar_position: 4
title: External Secrets Operator
description: Combine provider-keycloak with External Secrets Operator for client secret management
---

# External Secrets Operator Integration

This guide shows how to use [External Secrets Operator (ESO)](https://external-secrets.io/) alongside provider-keycloak to manage client secrets. A common pattern is to create a Keycloak Client with provider-keycloak and then use ESO's templating to assemble a Kubernetes Secret containing both the client ID and the generated client secret.

## Overview

```
┌──────────────────┐         ┌──────────────┐
│ provider-keycloak│────────►│   Keycloak   │
│  (Client CRD)    │         │              │
└──────────────────┘         └──────┬───────┘
                                    │ client_secret
                                    ▼
                             ┌──────────────┐        ┌─────────────────┐
                             │  K8s Secret  │◄───────│ ExternalSecret  │
                             │  (assembled) │        │  (ESO + tmpl)   │
                             └──────────────┘        └─────────────────┘
```

**Why this pattern?**
- provider-keycloak creates and reconciles the Client in Keycloak
- Keycloak generates the `client_secret` for confidential clients
- ESO fetches the secret from Keycloak (or from the Kubernetes Secret the provider writes) and templates it into the shape your application expects
- Your application gets a single Secret with all connection details

## Prerequisites

- [provider-keycloak installed and configured](../getting-started/installation.md)
- [External Secrets Operator](https://external-secrets.io/latest/introduction/getting-started/) installed in your cluster
- A Keycloak realm already created

## Step 1: Create a Confidential Client

Create the OIDC client with provider-keycloak. Keycloak will generate a client secret automatically for confidential clients:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: my-app
spec:
  forProvider:
    clientId: my-app
    name: my-app
    realmId: my-realm
    accessType: CONFIDENTIAL
    standardFlowEnabled: true
    validRedirectUris:
      - "https://my-app.example.com/callback"
  writeConnectionSecretToRef:
    name: my-app-client-conn
    namespace: crossplane-system
  providerConfigRef:
    name: keycloak-provider-config
```

The provider writes connection details (including the client secret) to the Kubernetes Secret specified in `writeConnectionSecretToRef`.

## Step 2: Set Up a SecretStore

Configure ESO to read from Kubernetes Secrets (using the `kubernetes` provider):

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: k8s-store
  namespace: my-app-namespace
spec:
  provider:
    kubernetes:
      remoteNamespace: crossplane-system
      server:
        caProvider:
          type: ConfigMap
          name: kube-root-ca.crt
          key: ca.crt
      auth:
        serviceAccount:
          name: eso-reader
```

Or, if you prefer a `ClusterSecretStore` for cluster-wide access:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: crossplane-secrets
spec:
  provider:
    kubernetes:
      remoteNamespace: crossplane-system
      server:
        caProvider:
          type: ConfigMap
          name: kube-root-ca.crt
          key: ca.crt
          namespace: kube-system
      auth:
        serviceAccount:
          name: eso-reader
          namespace: external-secrets
```

## Step 3: Create an ExternalSecret with Templating

Use ESO's templating to assemble a Secret with all the fields your application needs:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: my-app-oidc
  namespace: my-app-namespace
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: k8s-store
    kind: SecretStore
  target:
    name: my-app-oidc-secret
    template:
      engineVersion: v2
      data:
        OIDC_ISSUER_URL: "https://keycloak.example.com/realms/my-realm"
        OIDC_CLIENT_ID: "my-app"
        OIDC_CLIENT_SECRET: "{{ .clientSecret }}"
        OIDC_REDIRECT_URI: "https://my-app.example.com/callback"
  data:
    - secretKey: clientSecret
      remoteRef:
        key: my-app-client-conn
        property: attribute.client_secret
```

The resulting Secret `my-app-oidc-secret` will contain all OIDC configuration your application needs, ready to mount as environment variables.

## Helm Chart Pattern

When deploying with Helm, you can bundle the Client and ExternalSecret together. This is useful because provider-keycloak does not manage application-side Secrets — it only manages Keycloak resources.

```yaml
# templates/keycloak-client.yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: {{ .Release.Name }}
spec:
  forProvider:
    clientId: {{ .Values.oidc.clientId }}
    name: {{ .Release.Name }}
    realmId: {{ .Values.oidc.realm }}
    accessType: CONFIDENTIAL
    standardFlowEnabled: true
    validRedirectUris:
      {{- range .Values.oidc.redirectUris }}
      - {{ . | quote }}
      {{- end }}
  writeConnectionSecretToRef:
    name: {{ .Release.Name }}-kc-conn
    namespace: crossplane-system
  providerConfigRef:
    name: {{ .Values.oidc.providerConfigRef }}
```

```yaml
# templates/external-secret.yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: {{ .Release.Name }}-oidc
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: {{ .Values.eso.secretStoreName }}
    kind: {{ .Values.eso.secretStoreKind | default "ClusterSecretStore" }}
  target:
    name: {{ .Release.Name }}-oidc-secret
    template:
      engineVersion: v2
      data:
        OIDC_ISSUER_URL: {{ printf "https://%s/realms/%s" .Values.oidc.keycloakHost .Values.oidc.realm | quote }}
        OIDC_CLIENT_ID: {{ .Values.oidc.clientId | quote }}
        OIDC_CLIENT_SECRET: "{{ `{{ .clientSecret }}` }}"
  data:
    - secretKey: clientSecret
      remoteRef:
        key: {{ .Release.Name }}-kc-conn
        property: attribute.client_secret
```

```yaml
# values.yaml
oidc:
  clientId: my-app
  realm: my-realm
  keycloakHost: keycloak.example.com
  providerConfigRef: keycloak-provider-config
  redirectUris:
    - "https://my-app.example.com/callback"
eso:
  secretStoreName: crossplane-secrets
  secretStoreKind: ClusterSecretStore
```

## Direct Keycloak SecretStore (Alternative)

If you want ESO to fetch secrets directly from Keycloak's API instead of going through Kubernetes Secrets, you can use ESO's webhook provider or a custom provider. However, the Kubernetes provider approach shown above is simpler and recommended for most use cases.

## Troubleshooting

- **ExternalSecret stuck in `SecretSyncedError`**: Check that the source Secret exists in `crossplane-system` and the ESO service account has RBAC permissions to read it
- **Empty client secret**: Ensure the Client `accessType` is `CONFIDENTIAL` — public clients do not have a client secret
- **Stale secret values**: Decrease `refreshInterval` on the ExternalSecret or manually trigger a refresh
- **Template errors**: Validate your template syntax by checking the ExternalSecret status conditions: `kubectl describe externalsecret my-app-oidc`
