---
title: Resources
weight: 3
---

# Resources

These pages are curated, human-readable entry points for the provider-keycloak
managed resources. They focus on when to use each resource, common examples, and
the fields most users need first.

For exhaustive field schemas, default values, references, selectors, and status
fields, use the generated CRDs in `package/crds/`. Those CRDs are generated from
the provider APIs and are the source of truth when a resource page and schema
details differ.

## Documentation model

| Content type | Maintained as | Source of truth |
|--------------|---------------|-----------------|
| Resource overview pages | Curated docs | `docs/content/docs/using/resources/` |
| Complete field reference | Generated artifact | `package/crds/*.yaml` OpenAPI schemas |
| API group/version/kind | Generated artifact | CRD metadata in `package/crds/` |
| Operational walkthroughs | Authored guides | `docs/content/docs/using/guides/` and `examples/` |

## Recommended automation

The following sections should be generated or checked automatically from
`package/crds/` to prevent documentation drift:

- API reference blocks: API group, version, kind, and plural name.
- Field tables for `spec.forProvider`, references, selectors, and status.
- Links from resource pages to their matching CRD files.
- Shared Crossplane boilerplate such as `providerConfigRef`,
  `deletionPolicy`, `managementPolicies`, and secret references.

Keep the narrative examples and "when should I use this?" guidance authored by
humans; generate the exhaustive schema details from the CRDs.
