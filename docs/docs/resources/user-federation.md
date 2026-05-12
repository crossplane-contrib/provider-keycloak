---
sidebar_position: 8
title: User Federation
description: Integrate LDAP/Active Directory with Keycloak
---

# User Federation

User federation allows Keycloak to use external user stores such as LDAP or Active Directory.

## API Reference

- **API Group**: `ldap.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kind**: `UserFederation`

## OpenLDAP Integration

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: UserFederation
metadata:
  name: openldap
spec:
  forProvider:
    name: "openldap"
    realmId: my-realm
    connectionUrl: "ldap://ldap.example.com:389"
    startTls: false
    bindDn: "cn=admin,dc=example,dc=com"
    bindCredentialSecretRef:
      key: "password"
      name: "ldap-password"
      namespace: "crossplane-system"
    editMode: "UNSYNCED"
    usersDn: "ou=users,dc=example,dc=com"
    usernameLdapAttribute: "uid"
    rdnLdapAttribute: "uid"
    uuidLdapAttribute: "entryUUID"
    userObjectClasses:
      - "inetOrgPerson"
      - "shadowAccount"
    searchScope: "SUBTREE"
    importEnabled: true
    batchSizeForSync: 100
    changedSyncPeriod: 604800
    trustEmail: true
  providerConfigRef:
    name: keycloak-provider-config
```

## Active Directory Integration

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: UserFederation
metadata:
  name: active-directory
spec:
  forProvider:
    name: "active-directory"
    realmId: my-realm
    connectionUrl: "ldap://ad.corp.example.com:389"
    startTls: true
    bindDn: "cn=svc-keycloak,ou=services,dc=corp,dc=example,dc=com"
    bindCredentialSecretRef:
      key: "password"
      name: "ad-password"
      namespace: "crossplane-system"
    editMode: "UNSYNCED"
    usersDn: "ou=users,dc=corp,dc=example,dc=com"
    usernameLdapAttribute: "sAMAccountName"
    rdnLdapAttribute: "cn"
    uuidLdapAttribute: "sAMAccountName"
    userObjectClasses:
      - "person"
      - "organizationalPerson"
      - "user"
    searchScope: "SUBTREE"
    importEnabled: true
    batchSizeForSync: 100
    changedSyncPeriod: 604800
    trustEmail: true
    validatePasswordPolicy: false
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name for the federation |
| `realmId` | string | Realm to federate |
| `connectionUrl` | string | LDAP server URL |
| `bindDn` | string | LDAP bind DN for authentication |
| `bindCredentialSecretRef` | object | Reference to bind password secret |
| `usersDn` | string | DN of the user tree |
| `usernameLdapAttribute` | string | LDAP attribute for username |
| `userObjectClasses` | []string | LDAP object classes to search |
| `editMode` | string | `READ_ONLY`, `WRITABLE`, or `UNSYNCED` |
| `importEnabled` | bool | Whether to import users on demand |
| `searchScope` | string | `ONE_LEVEL` or `SUBTREE` |
| `startTls` | bool | Use StartTLS for connection |

## Edit Modes

| Mode | Description |
|------|-------------|
| `READ_ONLY` | Users are read from LDAP, changes in Keycloak are not synced back |
| `WRITABLE` | Changes in Keycloak are synced back to LDAP |
| `UNSYNCED` | Users are imported but Keycloak changes are stored locally only |
