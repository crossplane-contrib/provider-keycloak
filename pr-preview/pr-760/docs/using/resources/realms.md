# Realms

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


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
    realm: &#34;dev&#34;
    attributes:
      userProfileEnabled: &#34;true&#34;
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
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
    realm: &#34;dev-durations&#34;
    enabled: true
    accessTokenLifespan: &#34;5m0s&#34;
    accessTokenLifespanForImplicitFlow: &#34;1800s&#34;
    ssoSessionIdleTimeout: &#34;30m0s&#34;
    ssoSessionMaxLifespan: &#34;10h0m0s&#34;
    ssoSessionIdleTimeoutRememberMe: &#34;0s&#34;
    ssoSessionMaxLifespanRememberMe: &#34;0s&#34;
    offlineSessionIdleTimeout: &#34;720h0m0s&#34;
    offlineSessionMaxLifespan: &#34;1440h0m0s&#34;
    clientSessionIdleTimeout: &#34;0s&#34;
    clientSessionMaxLifespan: &#34;0s&#34;
    accessCodeLifespan: &#34;1m0s&#34;
    accessCodeLifespanUserAction: &#34;5m0s&#34;
    accessCodeLifespanLogin: &#34;30m0s&#34;
    actionTokenGeneratedByAdminLifespan: &#34;12h0m0s&#34;
    actionTokenGeneratedByUserLifespan: &#34;5m0s&#34;
    oauth2DeviceCodeLifespan: &#34;10m0s&#34;
    oauth2DevicePollingInterval: 5
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
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
    name: &#34;keycloak-provider-config&#34;
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
    realm: &#34;orgs&#34;
    organizationsEnabled: true
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
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


