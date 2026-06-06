---
title: Guides
weight: 2
---

# Guides

Guides are authored, scenario-based walkthroughs. They should explain why the
resources are combined in a specific way and link to runnable manifests in
`examples/` whenever possible.

Use guides for end-to-end workflows such as:

- SSO integrations like ArgoCD.
- Kubernetes API server OIDC configuration.
- LDAP/Active Directory federation.
- External Secrets Operator templating.
- Local kind-based demos with Keycloak, Traefik, and sample applications.

Do not duplicate full CRD schemas in guides. Link to the relevant resource pages
and keep generated field details in `package/crds/`.
