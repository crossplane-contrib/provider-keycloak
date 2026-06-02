---
sidebar_position: 11
title: Client Authorization
description: Manage fine-grained authorization resources, permissions, and policies
---

# Client Authorization

Keycloak provides fine-grained authorization services for clients. This includes defining resources, creating policies based on various criteria, and assigning permissions that tie resources to policies.

> **Note:** Authorization requires a confidential client with authorization enabled. See [Clients](./clients.md) for setting up the resource server.

## API Reference

- **API Group**: `openidclient.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kinds**: `ClientAuthorizationResource`, `ClientAuthorizationPermission`, `ClientClientPolicy`, `ClientGroupPolicy`, `ClientRolePolicy`, `ClientUserPolicy`, `ClientPermissions`

## ClientAuthorizationResource

Define a protected resource on a resource server.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientAuthorizationResource
metadata:
  name: api-documents
spec:
  forProvider:
    name: "documents"
    displayName: "Documents API"
    realmId: "my-realm"
    resourceServerId: "client-uuid"
    ownerManagedAccess: false
    iconUri: "https://example.com/icons/documents.png"
    attributes:
      category: '["api"]'
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientAuthorizationPermission

A permission links resources or resource types to authorization policies.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientAuthorizationPermission
metadata:
  name: view-documents-permission
spec:
  forProvider:
    name: "view-documents"
    description: "Permission to view documents"
    realmId: "my-realm"
    resourceServerId: "client-uuid"
    decisionStrategy: "UNANIMOUS"
    policies:
      - "policy-id-1"
      - "policy-id-2"
    resourceType: "documents"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientClientPolicy

A policy that grants access based on which client is making the request.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientClientPolicy
metadata:
  name: trusted-clients-policy
spec:
  forProvider:
    name: "trusted-clients"
    description: "Allow access from trusted clients"
    realmId: "my-realm"
    resourceServerId: "client-uuid"
    decisionStrategy: "UNANIMOUS"
    logic: "POSITIVE"
    clients:
      - "trusted-client-uuid-1"
      - "trusted-client-uuid-2"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientGroupPolicy

A policy that grants access based on group membership.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientGroupPolicy
metadata:
  name: admin-group-policy
spec:
  forProvider:
    name: "admin-group-access"
    description: "Allow access for admin group members"
    realmId: "my-realm"
    resourceServerId: "client-uuid"
    decisionStrategy: "UNANIMOUS"
    logic: "POSITIVE"
    groupsClaim: "groups"
    groups:
      - id: "group-uuid"
        extendChildren: true
        path: "/admins"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientRolePolicy

A policy that grants access based on assigned roles.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientRolePolicy
metadata:
  name: manager-role-policy
spec:
  forProvider:
    name: "manager-role-access"
    description: "Allow access for users with manager role"
    realmId: "my-realm"
    resourceServerId: "client-uuid"
    type: "role"
    decisionStrategy: "UNANIMOUS"
    logic: "POSITIVE"
    role:
      - id: "role-uuid"
        required: true
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientUserPolicy

A policy that grants access to specific users.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientUserPolicy
metadata:
  name: specific-users-policy
spec:
  forProvider:
    name: "specific-users"
    description: "Allow access for specific users"
    realmId: "my-realm"
    resourceServerId: "client-uuid"
    decisionStrategy: "AFFIRMATIVE"
    logic: "POSITIVE"
    users:
      - "user-uuid-1"
      - "user-uuid-2"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientPermissions

Enable and configure fine-grained permissions on the client itself (e.g., who can manage the client, exchange tokens, etc.).

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientPermissions
metadata:
  name: backend-client-permissions
spec:
  forProvider:
    clientId: "client-uuid"
    realmId: "my-realm"
    viewScope:
      - policies:
          - "admin-group-policy-id"
        description: "View client"
        decisionStrategy: "UNANIMOUS"
    manageScope:
      - policies:
          - "admin-group-policy-id"
        description: "Manage client"
        decisionStrategy: "UNANIMOUS"
    configureScope:
      - policies:
          - "admin-group-policy-id"
        description: "Configure client"
        decisionStrategy: "UNANIMOUS"
    tokenExchangeScope:
      - policies:
          - "trusted-clients-policy-id"
        description: "Token exchange"
        decisionStrategy: "UNANIMOUS"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

### Policy Common Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Policy name |
| `description` | string | Policy description |
| `realmId` | string | Realm this policy belongs to |
| `resourceServerId` | string | UUID of the resource server (client) |
| `decisionStrategy` | string | `UNANIMOUS`, `AFFIRMATIVE`, or `CONSENSUS` |
| `logic` | string | `POSITIVE` or `NEGATIVE` (default `POSITIVE`) |

### ClientAuthorizationPermission

| Field | Type | Description |
|-------|------|-------------|
| `policies` | []string | List of policy IDs to evaluate |
| `resourceType` | string | Resource type this permission applies to |

### ClientAuthorizationResource

| Field | Type | Description |
|-------|------|-------------|
| `displayName` | string | Display name for the resource |
| `ownerManagedAccess` | bool | Enable user-managed access (default `false`) |
| `iconUri` | string | Icon URI for the resource |
| `attributes` | map | Resource attributes |
