# User Federation

Use user federation when Keycloak should authenticate users against an external directory or user store instead of managing every account in its own database. LDAP and Active Directory federation let users sign in with their existing directory credentials, while mapper resources control how LDAP attributes, groups, and roles are projected into Keycloak. Use the custom user federation CRD when you have a Keycloak user storage SPI provider that is not LDAP-based.

## API Reference

| Kind | API Group | Terraform | CRD Explorer |
|------|-----------|-----------|---|
| `UserFederation` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_user_federation`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_user_federation) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/UserFederation/v1alpha1) |
| `UserAttributeMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_user_attribute_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_user_attribute_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/UserAttributeMapper/v1alpha1) |
| `FullNameMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_full_name_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_full_name_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/FullNameMapper/v1alpha1) |
| `GroupMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_group_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_group_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/GroupMapper/v1alpha1) |
| `RoleMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_role_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_role_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/RoleMapper/v1alpha1) |
| `HardcodedAttributeMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_hardcoded_attribute_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_hardcoded_attribute_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/HardcodedAttributeMapper/v1alpha1) |
| `HardcodedGroupMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_hardcoded_group_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_hardcoded_group_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/HardcodedGroupMapper/v1alpha1) |
| `HardcodedRoleMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_hardcoded_role_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_hardcoded_role_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/HardcodedRoleMapper/v1alpha1) |
| `MsadUserAccountControlMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_msad_user_account_control_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_msad_user_account_control_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/MsadUserAccountControlMapper/v1alpha1) |
| `MsadLdsUserAccountControlMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_msad_lds_user_account_control_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_msad_lds_user_account_control_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/MsadLdsUserAccountControlMapper/v1alpha1) |
| `CustomMapper` | `ldap.keycloak.crossplane.io/v1alpha1` | [`keycloak_ldap_custom_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/ldap_custom_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/ldap.keycloak.crossplane.io/CustomMapper/v1alpha1) |
| `UserFederation` | `user.keycloak.crossplane.io/v1alpha1` | [`keycloak_custom_user_federation`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/custom_user_federation) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/user.keycloak.crossplane.io/UserFederation/v1alpha1) |

## Working YAML examples

### LDAP UserFederation

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: UserFederation
metadata:
  name: ldap-user-federation
spec:
  deletionPolicy: Delete
  forProvider:
    bindCredentialSecretRef:
      key: secret
      name: bind-credential-secret
      namespace: dev
    bindDn: cn=admin,dc=example,dc=org
    connectionTimeout: 5s
    connectionUrl: ldap://openldap
    enabled: false
    kerberos:
      - kerberosRealm: FOO.LOCAL
        keyTab: /etc/host.keytab
        serverPrincipal: HTTP/host.foo.com@FOO.LOCAL
    name: openldap
    rdnLdapAttribute: cn
    readTimeout: 10s
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    userObjectClasses:
      - simpleSecurityObject
      - organizationalRole
    usernameLdapAttribute: cn
    usersDn: dc=example,dc=org
    uuidLdapAttribute: entryDN
    deleteDefaultMappers: false
  providerConfigRef:
    name: "keycloak-provider-config"
```

### UserAttributeMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: UserAttributeMapper
metadata:
  name: ldap-user-attribute-mapper
spec:
  deletionPolicy: Delete
  forProvider:
    ldapAttribute: bar
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    name: user-attribute-mapper
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    userModelAttribute: foo
  providerConfigRef:
    name: "keycloak-provider-config"
```

### FullNameMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: FullNameMapper
metadata:
  name: ldap-full-name-mapper
spec:
  deletionPolicy: Delete
  forProvider:
    ldapFullNameAttribute: cn
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    name: full-name-mapper
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### GroupMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: GroupMapper
metadata:
  name: ldap-group-mapper
spec:
  deletionPolicy: Delete
  forProvider:
    groupNameLdapAttribute: cn
    groupObjectClasses:
      - groupOfNames
    ldapGroupsDn: dc=example,dc=org
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    memberofLdapAttribute: memberOf
    membershipAttributeType: DN
    membershipLdapAttribute: member
    membershipUserLdapAttribute: cn
    name: group-mapper
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### RoleMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: RoleMapper
metadata:
  name: ldap-role-mapper
spec:
  deletionPolicy: Delete
  forProvider:
    ldapRolesDn: dc=example,dc=org
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    memberofLdapAttribute: memberOf
    membershipAttributeType: DN
    membershipLdapAttribute: member
    membershipUserLdapAttribute: cn
    name: role-mapper
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    roleNameLdapAttribute: cn
    roleObjectClasses:
      - groupOfNames
    userRolesRetrieveStrategy: GET_ROLES_FROM_USER_MEMBEROF_ATTRIBUTE
  providerConfigRef:
    name: "keycloak-provider-config"
```

### HardcodedRoleMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: HardcodedRoleMapper
metadata:
  name: assign-test-role-to-all-users
spec:
  deletionPolicy: Delete
  forProvider:
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    name: assign-test-role-to-all-users
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    roleRef:
      name: test
  providerConfigRef:
    name: "keycloak-provider-config"
```

### HardcodedGroupMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: HardcodedGroupMapper
metadata:
  name: assign-group-to-users
spec:
  deletionPolicy: Delete
  forProvider:
    groupRef:
      name: test
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    name: assign-group-to-users
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### HardcodedAttributeMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: HardcodedAttributeMapper
metadata:
  name: assign-bar-to-foo
spec:
  deletionPolicy: Delete
  forProvider:
    attributeName: foo
    attributeValue: bar
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    name: assign-foo-to-bar
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### MsadUserAccountControlMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: MsadUserAccountControlMapper
metadata:
  name: msad-user-account-control-mapper
spec:
  deletionPolicy: Delete
  forProvider:
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    name: msad-user-account-control-mapper
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### MsadLdsUserAccountControlMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: MsadLdsUserAccountControlMapper
metadata:
  name: msad-lds-user-account-control-mapper
spec:
  deletionPolicy: Delete
  forProvider:
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    name: msad-lds-user-account-control-mapper
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### CustomMapper

```yaml
apiVersion: ldap.keycloak.crossplane.io/v1alpha1
kind: CustomMapper
metadata:
  name: custom-mapper
spec:
  deletionPolicy: Delete
  forProvider:
    config:
      ldap.full.name.attribute: cn
    ldapUserFederationIdRef:
      name: ldap-user-federation
      policy:
        resolve: Always
    name: custom-mapper
    providerId: "full-name-ldap-mapper"
    providerType: "org.keycloak.storage.ldap.mappers.LDAPStorageMapper"
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Custom UserFederation

```yaml
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: UserFederation
metadata:
  name: custom-user-federation
spec:
  deletionPolicy: Delete
  forProvider:
    config:
      dummyBool: true
      dummyString: foobar
      multivalue: value1##value2
    enabled: true
    name: custom
    providerId: custom
    realmIdSelector:
      matchLabels:
        testing.upbound.io/example-name: realm
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Related Resources

- [Realms](./realms.md)
- [Users](./users.md)
- [Groups](./groups.md)
- [Roles](./roles.md)
- [Identity Providers](./identity-providers.md)

