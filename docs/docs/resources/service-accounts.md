---
sidebar_position: 12
title: Service Accounts
description: Manage service account role assignments for clients
---

# Service Accounts

When a client has `serviceAccountsEnabled: true`, Keycloak creates a service account user for that client. You can assign realm-level and client-level roles to this service account.

> **Note:** The client must be confidential with service accounts enabled. See [Clients](./clients.md).

## API Reference

- **API Group**: `openidclient.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kinds**: `ClientServiceAccountRealmRole`, `ClientServiceAccountRole`

## ClientServiceAccountRealmRole

Assign a realm-level role to a client's service account.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientServiceAccountRealmRole
metadata:
  name: backend-admin-role
spec:
  forProvider:
    realmId: "my-realm"
    serviceAccountUserId: "service-account-user-uuid"
    role: "admin"
  providerConfigRef:
    name: keycloak-provider-config
```

### Using a Reference to the Client

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientServiceAccountRealmRole
metadata:
  name: backend-realm-role
spec:
  forProvider:
    realmId: "my-realm"
    serviceAccountUserClientIdRef:
      name: backend-service
    role: "realm-management"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientServiceAccountRole

Assign a client-level role to a client's service account. This is used when one client's service account needs a role defined in another client.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientServiceAccountRole
metadata:
  name: backend-client-role
spec:
  forProvider:
    realmId: "my-realm"
    serviceAccountUserId: "service-account-user-uuid"
    clientId: "target-client-uuid"
    role: "manage-users"
  providerConfigRef:
    name: keycloak-provider-config
```

### Using References

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientServiceAccountRole
metadata:
  name: backend-manage-role
spec:
  forProvider:
    realmId: "my-realm"
    serviceAccountUserClientIdRef:
      name: backend-service
    clientIdRef:
      name: realm-management-client
    role: "manage-users"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

### ClientServiceAccountRealmRole

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm the client and role belong to |
| `serviceAccountUserId` | string | UUID of the service account user |
| `serviceAccountUserClientIdRef` | ref | Reference to the client owning the service account |
| `role` | string | Realm role name to assign |

### ClientServiceAccountRole

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm the clients and role belong to |
| `serviceAccountUserId` | string | UUID of the service account user |
| `serviceAccountUserClientIdRef` | ref | Reference to the client owning the service account |
| `clientId` | string | UUID of the client providing the role |
| `role` | string | Client role name to assign |
