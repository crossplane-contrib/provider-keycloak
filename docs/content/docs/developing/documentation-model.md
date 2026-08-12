---
title: Documentation Model
weight: 2
---

Provider Keycloak has a large API surface. Keep the documentation useful by
separating authored guidance from generated reference material.

## Authored content

Write and review these pages by hand:

- Getting started pages that teach the first successful installation and realm.
- Scenario guides that combine several resources into a working integration.
- Troubleshooting pages that explain symptoms, causes, and fixes.
- AI usage pages that describe how to consume the docs.

Authored pages should answer "when and why should I use this?" and include
tested examples or links to manifests in `examples/`.

## Generated or schema-derived content

The following content should come from generated artifacts or be checked against
them:

- Complete `spec.forProvider` field lists.
- Reference and selector fields.
- Status fields and connection secret keys.
- API group, version, kind, plural name, and scope.
- Repeated Crossplane fields such as `providerConfigRef`, `deletionPolicy`, and
  `managementPolicies`.

The source of truth for this data is `package/crds/*.yaml`. When the APIs change,
regenerate the provider artifacts and update only the curated explanations that
need human context.

## Resource page checklist

Each resource page should include:

- A short explanation of what the resource manages.
- A small API reference block.
- One or more realistic examples.
- Links to related guides or examples.
- A pointer back to `package/crds/` for exhaustive schema details.

Avoid hand-maintaining complete field tables in Markdown unless they are
generated from the CRD OpenAPI schema.
