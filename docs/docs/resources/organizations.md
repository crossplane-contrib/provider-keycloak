---
sidebar_position: 13
title: Organizations
description: Manage Keycloak organizations for multi-tenancy
---

# Organizations

Organizations provide multi-tenancy support in Keycloak, allowing you to group users under organizational entities with their own domains and attributes.

## API Reference

- **API Group**: `organization.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kind**: `Organization`

## Basic Organization

```yaml
apiVersion: organization.keycloak.crossplane.io/v1alpha1
kind: Organization
metadata:
  name: acme-corp
spec:
  forProvider:
    name: "Acme Corporation"
    alias: "acme-corp"
    realm: "my-realm"
    enabled: true
  providerConfigRef:
    name: keycloak-provider-config
```

## Organization with Domains and Attributes

```yaml
apiVersion: organization.keycloak.crossplane.io/v1alpha1
kind: Organization
metadata:
  name: partner-org
spec:
  forProvider:
    name: "Partner Organization"
    alias: "partner-org"
    description: "External partner organization"
    realm: "my-realm"
    enabled: true
    domain:
      - name: "partner.com"
        verified: true
      - name: "partner.org"
        verified: false
    attributes:
      tier: "premium"
      region: "us-east"
    redirectUrl: "https://partner.com/welcome"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Organization display name |
| `alias` | string | Unique alias identifier |
| `description` | string | Description of the organization |
| `realm` | string | Realm this organization belongs to |
| `enabled` | bool | Whether the organization is active |
| `domain` | []object | List of domains associated with the organization |
| `attributes` | map | Custom key-value attributes |
| `redirectUrl` | string | Landing page URL after registration or invitation |
