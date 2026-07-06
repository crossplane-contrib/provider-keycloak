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

{{< cards >}}
  {{< card link="realms/" title="Realms" icon="template" subtitle="Realm · realm.keycloak.crossplane.io" >}}
  {{< card link="realm-settings/" title="Realm Settings" icon="adjustments" subtitle="RealmEvents · RequiredAction · UserProfile · Keystores · Client Policies" >}}
  {{< card link="clients/" title="Clients (OIDC)" icon="shield-check" subtitle="Client · openidclient.keycloak.crossplane.io" >}}
  {{< card link="saml-clients/" title="SAML Clients" icon="shield-check" subtitle="Client · ClientScope · samlclient.keycloak.crossplane.io" >}}
  {{< card link="openid-client-scopes/" title="Client Scopes" icon="tag" subtitle="ClientScope · ClientDefaultScopes · ClientOptionalScopes" >}}
  {{< card link="client-authorization/" title="Client Authorization" icon="lock-closed" subtitle="Resources · Permissions · Policies" >}}
  {{< card link="service-accounts/" title="Service Accounts" icon="user-circle" subtitle="ServiceAccountRealmRole · ServiceAccountRole" >}}
  {{< card link="users/" title="Users" icon="users" subtitle="User · Groups · Roles · Permissions · user.keycloak.crossplane.io" >}}
  {{< card link="roles/" title="Roles" icon="badge-check" subtitle="Role · role.keycloak.crossplane.io" >}}
  {{< card link="groups/" title="Groups" icon="user-group" subtitle="Group · Memberships · Roles · Permissions" >}}
  {{< card link="protocol-mappers/" title="Protocol Mappers" icon="paper-clip" subtitle="ProtocolMapper · RoleMapper · GroupMembershipProtocolMapper" >}}
  {{< card link="identity-providers/" title="Identity Providers" icon="switch-horizontal" subtitle="OIDC · SAML · Google · Kubernetes · OpenShift · SPIFFE" >}}
  {{< card link="user-federation/" title="User Federation" icon="server" subtitle="LDAP/AD federation and all mapper types" >}}
  {{< card link="authentication-flows/" title="Authentication Flows" icon="arrows-expand" subtitle="Flow · Subflow · Execution · ExecutionConfig · Bindings" >}}
  {{< card link="default-config/" title="Default Config" icon="star" subtitle="DefaultGroups · DefaultRoles" >}}
  {{< card link="organizations/" title="Organizations" icon="office-building" subtitle="Organization · Keycloak 26.6+ multi-tenancy" >}}
  {{< card link="workflows/" title="Workflows" icon="chip" subtitle="Workflow · event-driven automation (Keycloak 26.5+)" >}}
{{< /cards >}}

