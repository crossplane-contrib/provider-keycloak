---
title: Common Patterns
weight: 2
---

Most provider-keycloak resources follow the same Crossplane manifest structure.
Use these patterns across guides and examples instead of repeating long
explanations on every resource page.

## ProviderConfig reference

Every managed resource should point at the Keycloak credentials it should use:

```yaml
spec:
  providerConfigRef:
    name: keycloak-provider-config
```

See [ProviderConfig](/docs/using/reference/provider-config/) for credential
setup.

## Deletion policy

Use `deletionPolicy` to control what happens to the external Keycloak object when
the Kubernetes resource is deleted:

```yaml
spec:
  deletionPolicy: Delete
```

- `Delete` removes the external Keycloak object.
- `Orphan` leaves the external Keycloak object in place.

Use `Orphan` for shared or manually managed Keycloak objects that should survive
GitOps cleanup.

## Import existing resources

To adopt objects that already exist in Keycloak, create the managed resource
with the same identifying fields and start with non-destructive policies:

```yaml
spec:
  managementPolicies: [Observe, Update]
  deletionPolicy: Orphan
```

Use `Observe` first if you want a read-only dry run before allowing updates.

For many resources (for example `Realm`, `Client`, `Group`, and `User`), the
provider can resolve the existing object from identifying fields in
`spec.forProvider`, so you usually do not need to set an external name
annotation manually.

Some resources use provider-generated IDs for import. For those resources, set
the external identifier explicitly:

```yaml
metadata:
  annotations:
    crossplane.io/external-name: "<provider-id>"
```

## References and selectors

Many resources can use direct IDs, Crossplane references, or selectors. Prefer
references when another managed resource owns the target object:

```yaml
spec:
  forProvider:
    realmIdRef:
      name: example-realm
```

Use direct IDs when the target object is managed outside Crossplane:

```yaml
spec:
  forProvider:
    realmId: example
```

## Secret references

Store credentials in Kubernetes Secrets and reference them from provider
resources:

```yaml
spec:
  forProvider:
    clientSecretSecretRef:
      namespace: crossplane-system
      name: oidc-client
      key: client-secret
```

See [Credentials](/docs/using/reference/credentials/) for secret formats.

## Complete schemas

Resource pages show common fields and examples. The generated CRDs in
`package/crds/` contain the complete OpenAPI schema for every field, including
references, selectors, status, and connection details.
