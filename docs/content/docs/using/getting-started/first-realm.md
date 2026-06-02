---
sidebar_position: 3
title: First Realm
description: Create your first Keycloak realm with provider-keycloak
---

# Create Your First Realm

This guide walks you through creating a complete Keycloak realm with a client and user.

## Prerequisites

- [Provider installed](./installation.md)
- [Credentials configured](./configuration.md)

## Step 1: Create a Realm

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: my-app-realm
spec:
  forProvider:
    realm: "my-app"
    enabled: true
    displayName: "My Application"
  providerConfigRef:
    name: keycloak-provider-config
```

Apply and verify:

```bash
kubectl apply -f realm.yaml
kubectl get realm my-app-realm
```

Wait for the realm to become `READY`:

```bash
kubectl wait realm my-app-realm --for=condition=Ready --timeout=60s
```

## Step 2: Create a Client

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: my-web-app
spec:
  forProvider:
    clientId: my-web-app
    realmId: my-app
    accessType: public
    standardFlowEnabled: true
    validRedirectUris:
      - "http://localhost:3000/callback"
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 3: Create a User

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: demo-user
spec:
  forProvider:
    realmId: "my-app"
    username: "demo"
    email: "demo@example.com"
    firstName: "Demo"
    lastName: "User"
    enabled: true
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 4: Verify

```bash
# Check all resources are ready
kubectl get realm,client,user -l crossplane.io/provider=provider-keycloak
```

All resources should show `READY: True` and `SYNCED: True`.

## How Reconciliation Works

The provider continuously reconciles your desired state (the YAML manifests) with the actual state in Keycloak:

1. **Create**: When you apply a manifest, the provider creates the resource in Keycloak
2. **Update**: If you modify the manifest, the provider updates Keycloak accordingly
3. **Drift Detection**: If someone changes Keycloak directly (e.g., via the admin console), the provider detects the drift and corrects it
4. **Delete**: When you delete the Kubernetes resource, the provider removes it from Keycloak (unless `deletionPolicy: Orphan` is set)

## Next Steps

- Learn about [Clients](../resources/clients.md) for more advanced configurations
- Set up [Roles](../resources/roles.md) and [Groups](../resources/groups.md)
- Follow the [SSO with ArgoCD](../guides/sso-with-argocd.md) guide for a real-world example
