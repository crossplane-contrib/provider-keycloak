---
sidebar_position: 9
title: Authentication Flows
description: Customize Keycloak login, MFA, and flow bindings declaratively
---

Use authentication flow resources when the default Keycloak login process is not enough. They let you define custom browser, registration, direct-grant, or client-authentication flows; nest subflows; add execution steps such as MFA, OTP, WebAuthn, or identity-provider redirects; and bind the finished flow to the realm behavior that should use it.

## API Reference

| Kind | API Group | Terraform | CRD Explorer |
|------|-----------|-----------|---|
| `Flow` | `authenticationflow.keycloak.crossplane.io/v1alpha1` | [`keycloak_authentication_flow`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/authentication_flow) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/authenticationflow.keycloak.crossplane.io/Flow/v1alpha1) |
| `Subflow` | `authenticationflow.keycloak.crossplane.io/v1alpha1` | [`keycloak_authentication_subflow`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/authentication_subflow) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/authenticationflow.keycloak.crossplane.io/Subflow/v1alpha1) |
| `Execution` | `authenticationflow.keycloak.crossplane.io/v1alpha1` | [`keycloak_authentication_execution`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/authentication_execution) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/authenticationflow.keycloak.crossplane.io/Execution/v1alpha1) |
| `ExecutionConfig` | `authenticationflow.keycloak.crossplane.io/v1alpha1` | [`keycloak_authentication_execution_config`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/authentication_execution_config) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/authenticationflow.keycloak.crossplane.io/ExecutionConfig/v1alpha1) |
| `Bindings` | `authenticationflow.keycloak.crossplane.io/v1alpha1` | [`keycloak_authentication_bindings`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/authentication_bindings) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/authenticationflow.keycloak.crossplane.io/Bindings/v1alpha1) |

## Working YAML examples

### Flow

Use a `Flow` to create the top-level container for a custom authentication sequence.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Flow
metadata:
  name: flow
spec:
  deletionPolicy: Delete
  forProvider:
    alias: my-flow-alias
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Subflow

Use a `Subflow` to group steps inside a parent flow and apply its own requirement.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Subflow
metadata:
  name: subflow
  labels:
    subflow-type: test-subflow
spec:
  deletionPolicy: Delete
  forProvider:
    alias: my-subflow-alias-1
    parentFlowAliasRef:
      name: flow
      policy:
        resolve: Always
    providerId: basic-flow
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    requirement: ALTERNATIVE
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Execution using `parentFlowAliasRef`

Use an `Execution` directly under a top-level flow when the step should run without an intermediate subflow.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Execution
metadata:
  name: execution-one
spec:
  deletionPolicy: Delete
  forProvider:
    authenticator: auth-cookie
    parentFlowAliasRef:
      name: flow
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    requirement: ALTERNATIVE
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Execution using `parentSubflowAliasRef`

Use `parentSubflowAliasRef` when the execution should be nested inside a specific subflow object.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Execution
metadata:
  name: execution-in-subflow-ref
spec:
  deletionPolicy: Delete
  forProvider:
    authenticator: auth-username-password-form
    parentSubflowAliasRef:
      name: subflow
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    requirement: REQUIRED
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Execution using `parentSubflowAliasSelector`

Use the selector form when you want to target a subflow by labels instead of by a fixed name.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Execution
metadata:
  name: execution-in-subflow-selector
spec:
  deletionPolicy: Delete
  forProvider:
    authenticator: auth-otp-form
    parentSubflowAliasSelector:
      matchLabels:
        subflow-type: test-subflow
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
    requirement: REQUIRED
  providerConfigRef:
    name: "keycloak-provider-config"
```

### ExecutionConfig

Use `ExecutionConfig` when an execution needs extra configuration, such as the default identity provider for an IdP redirector.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: ExecutionConfig
metadata:
  name: execution-identity-redirect-config
spec:
  deletionPolicy: Delete
  forProvider:
    alias: my-config-alias
    config:
      defaultProvider: my-config-default-idp
    executionIdRef:
      name: execution-identity-redirect
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Bindings

Use `Bindings` to assign your custom flow to browser, registration, direct grant, or other realm authentication entry points.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Bindings
metadata:
  name: browser-authentication-binding
spec:
  deletionPolicy: Delete
  forProvider:
    dockerAuthenticationFlowRef:
      name: "flow"
      policy:
        resolve: Always
    realmIdRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Key fields

### Flow and Subflow

| Field | Applies to | Why it matters |
|-------|------------|----------------|
| `alias` | `Flow`, `Subflow` | Stable name used by executions and bindings. |
| `realmIdRef` | `Flow`, `Subflow` | Selects the realm that owns the flow. |
| `providerId` | `Subflow` | Chooses the Keycloak flow type, typically `basic-flow`. |
| `parentFlowAliasRef` | `Subflow` | Attaches the subflow to its parent flow. |
| `requirement` | `Subflow` | Controls whether the subflow is `REQUIRED`, `ALTERNATIVE`, and so on. |

### Execution and ExecutionConfig

| Field | Applies to | Why it matters |
|-------|------------|----------------|
| `authenticator` | `Execution` | Selects the actual Keycloak authenticator, such as `auth-cookie` or `auth-otp-form`. |
| `parentFlowAliasRef` | `Execution` | Places the execution directly under a top-level flow. |
| `parentSubflowAliasRef` / `parentSubflowAliasSelector` | `Execution` | Places the execution inside a specific subflow. |
| `requirement` | `Execution` | Determines whether the authenticator is required, optional, or alternative. |
| `executionIdRef` | `ExecutionConfig` | Resolves the execution that receives the configuration block. |
| `config` | `ExecutionConfig` | Holds authenticator-specific configuration such as `defaultProvider`. |

### Bindings

| Field | Why it matters |
|-------|----------------|
| `realmIdRef` | Selects the realm whose authentication bindings are being changed. |
| `browserFlowRef` | Binds a custom flow to browser logins. |
| `registrationFlowRef` | Binds a custom flow to self-registration. |
| `directGrantFlowRef` | Binds a flow to direct access grant authentication. |
| `resetCredentialsFlowRef` | Binds a flow to reset-credentials behavior. |
| `clientAuthenticationFlowRef` | Binds a flow to client authentication. |
| `dockerAuthenticationFlowRef` | Binds a flow to Docker authentication. |

## Realm vs. Bindings ownership

The six flow bindings (`browserFlow`, `registrationFlow`, `directGrantFlow`,
`resetCredentialsFlow`, `clientAuthenticationFlow`, `dockerAuthenticationFlow`)
are fields of the Keycloak realm itself. They can be set either on the `Realm`
managed resource or through a `Bindings` resource. Pick **one** owner per field:

- Declare a binding on **either** the `Realm` **or** the `Bindings` resource,
  never on both. If both declare different values, each reconcile of one
  resource undoes the other and the value oscillates.
- A `Realm` that leaves these fields unset does not manage them: the provider
  does not late-initialize them into `spec.forProvider`, and a realm update
  re-sends the currently observed value, so a `Bindings` resource for the same
  realm keeps its value.
- Realms created with an older provider version may already carry
  late-initialized values (for example `browserFlow: browser`) in
  `spec.forProvider`. Remove those fields from the `Realm` (for example with
  `kubectl edit`) if a `Bindings` resource should own them; otherwise the
  `Realm` keeps resetting them.

### Binding a custom flow on a new realm

Keycloak rejects a realm update whose binding references a flow alias that
does not exist yet (`500 Failed to update realm`; the previous value is kept).
A `Realm` that declares `browserFlow: my-custom-flow` in the same change that
creates the realm therefore stays `Synced=False` until the flow exists, which
can deadlock sync-wave ordered GitOps setups (the flow can only be created
after the realm). Either:

- use a `Bindings` resource with `browserFlowRef` (it waits for the referenced
  `Flow` to resolve before applying the binding), or
- roll out in two phases: create the realm and the flow first, then set the
  binding on the `Realm` in a follow-up change.

## Related Resources

- **[Identity Providers](./identity-providers.md)** — Combine IdP redirectors and external authentication with custom flows.
- **[Clients](./clients.md)** — Understand which applications consume the flows you bind.
- **[Realms](./realms.md)** — Manage the realm that owns the flows and bindings.

