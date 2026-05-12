---
sidebar_position: 3
title: LDAP Integration
description: Integrate corporate LDAP/Active Directory with Keycloak
---

# LDAP Integration

This guide shows how to configure Keycloak to authenticate users against an LDAP directory (OpenLDAP or Active Directory).

## Overview

```
┌──────────┐     Auth     ┌──────────────┐    LDAP Bind    ┌──────────┐
│   User   │─────────────►│   Keycloak   │────────────────►│   LDAP   │
└──────────┘              └──────────────┘                  └──────────┘
                                 ▲
                                 │ Manages
                          ┌──────────────┐
                          │   Provider   │
                          │   Keycloak   │
                          └──────────────┘
```

## Step 1: Create the LDAP Password Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ldap-password
  namespace: crossplane-system
type: Opaque
stringData:
  password: "your-ldap-bind-password"
```

## Step 2: Configure User Federation

### OpenLDAP

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: UserFederation
metadata:
  name: openldap-federation
spec:
  forProvider:
    name: "Corporate LDAP"
    realmId: my-realm
    connectionUrl: "ldap://ldap.example.com:389"
    startTls: true
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
    searchScope: "SUBTREE"
    importEnabled: true
    batchSizeForSync: 100
    changedSyncPeriod: 604800
    trustEmail: true
  providerConfigRef:
    name: keycloak-provider-config
```

### Active Directory

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: UserFederation
metadata:
  name: ad-federation
spec:
  forProvider:
    name: "Active Directory"
    realmId: my-realm
    connectionUrl: "ldap://ad.corp.example.com:389"
    startTls: true
    bindDn: "cn=svc-keycloak,ou=services,dc=corp,dc=example,dc=com"
    bindCredentialSecretRef:
      key: "password"
      name: "ldap-password"
      namespace: "crossplane-system"
    editMode: "UNSYNCED"
    usersDn: "ou=users,dc=corp,dc=example,dc=com"
    usernameLdapAttribute: "sAMAccountName"
    rdnLdapAttribute: "cn"
    uuidLdapAttribute: "objectGUID"
    userObjectClasses:
      - "person"
      - "organizationalPerson"
      - "user"
    searchScope: "SUBTREE"
    importEnabled: true
    batchSizeForSync: 100
    changedSyncPeriod: 604800
    trustEmail: true
  providerConfigRef:
    name: keycloak-provider-config
```

## Step 3: Map LDAP Groups to Keycloak Groups

After user federation is configured, LDAP users can authenticate through Keycloak. To map LDAP groups, you can use Keycloak's built-in LDAP group mapper through the admin console or use the [function-keycloak-builtin-objects](https://gitlab.com/corewire/images/crossplane/function-keycloak-builtin-objects) composition function.

## Edit Mode Reference

| Mode | Keycloak → LDAP | LDAP → Keycloak | Use Case |
|------|:---:|:---:|----------|
| `READ_ONLY` | ✗ | ✓ | LDAP is source of truth, no write-back |
| `WRITABLE` | ✓ | ✓ | Bidirectional sync |
| `UNSYNCED` | ✗ | ✓ (import only) | Import users, store changes locally |

## Troubleshooting

- **Connection refused**: Verify the LDAP URL and ensure network connectivity from the Keycloak pod
- **Invalid credentials**: Check the bind DN and password in the secret
- **No users found**: Verify `usersDn` and `userObjectClasses` match your LDAP schema
- **StartTLS failures**: Ensure the LDAP server supports StartTLS and certificates are valid
