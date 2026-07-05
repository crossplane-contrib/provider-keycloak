---
sidebar_position: 1
title: Realms
description: Create and manage Keycloak realms, the top-level container for every other resource
---

Use a `Realm` when you need an isolated Keycloak boundary for a tenant, environment, or application domain. Because every other Keycloak resource belongs to a realm, this is usually the first resource you create for a new deployment.

## API Reference

- **`Realm`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_realm`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/Realm/v1alpha1)

## Working YAML Examples

### Basic realm

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: dev
spec:
  deletionPolicy: Delete
  forProvider:
    realm: "dev"
    attributes:
      userProfileEnabled: "true"
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Realm with timeouts and lifespans

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: dev-durations
spec:
  deletionPolicy: Delete
  forProvider:
    realm: "dev-durations"
    enabled: true
    accessTokenLifespan: "5m0s"
    accessTokenLifespanForImplicitFlow: "1800s"
    ssoSessionIdleTimeout: "30m0s"
    ssoSessionMaxLifespan: "10h0m0s"
    ssoSessionIdleTimeoutRememberMe: "0s"
    ssoSessionMaxLifespanRememberMe: "0s"
    offlineSessionIdleTimeout: "720h0m0s"
    offlineSessionMaxLifespan: "1440h0m0s"
    clientSessionIdleTimeout: "0s"
    clientSessionMaxLifespan: "0s"
    accessCodeLifespan: "1m0s"
    accessCodeLifespanUserAction: "5m0s"
    accessCodeLifespanLogin: "30m0s"
    actionTokenGeneratedByAdminLifespan: "12h0m0s"
    actionTokenGeneratedByUserLifespan: "5m0s"
    oauth2DeviceCodeLifespan: "10m0s"
    oauth2DevicePollingInterval: 5
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Managing an existing realm without deleting it

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: existing-master
spec:
  deletionPolicy: Orphan
  forProvider:
    realm: master
    displayName: Customized Keycloak
  providerConfigRef:
    name: "keycloak-provider-config"
  managementPolicies: [Observe, Update]
```

### Realm with organizations enabled

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: orgs
spec:
  deletionPolicy: Delete
  forProvider:
    realm: "orgs"
    organizationsEnabled: true
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Key Fields

| Field | Description |
| --- | --- |
| `realm` | Realm ID and top-level container name used by all child resources. |
| `enabled` | Turns the realm on or off. |
| `displayName` | Human-friendly name shown in the Keycloak UI. |
| `passwordPolicy` | Password rules enforced for users in the realm. |
| `attributes` | Extra realm settings such as feature flags and provider-specific options. |
| `smtpServer` | Outbound email settings for verification, reset, and notification flows. |
| `otpPolicy` | Realm-wide OTP settings for MFA behavior. |
| `organizationsEnabled` | Enables organization features in supported Keycloak versions. |
| `accessTokenLifespan` | Default lifetime for access tokens. |
| `accessTokenLifespanForImplicitFlow` | Access token lifetime for implicit flow clients. |
| `ssoSessionIdleTimeout` | Idle timeout before a normal SSO session expires. |
| `ssoSessionMaxLifespan` | Maximum duration of a normal SSO session. |
| `ssoSessionIdleTimeoutRememberMe` | Idle timeout for remember-me SSO sessions. |
| `ssoSessionMaxLifespanRememberMe` | Maximum duration for remember-me SSO sessions. |
| `offlineSessionIdleTimeout` | Idle timeout for offline sessions and refresh tokens. |
| `offlineSessionMaxLifespan` | Maximum duration for offline sessions. |
| `clientSessionIdleTimeout` | Idle timeout for client sessions. |
| `clientSessionMaxLifespan` | Maximum duration for client sessions. |
| `accessCodeLifespan` | Lifetime of authorization codes. |
| `accessCodeLifespanUserAction` | Lifetime for user action tokens during browser flows. |
| `accessCodeLifespanLogin` | Maximum time allowed to complete login. |
| `actionTokenGeneratedByAdminLifespan` | Lifetime for admin-generated action tokens. |
| `actionTokenGeneratedByUserLifespan` | Lifetime for user-generated action tokens. |
| `oauth2DeviceCodeLifespan` | Lifetime for OAuth 2.0 device codes. |
| `oauth2DevicePollingInterval` | Polling interval for device authorization clients. |

## Related Resources

- [Realm Settings](./realm-settings.md)
- [Default Configuration](./default-config.md)
