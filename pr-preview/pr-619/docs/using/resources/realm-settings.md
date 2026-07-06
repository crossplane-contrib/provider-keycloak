# Realm Settings

Use these resources after a `Realm` exists and you need to shape how that realm behaves in production. They cover audit logging, user onboarding requirements, custom profile fields, signing keys, default scopes, and client policy enforcement.

## API Reference

- **`RealmEvents`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_realm_events`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm_events) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/RealmEvents/v1alpha1)
- **`RequiredAction`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_required_action`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/required_action) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/RequiredAction/v1alpha1)
- **`UserProfile`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_realm_user_profile`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm_user_profile) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/UserProfile/v1alpha1)
- **`KeystoreRsa`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_realm_keystore_rsa`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm_keystore_rsa) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/KeystoreRsa/v1alpha1)
- **`DefaultClientScopes`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_realm_default_client_scopes`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm_default_client_scopes) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/DefaultClientScopes/v1alpha1)
- **`OptionalClientScopes`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_realm_optional_client_scopes`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm_optional_client_scopes) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/OptionalClientScopes/v1alpha1)
- **`ClientPolicyProfile`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_realm_client_policy_profile`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm_client_policy_profile) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/ClientPolicyProfile/v1alpha1)
- **`ClientPolicyProfilePolicy`** — API: `realm.keycloak.crossplane.io/v1alpha1` — Terraform: [`keycloak_realm_client_policy_profile_policy`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/realm_client_policy_profile_policy) — CRD Explorer: [View Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/realm.keycloak.crossplane.io/ClientPolicyProfilePolicy/v1alpha1)

## Working YAML Examples

### RealmEvents

Use `RealmEvents` to configure audit logging for user events such as `LOGIN` and `LOGOUT`, plus admin event tracking.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: RealmEvents
metadata:
  name: realm-events
spec:
  deletionPolicy: Delete
  forProvider:
    adminEventsDetailsEnabled: true
    adminEventsEnabled: true
    enabledEventTypes:
      - LOGIN
      - LOGOUT
    eventsEnabled: true
    eventsExpiration: 3600
    eventsListeners:
      - jboss-logging
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### RequiredAction

Use `RequiredAction` to enable tasks users must complete, such as setting a password, verifying email, or registering WebAuthn.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: RequiredAction
metadata:
  name: required-action
spec:
  deletionPolicy: Delete
  forProvider:
    alias: webauthn-register
    enabled: true
    name: Webauthn Register
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### UserProfile

Use `UserProfile` to define custom user attributes with validation, permissions, and grouping.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: UserProfile
metadata:
  name: userprofile
spec:
  deletionPolicy: Delete
  forProvider:
    attribute:
      - displayName: ""
        group: ""
        multiValued: false
        name: username
      - displayName: ""
        group: ""
        multiValued: false
        name: email
      - annotations:
          foo: bar
        displayName: Field 1
        enabledWhenScope:
          - offline_access
        group: group1
        multiValued: false
        name: field1
        permissions:
          - edit:
              - admin
              - user
            view:
              - admin
              - user
        requiredForRoles:
          - user
        requiredForScopes:
          - offline_access
        validator:
          - name: person-name-prohibited-characters
          - config:
              error-message: Nope
              pattern: ^[a-z]+$
            name: pattern
    group:
      - annotations:
          foo: bar
          foo2: '{"key":"val"}'
        displayDescription: A first group
        displayHeader: Group 1
        name: group1
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    unmanagedAttributePolicy: ENABLED
  providerConfigRef:
    name: "keycloak-provider-config"
```

### KeystoreRsa

Use `KeystoreRsa` to manage RSA signing keys used for token signing and verification.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: KeystoreRsa
metadata:
  name: keystore-rsa
spec:
  deletionPolicy: Delete
  forProvider:
    active: true
    algorithm: RS256
    certificateSecretRef:
      key: cert
      name: rsa-key
      namespace: dev
    enabled: true
    name: my-rsa-key
    priority: 100
    privateKeySecretRef:
      key: priv
      name: rsa-key
      namespace: dev
    providerId: rsa
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### DefaultClientScopes

Use `DefaultClientScopes` to define which client scopes are assigned automatically to new clients in the realm.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: DefaultClientScopes
metadata:
  name: dev-default-scopes
spec:
  deletionPolicy: Delete
  forProvider:
    realmId: "dev"
    defaultScopes:
      - profile
      - email
      - roles
      - web-origins
      - phone
  providerConfigRef:
    name: "keycloak-provider-config"
```

### OptionalClientScopes

Use `OptionalClientScopes` to define which scopes clients may request optionally.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: OptionalClientScopes
metadata:
  name: dev-optional-scopes
spec:
  deletionPolicy: Delete
  forProvider:
    realmId: "dev"
    optionalScopes:
      - acr
      - role_list
  providerConfigRef:
    name: "keycloak-provider-config"
```

### ClientPolicyProfile

Use `ClientPolicyProfile` to define policy profiles with executors that enforce standards on clients.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: ClientPolicyProfile
metadata:
  name: client-policy-profile
spec:
  deletionPolicy: Delete
  forProvider:
    executor:
      - configuration:
          auto-configure: "true"
        name: intent-client-bind-checker
      - name: secure-session
    name: my-profile
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### ClientPolicyProfilePolicy

Use `ClientPolicyProfilePolicy` to define the conditions that trigger one or more client policy profiles.

```yaml
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: ClientPolicyProfilePolicy
metadata:
  name: client-policy-profile-policy
spec:
  deletionPolicy: Delete
  forProvider:
    condition:
      - configuration:
          protocol: openid-connect
        name: client-type
    description: Some desc
    name: my-policy
    profilesRefs:
      - name: client-policy-profile
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Key Fields

| Resource | Key fields | Description |
| --- | --- | --- |
| `RealmEvents` | `realmIdRef`, `eventsEnabled`, `enabledEventTypes`, `eventsListeners`, `adminEventsEnabled`, `adminEventsDetailsEnabled`, `eventsExpiration` | Controls user and admin audit event capture and retention. |
| `RequiredAction` | `realmIdRef`, `alias`, `name`, `enabled` | Enables built-in actions users must complete during account lifecycle flows. |
| `UserProfile` | `realmIdRef`, `attribute`, `group`, `unmanagedAttributePolicy` | Defines custom profile schema, validation, permissions, and grouping. |
| `KeystoreRsa` | `realmIdRef`, `name`, `providerId`, `algorithm`, `active`, `enabled`, `priority`, `privateKeySecretRef`, `certificateSecretRef` | Manages RSA key material used by the realm. |
| `DefaultClientScopes` | `realmId`, `defaultScopes` | Declares scopes assigned automatically to new clients. |
| `OptionalClientScopes` | `realmId`, `optionalScopes` | Declares scopes clients can request optionally. |
| `ClientPolicyProfile` | `realmIdRef`, `name`, `executor` | Defines reusable client policy executors. |
| `ClientPolicyProfilePolicy` | `realmIdRef`, `name`, `description`, `condition`, `profilesRefs` | Binds policy conditions to one or more client policy profiles. |

## Related Resources

- [Realms](./realms.md)
- [Default Configuration](./default-config.md)

