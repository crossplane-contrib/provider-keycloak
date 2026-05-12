---
title: Provider Keycloak Documentation
description: Crossplane provider for declarative Keycloak management
hide_table_of_contents: true
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
| 📦 [Installation](./docs/getting-started/installation) | Install the provider into your Crossplane cluster |
| ⚙️ [Configuration](./docs/getting-started/configuration) | Connect to your Keycloak instance |
| 🚀 [First Realm](./docs/getting-started/first-realm) | Create your first realm, client, and user |
| 📖 [Resources](./docs/resources/realms) | Reference for all managed resource types |
| 🗺️ [Guides](./docs/guides/sso-with-argocd) | Real-world walkthroughs (ArgoCD SSO, K8s OIDC, LDAP, ESO) |

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
