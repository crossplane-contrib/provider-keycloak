---
sidebar_position: 11
title: Client Authorization
description: Manage Keycloak authorization resources, permissions, and policies for OpenID clients
---

Use these resources when a client needs Keycloak Authorization Services for fine-grained access control and UMA-style policy evaluation. Define resources, permissions, and policies to protect APIs and services. The client must have `authorization` enabled.

## API Reference

| Kind | API Group | Terraform Resource | CRD Explorer |
|------|-----------|-------------------|---|
| ClientAuthorizationResource | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_authorization_resource`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_authorization_resource) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientAuthorizationResource/v1alpha1) |
| ClientAuthorizationPermission | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_authorization_permission`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_authorization_permission) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientAuthorizationPermission/v1alpha1) |
| ClientClientPolicy | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_client_policy`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_client_policy) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientClientPolicy/v1alpha1) |
| ClientGroupPolicy | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_group_policy`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_group_policy) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientGroupPolicy/v1alpha1) |
| ClientRolePolicy | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_role_policy`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_role_policy) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientRolePolicy/v1alpha1) |
| ClientUserPolicy | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_user_policy`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_user_policy) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientUserPolicy/v1alpha1) |
| ClientRegexPolicy | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_regex_policy`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_regex_policy) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientRegexPolicy/v1alpha1) |
| ClientPermissions | `openidclient.keycloak.crossplane.io/v1alpha1` | [`keycloak_openid_client_permissions`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/openid_client_permissions) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/openidclient.keycloak.crossplane.io/ClientPermissions/v1alpha1) |

## Working YAML Examples

### `ClientAuthorizationResource`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientAuthorizationResource
metadata:
  name: my-authz-resource
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: my-authz-resource
    displayName: My Authorization Resource
    type: "urn:test:resources:default"
    uris:
      - "/protected/resource"
    resourceServerIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

### `ClientAuthorizationPermission`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientAuthorizationPermission
metadata:
  name: my-authz-permission
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: my-authz-permission
    description: Permission covering all resources of a given type
    type: resource
    resourceType: "urn:test:resources:default"
    decisionStrategy: UNANIMOUS
    resourceServerIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

### `ClientClientPolicy`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientClientPolicy
metadata:
  name: my-client-policy
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: my-client-policy
    clientsRefs:
      - name: "test"
        policy:
          resolve: Always
    decisionStrategy: UNANIMOUS
    logic: POSITIVE
    resourceServerIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

### `ClientClientPolicy` with OIDC and SAML clients

`ClientClientPolicy` also supports SAML clients through `samlClientsRefs`.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientClientPolicy
metadata:
  name: my-oidc-and-saml-client-policy
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: my-oidc-and-saml-client-policy
    clientsRefs:
      - name: "test"
        policy:
          resolve: Always
    samlClientsRefs:
      - name: saml-client
        policy:
          resolve: Always
    decisionStrategy: UNANIMOUS
    logic: POSITIVE
    resourceServerIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

### `ClientGroupPolicy`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientGroupPolicy
metadata:
  name: my-group-policy
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: my-group-policy
    groups:
      - path: /test
        extendChildren: false
        idRef:
          name: "test"
    decisionStrategy: UNANIMOUS
    logic: POSITIVE
    resourceServerIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

### `ClientRolePolicy`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientRolePolicy
metadata:
  name: my-role-policy
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: my-role-policy
    type: role
    role:
      - required: true
        idRef:
          name: "test"
    decisionStrategy: UNANIMOUS
    logic: POSITIVE
    resourceServerIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

### `ClientUserPolicy`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientUserPolicy
metadata:
  name: my-user-policy
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    name: my-user-policy
    usersRefs:
      - name: "tim-tester"
        policy:
          resolve: Always
    decisionStrategy: UNANIMOUS
    logic: POSITIVE
    resourceServerIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

### `ClientRegexPolicy`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientRegexPolicy
metadata:
  name: regex-policy
spec:
  deletionPolicy: Delete
  forProvider:
    decisionStrategy: UNANIMOUS
    logic: POSITIVE
    name: regex-policy
    pattern: ^sample.+$
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    resourceServerIdRef:
      name: "test"
      policy:
        resolve: Always
    targetClaim: sample-claim
  providerConfigRef:
    name: "keycloak-provider-config"
```

### `ClientPermissions`

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientPermissions
metadata:
  name: my-permission
spec:
  providerConfigRef:
    name: "keycloak-provider-config"
  deletionPolicy: Delete
  forProvider:
    clientIdRef:
      name: "test"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
```

## Related Resources

- [Clients](./clients.md)
- [Groups](./groups.md)
- [Roles](./roles.md)
- [Users](./users.md)
- [SAML Clients](./saml-clients.md)
- [Realms](./realms.md)
