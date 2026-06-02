---
sidebar_position: 15
title: Realm Settings
description: Manage realm-level sub-resources like events, required actions, user profiles, and keystores
---

# Realm Settings

Beyond the core [Realm](./realms.md) resource, Keycloak provides additional realm-level configuration for events, required actions, user profiles, keystores, default scopes, and client policies.

## API Reference

- **API Group**: `realm.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kinds**: `RealmEvents`, `RequiredAction`, `UserProfile`, `KeystoreRsa`, `DefaultClientScopes`, `OptionalClientScopes`, `ClientPolicyProfile`, `ClientPolicyProfilePolicy`

## RealmEvents

Configure event logging and listeners for a realm.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: RealmEvents
metadata:
  name: my-realm-events
spec:
  forProvider:
    realmId: "my-realm"
    eventsEnabled: true
    eventsExpiration: 604800
    eventsListeners:
      - "jboss-logging"
    enabledEventTypes:
      - "LOGIN"
      - "LOGIN_ERROR"
      - "LOGOUT"
      - "REGISTER"
    adminEventsEnabled: true
    adminEventsDetailsEnabled: true
  providerConfigRef:
    name: keycloak-provider-config
```

## RequiredAction

Configure required actions that users must complete (e.g., verify email, configure OTP).

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: RequiredAction
metadata:
  name: verify-email-action
spec:
  forProvider:
    realmId: "my-realm"
    alias: "VERIFY_EMAIL"
    name: "Verify Email"
    enabled: true
    defaultAction: true
    priority: 10
  providerConfigRef:
    name: keycloak-provider-config
```

### Required Action with Config

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: RequiredAction
metadata:
  name: configure-otp-action
spec:
  forProvider:
    realmId: "my-realm"
    alias: "CONFIGURE_TOTP"
    name: "Configure OTP"
    enabled: true
    defaultAction: false
    priority: 20
    config:
      otpPolicyAlgorithm: "HmacSHA1"
  providerConfigRef:
    name: keycloak-provider-config
```

## UserProfile

Define the user profile attributes and groups for a realm.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: UserProfile
metadata:
  name: my-realm-user-profile
spec:
  forProvider:
    realmId: "my-realm"
    unmanagedAttributePolicy: "ADMIN_VIEW"
    attribute:
      - name: "username"
        displayName: "Username"
        required:
          - roles:
              - "user"
        permissions:
          - edit:
              - "admin"
            view:
              - "admin"
              - "user"
        validator:
          - name: "length"
            config:
              min: "3"
              max: "64"
      - name: "email"
        displayName: "Email"
        required:
          - roles:
              - "user"
        validator:
          - name: "email"
    group:
      - name: "user-metadata"
        displayHeader: "User Metadata"
        displayDescription: "Additional user information"
  providerConfigRef:
    name: keycloak-provider-config
```

## KeystoreRsa

Manage RSA keystores for realm-level signing and encryption.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: KeystoreRsa
metadata:
  name: my-realm-rsa-key
spec:
  forProvider:
    name: "rsa-signing-key"
    realmId: "my-realm"
    enabled: true
    active: true
    priority: 100
    algorithm: "RS256"
    privateKeySecretRef:
      name: rsa-private-key
      namespace: crossplane-system
      key: private-key
    certificateSecretRef:
      name: rsa-certificate
      namespace: crossplane-system
      key: certificate
  providerConfigRef:
    name: keycloak-provider-config
```

## DefaultClientScopes

Define realm-level default client scopes assigned to all new clients.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: DefaultClientScopes
metadata:
  name: my-realm-default-scopes
spec:
  forProvider:
    realmId: "my-realm"
    defaultScopes:
      - "profile"
      - "email"
      - "roles"
      - "web-origins"
  providerConfigRef:
    name: keycloak-provider-config
```

## OptionalClientScopes

Define realm-level optional client scopes available to all clients.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: OptionalClientScopes
metadata:
  name: my-realm-optional-scopes
spec:
  forProvider:
    realmId: "my-realm"
    optionalScopes:
      - "address"
      - "phone"
      - "offline_access"
      - "microprofile-jwt"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientPolicyProfile

Define a client policy profile with executors that enforce rules on clients.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: ClientPolicyProfile
metadata:
  name: secure-client-profile
spec:
  forProvider:
    name: "secure-client-profile"
    description: "Profile enforcing security best practices"
    realmId: "my-realm"
    executor:
      - executorAlias: "secure-ciba-auth-req-signed"
        executorId: "secure-ciba-auth-request-signed"
      - executorAlias: "pkce-enforcer"
        executorId: "pkce-enforcer"
  providerConfigRef:
    name: keycloak-provider-config
```

## ClientPolicyProfilePolicy

Define a policy that associates conditions with client policy profiles.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: ClientPolicyProfilePolicy
metadata:
  name: secure-client-policy
spec:
  forProvider:
    name: "secure-client-policy"
    description: "Enforce secure profile on confidential clients"
    realmId: "my-realm"
    enabled: true
    profiles:
      - "secure-client-profile"
    condition:
      - conditionAlias: "client-access-type"
        conditionId: "client-accesstype"
        config:
          is-confidential-client: "true"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

### RealmEvents

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm to configure events for |
| `eventsEnabled` | bool | Enable saving login events (default `false`) |
| `eventsExpiration` | number | Event retention time in seconds |
| `eventsListeners` | []string | Event listener names |
| `enabledEventTypes` | []string | Event types to record |
| `adminEventsEnabled` | bool | Enable saving admin events (default `false`) |
| `adminEventsDetailsEnabled` | bool | Include details in admin events (default `false`) |

### RequiredAction

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm this action belongs to |
| `alias` | string | Action alias (e.g., `VERIFY_EMAIL`, `CONFIGURE_TOTP`) |
| `name` | string | Display name in the UI |
| `enabled` | bool | Whether the action is available |
| `defaultAction` | bool | Apply to all new users by default |
| `priority` | number | Execution order (lower = higher priority) |
| `config` | map | Action-specific configuration |

### KeystoreRsa

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name of the key |
| `realmId` | string | Realm this keystore belongs to |
| `active` | bool | Use for signing (default `true`) |
| `enabled` | bool | Key is accessible (default `true`) |
| `algorithm` | string | Algorithm (default `RS256`) |
| `priority` | number | Provider priority (default `0`) |
| `privateKeySecretRef` | ref | Reference to the private key secret |
| `certificateSecretRef` | ref | Reference to the certificate secret |
