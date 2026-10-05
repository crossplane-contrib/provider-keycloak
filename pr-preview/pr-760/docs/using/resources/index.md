# Resources

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


Complete reference for all provider-keycloak managed resources. Every CRD is
documented with working examples taken from the project&#39;s end-to-end tests,
links to the underlying Terraform resource, and guidance on when to use each
resource.

For exhaustive field schemas, default values, references, selectors, and status
fields, see the generated CRDs in
[`package/crds/`](https://github.com/crossplane-contrib/provider-keycloak/tree/main/package/crds)
or browse all CRDs interactively on the
[Upbound Marketplace CRD Explorer](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/crds).

{{&lt; cards &gt;}}
  {{&lt; card link=&#34;realms/&#34; title=&#34;Realms&#34; icon=&#34;template&#34; subtitle=&#34;Realm · realm.keycloak.crossplane.io&#34; &gt;}}
  {{&lt; card link=&#34;realm-settings/&#34; title=&#34;Realm Settings&#34; icon=&#34;adjustments&#34; subtitle=&#34;RealmEvents · RequiredAction · UserProfile · Keystores · Client Policies&#34; &gt;}}
  {{&lt; card link=&#34;clients/&#34; title=&#34;Clients (OIDC)&#34; icon=&#34;shield-check&#34; subtitle=&#34;Client · openidclient.keycloak.crossplane.io&#34; &gt;}}
  {{&lt; card link=&#34;saml-clients/&#34; title=&#34;SAML Clients&#34; icon=&#34;shield-check&#34; subtitle=&#34;Client · ClientScope · samlclient.keycloak.crossplane.io&#34; &gt;}}
  {{&lt; card link=&#34;openid-client-scopes/&#34; title=&#34;Client Scopes&#34; icon=&#34;tag&#34; subtitle=&#34;ClientScope · ClientDefaultScopes · ClientOptionalScopes&#34; &gt;}}
  {{&lt; card link=&#34;client-authorization/&#34; title=&#34;Client Authorization&#34; icon=&#34;lock-closed&#34; subtitle=&#34;Resources · Permissions · Policies&#34; &gt;}}
  {{&lt; card link=&#34;service-accounts/&#34; title=&#34;Service Accounts&#34; icon=&#34;user-circle&#34; subtitle=&#34;ServiceAccountRealmRole · ServiceAccountRole&#34; &gt;}}
  {{&lt; card link=&#34;users/&#34; title=&#34;Users&#34; icon=&#34;users&#34; subtitle=&#34;User · Groups · Roles · Permissions · user.keycloak.crossplane.io&#34; &gt;}}
  {{&lt; card link=&#34;roles/&#34; title=&#34;Roles&#34; icon=&#34;badge-check&#34; subtitle=&#34;Role · role.keycloak.crossplane.io&#34; &gt;}}
  {{&lt; card link=&#34;groups/&#34; title=&#34;Groups&#34; icon=&#34;user-group&#34; subtitle=&#34;Group · Memberships · Roles · Permissions&#34; &gt;}}
  {{&lt; card link=&#34;protocol-mappers/&#34; title=&#34;Protocol Mappers&#34; icon=&#34;paper-clip&#34; subtitle=&#34;ProtocolMapper · RoleMapper · GroupMembershipProtocolMapper&#34; &gt;}}
  {{&lt; card link=&#34;identity-providers/&#34; title=&#34;Identity Providers&#34; icon=&#34;switch-horizontal&#34; subtitle=&#34;OIDC · SAML · Google · Kubernetes · OpenShift · SPIFFE&#34; &gt;}}
  {{&lt; card link=&#34;user-federation/&#34; title=&#34;User Federation&#34; icon=&#34;server&#34; subtitle=&#34;LDAP/AD federation and all mapper types&#34; &gt;}}
  {{&lt; card link=&#34;authentication-flows/&#34; title=&#34;Authentication Flows&#34; icon=&#34;arrows-expand&#34; subtitle=&#34;Flow · Subflow · Execution · ExecutionConfig · Bindings&#34; &gt;}}
  {{&lt; card link=&#34;default-config/&#34; title=&#34;Default Config&#34; icon=&#34;star&#34; subtitle=&#34;DefaultGroups · DefaultRoles&#34; &gt;}}
  {{&lt; card link=&#34;organizations/&#34; title=&#34;Organizations&#34; icon=&#34;office-building&#34; subtitle=&#34;Organization · Keycloak 26.6&#43; multi-tenancy&#34; &gt;}}
  {{&lt; card link=&#34;workflows/&#34; title=&#34;Workflows&#34; icon=&#34;chip&#34; subtitle=&#34;Workflow · event-driven automation (Keycloak 26.5&#43;)&#34; &gt;}}
{{&lt; /cards &gt;}}


