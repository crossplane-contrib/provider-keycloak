---
sidebar_position: 9
title: Authentication Flows
description: Manage Keycloak authentication flows declaratively
---

# Authentication Flows

Authentication flows define the sequence of steps a user must complete to authenticate. Flows consist of executions (individual authenticator steps), subflows (nested groups of steps), and bindings that attach flows to authentication contexts.

## API Reference

> **Schema source:** This page highlights common fields and examples. For the complete OpenAPI schema, including references, selectors, status fields, and connection details, see the generated CRDs in `package/crds/`.

- **API Group**: `authenticationflow.keycloak.crossplane.io`
- **API Version**: `v1alpha1`
- **Kinds**: `Flow`, `Subflow`, `Execution`, `ExecutionConfig`, `Bindings`

## Flow

A top-level authentication flow that groups executions and subflows.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Flow
metadata:
  name: custom-browser-flow
spec:
  forProvider:
    alias: "custom-browser"
    description: "Custom browser authentication flow"
    providerId: "basic-flow"
    realmId: "my-realm"
  providerConfigRef:
    name: keycloak-provider-config
```

## Subflow

A nested flow within a parent flow, used to group related authentication steps.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Subflow
metadata:
  name: mfa-subflow
spec:
  forProvider:
    alias: "mfa-subflow"
    description: "Multi-factor authentication subflow"
    authenticator: ""
    parentFlowAlias: "custom-browser"
    providerId: "basic-flow"
    priority: 10
    realmId: "my-realm"
  providerConfigRef:
    name: keycloak-provider-config
```

## Execution

An individual authenticator step within a flow or subflow.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Execution
metadata:
  name: username-password-form
spec:
  forProvider:
    authenticator: "auth-username-password-form"
    parentFlowAlias: "custom-browser"
    priority: 0
    realmId: "my-realm"
  providerConfigRef:
    name: keycloak-provider-config
```

### Execution in a Subflow

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Execution
metadata:
  name: otp-form
spec:
  forProvider:
    authenticator: "auth-otp-form"
    parentSubflowAlias: "mfa-subflow"
    parentFlowAlias: "custom-browser"
    priority: 0
    realmId: "my-realm"
  providerConfigRef:
    name: keycloak-provider-config
```

## ExecutionConfig

Configuration for a specific execution step.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: ExecutionConfig
metadata:
  name: idp-redirector-config
spec:
  forProvider:
    alias: "google-redirector"
    executionId: "execution-uuid"
    realmId: "my-realm"
    config:
      defaultProvider: "google"
  providerConfigRef:
    name: keycloak-provider-config
```

## Bindings

Bind authentication flows to specific authentication contexts in a realm.

```yaml
apiVersion: authenticationflow.keycloak.crossplane.io/v1alpha1
kind: Bindings
metadata:
  name: realm-auth-bindings
spec:
  forProvider:
    realmId: "my-realm"
    browserFlow: "custom-browser"
    directGrantFlow: "direct grant"
    clientAuthenticationFlow: "clients"
    dockerAuthenticationFlow: "docker auth"
    registrationFlow: "registration"
    resetCredentialsFlow: "reset credentials"
  providerConfigRef:
    name: keycloak-provider-config
```

## Key Fields

### Flow

| Field | Type | Description |
|-------|------|-------------|
| `alias` | string | Unique alias for the flow |
| `description` | string | Description of the flow |
| `providerId` | string | `basic-flow` or `client-flow` |
| `realmId` | string | Realm this flow belongs to |

### Subflow

| Field | Type | Description |
|-------|------|-------------|
| `alias` | string | Unique alias for the subflow |
| `authenticator` | string | Authenticator name |
| `parentFlowAlias` | string | Alias of the parent flow |
| `providerId` | string | `basic-flow` or `client-flow` |
| `priority` | number | Execution priority order |
| `realmId` | string | Realm this subflow belongs to |

### Execution

| Field | Type | Description |
|-------|------|-------------|
| `authenticator` | string | Name of the authenticator |
| `parentFlowAlias` | string | Alias of the parent flow |
| `parentSubflowAlias` | string | Alias of the parent subflow (optional) |
| `priority` | number | Execution priority order |
| `realmId` | string | Realm this execution belongs to |

### Bindings

| Field | Type | Description |
|-------|------|-------------|
| `realmId` | string | Realm to configure bindings for |
| `browserFlow` | string | Flow alias for browser authentication |
| `directGrantFlow` | string | Flow alias for direct grant |
| `clientAuthenticationFlow` | string | Flow alias for client authentication |
| `registrationFlow` | string | Flow alias for registration |
| `resetCredentialsFlow` | string | Flow alias for password reset |
| `dockerAuthenticationFlow` | string | Flow alias for Docker authentication |
