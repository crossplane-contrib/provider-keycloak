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
| [Realm](/docs/using/resources/realms/) | `realm.keycloak.crossplane.io` | Keycloak realms and realm settings |
| [Client](/docs/using/resources/clients/) | `openidclient.keycloak.crossplane.io` | OpenID Connect clients |
| [SAML Client](/docs/using/resources/saml-clients/) | `samlclient.keycloak.crossplane.io` | SAML 2.0 clients |
| [User](/docs/using/resources/users/) | `user.keycloak.crossplane.io` | Users, user roles, user groups, permissions |
| [Role](/docs/using/resources/roles/) | `role.keycloak.crossplane.io` | Realm and client roles |
| [Group](/docs/using/resources/groups/) | `group.keycloak.crossplane.io` | User groups, memberships, permissions |
| [Protocol Mapper](/docs/using/resources/protocol-mappers/) | `client.keycloak.crossplane.io` | Token/assertion mappers |
| [Identity Provider](/docs/using/resources/identity-providers/) | `oidc.keycloak.crossplane.io`, `saml.keycloak.crossplane.io`, `identityprovider.keycloak.crossplane.io` | OIDC, SAML, Google, Kubernetes, OpenShift, SPIFFE identity providers |
| [User Federation](/docs/using/resources/user-federation/) | `ldap.keycloak.crossplane.io` | LDAP/AD federation and all mapper types |
| [Client Scopes](/docs/using/resources/openid-client-scopes/) | `openidclient.keycloak.crossplane.io` | OpenID client scopes (default and optional) |
| [Client Authorization](/docs/using/resources/client-authorization/) | `openidclient.keycloak.crossplane.io` | Fine-grained authorization (resources, permissions, policies) |
| [Service Accounts](/docs/using/resources/service-accounts/) | `openidclient.keycloak.crossplane.io` | Service account role assignments |
| [Authentication Flows](/docs/using/resources/authentication-flows/) | `authenticationflow.keycloak.crossplane.io` | Custom authentication flows, executions, bindings |
| [Default Config](/docs/using/resources/default-config/) | `defaults.keycloak.crossplane.io` | Default groups and roles for new users |
| [Organization](/docs/using/resources/organizations/) | `organization.keycloak.crossplane.io` | Multi-tenancy organizations (Keycloak 26.6+) |
| [Workflow](/docs/using/resources/workflows/) | `workflow.keycloak.crossplane.io` | Event-driven automation workflows (Keycloak 26.5+) |
| [Realm Settings](/docs/using/resources/realm-settings/) | `realm.keycloak.crossplane.io` | Events, required actions, user profiles, keystores, client policies |
| [Group Membership Mapper](/docs/using/resources/protocol-mappers/) | `openidgroup.keycloak.crossplane.io` | Group membership protocol mapper |

All CRDs are documented with working examples. For the complete OpenAPI schema of each resource, see the generated CRDs in [`package/crds/`](https://github.com/crossplane-contrib/provider-keycloak/tree/main/package/crds).

## Documentation Structure

- [Using](/docs/using/) — install, configure, and operate provider-keycloak
- [Developing](/docs/developing/) — docs and contributor-focused setup
- [AI Usage](/docs/ai-usage/) — AI-oriented files and entry points
