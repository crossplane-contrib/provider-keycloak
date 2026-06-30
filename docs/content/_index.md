---
title: Provider Keycloak Documentation
description: Crossplane provider for declarative Keycloak management
---

# Provider Keycloak

A [Crossplane](https://crossplane.io/) provider for declaratively managing [Keycloak](https://www.keycloak.org/) resources using Kubernetes custom resources.

Built with [Upjet](https://github.com/crossplane/upjet) on top of the [Keycloak Terraform Provider](https://github.com/keycloak/terraform-provider-keycloak).

---

## What It Does

- **Declarative Management** — Define Keycloak realms, clients, users, roles, and more as Kubernetes YAML
- **Continuous Reconciliation** — The provider detects and corrects drift between your desired state and Keycloak
- **GitOps Ready** — Store your entire Keycloak configuration in Git and apply it with your existing CD pipeline

---

## Quick Links

| | |
|---|---|
| 📦 [Installation](/docs/using/getting-started/installation/) | Install the provider into your Crossplane cluster |
| ⚙️ [Configuration](/docs/using/getting-started/configuration/) | Connect to your Keycloak instance |
| 🚀 [First Realm](/docs/using/getting-started/first-realm/) | Create your first realm, client, and user |
| 📖 [Resources](/docs/using/resources/realms/) | Reference for all managed resource types |
| 🗺️ [Guides](/docs/using/guides/sso-with-argocd/) | Real-world walkthroughs (ArgoCD SSO, K8s OIDC, LDAP, ESO, end-to-end kind) |

---

## Managed Resources

| Resource | API Group | Description |
|----------|-----------|-------------|
| Realm | `realm.keycloak.crossplane.io` | Keycloak realms |
| Client | `openidclient.keycloak.crossplane.io` | OIDC clients |
| User | `user.keycloak.crossplane.io` | Users |
| Role | `role.keycloak.crossplane.io` | Realm and client roles |
| Group | `group.keycloak.crossplane.io` | User groups |
| ProtocolMapper | `client.keycloak.crossplane.io` | Token/assertion mappers |
| IdentityProvider | `oidc.keycloak.crossplane.io` | OIDC identity providers |
| UserFederation | `ldap.keycloak.crossplane.io` | LDAP/AD federation |

See the full list at the [Upbound Marketplace](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak).

## Documentation Structure

- [Using](/docs/using/) — install, configure, and operate provider-keycloak
- [Developing](/docs/developing/) — docs and contributor-focused setup
- [AI Usage](/docs/ai-usage/) — AI-oriented files and entry points
