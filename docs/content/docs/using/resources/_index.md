---
title: Resources
weight: 3
---

# Resources

Complete reference for all provider-keycloak managed resources. Every CRD is
documented with working examples taken from the project's end-to-end tests,
links to the underlying Terraform resource, and guidance on when to use each
resource.

For exhaustive field schemas, default values, references, selectors, and status
fields, see the generated CRDs in
[`package/crds/`](https://github.com/crossplane-contrib/provider-keycloak/tree/main/package/crds)
or browse all CRDs interactively on the
[Upbound Marketplace CRD Explorer](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/crds).

## Resource Pages

| Page | API Groups | Resources |
|------|------------|-----------|
| [Realms](./realms/) | `realm.keycloak.crossplane.io` | Realm |
| [Realm Settings](./realm-settings/) | `realm.keycloak.crossplane.io` | RealmEvents, RequiredAction, UserProfile, KeystoreRsa, DefaultClientScopes, OptionalClientScopes, ClientPolicyProfile, ClientPolicyProfilePolicy |
| [Clients](./clients/) | `openidclient.keycloak.crossplane.io` | Client |
| [OpenID Client Scopes](./openid-client-scopes/) | `openidclient.keycloak.crossplane.io` | ClientScope, ClientDefaultScopes, ClientOptionalScopes |
| [Client Authorization](./client-authorization/) | `openidclient.keycloak.crossplane.io` | ClientAuthorizationResource, ClientAuthorizationPermission, ClientClientPolicy, ClientGroupPolicy, ClientRolePolicy, ClientUserPolicy, ClientRegexPolicy, ClientPermissions |
| [Service Accounts](./service-accounts/) | `openidclient.keycloak.crossplane.io` | ClientServiceAccountRealmRole, ClientServiceAccountRole |
| [SAML Clients](./saml-clients/) | `samlclient.keycloak.crossplane.io` | Client, ClientScope, ClientDefaultScopes |
| [Users](./users/) | `user.keycloak.crossplane.io` | User, Groups, Roles, Permissions, UserFederation |
| [Roles](./roles/) | `role.keycloak.crossplane.io` | Role |
| [Groups](./groups/) | `group.keycloak.crossplane.io` | Group, Memberships, Roles, Permissions |
| [Protocol Mappers](./protocol-mappers/) | `client.keycloak.crossplane.io`, `openidgroup.keycloak.crossplane.io` | ProtocolMapper, RoleMapper, GroupMembershipProtocolMapper |
| [Identity Providers](./identity-providers/) | `oidc.keycloak.crossplane.io`, `saml.keycloak.crossplane.io`, `identityprovider.keycloak.crossplane.io` | IdentityProvider (OIDC), GoogleIdentityProvider, IdentityProvider (SAML), IdentityProviderMapper, KubernetesIdentityProvider, OidcOpenShiftV4IdentityProvider, SpiffeIdentityProvider, ProviderTokenExchangeScopePermission |
| [User Federation](./user-federation/) | `ldap.keycloak.crossplane.io`, `user.keycloak.crossplane.io` | UserFederation, UserAttributeMapper, FullNameMapper, GroupMapper, RoleMapper, HardcodedAttributeMapper, HardcodedGroupMapper, HardcodedRoleMapper, MsadUserAccountControlMapper, MsadLdsUserAccountControlMapper, CustomMapper, UserFederation (custom) |
| [Authentication Flows](./authentication-flows/) | `authenticationflow.keycloak.crossplane.io` | Flow, Subflow, Execution, ExecutionConfig, Bindings |
| [Default Config](./default-config/) | `defaults.keycloak.crossplane.io` | DefaultGroups, Roles |
| [Organizations](./organizations/) | `organization.keycloak.crossplane.io` | Organization |
| [Workflows](./workflows/) | `workflow.keycloak.crossplane.io` | Workflow |
