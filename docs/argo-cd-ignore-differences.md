# Using `provider-keycloak` with ArgoCD

Crossplane managed resources are *eventually-consistent*: the controller may
write back into `spec.forProvider` for two reasons that ArgoCD will see as
drift unless explicitly ignored.

1. **Cross-resource reference resolution.** When you author a managed
   resource with a `*Ref` or `*Selector`, Crossplane resolves it once and
   writes the resolved value (and the canonical `*Ref`) into
   `spec.forProvider`. This is a deliberate cache so that subsequent
   reconciliations don't re-resolve.
2. **Late-initialisation.** The Terraform provider may return defaults for
   fields the user left empty. By default upjet copies those into
   `spec.forProvider`. This provider already disables that copy for the
   known noisy fields (see
   [`docs/assessments/2026-04-client-forprovider-spec-drift.md`](assessments/2026-04-client-forprovider-spec-drift.md)),
   so for new installs you should only need to handle case 1.

If you manage Keycloak resources via ArgoCD, you almost certainly want to
ignore the reference write-back fields. Two options:

## Option 1 (recommended): per-resource `ignoreDifferences`

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  ignoreDifferences:
    # OpenID Client — silence the resolver write-back of the
    # authenticationFlowBindingOverrides.browserId / directGrantId fields.
    - group: openidclient.keycloak.crossplane.io
      kind: Client
      jqPathExpressions:
        - '.spec.forProvider.authenticationFlowBindingOverrides[].browserId'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].browserIdRef'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].directGrantId'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].directGrantIdRef'

    # SAML Client — same fields.
    - group: samlclient.keycloak.crossplane.io
      kind: Client
      jqPathExpressions:
        - '.spec.forProvider.authenticationFlowBindingOverrides[].browserId'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].browserIdRef'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].directGrantId'
        - '.spec.forProvider.authenticationFlowBindingOverrides[].directGrantIdRef'

    # Realm authentication bindings — flow IDs returned by Keycloak.
    - group: authentication.keycloak.crossplane.io
      kind: Bindings
      jqPathExpressions:
        - '.spec.forProvider.browserFlow'
        - '.spec.forProvider.registrationFlow'
        - '.spec.forProvider.directGrantFlow'
        - '.spec.forProvider.resetCredentialsFlow'
        - '.spec.forProvider.clientAuthenticationFlow'
        - '.spec.forProvider.dockerAuthenticationFlow'

    # Roles — composite role IDs.
    - group: role.keycloak.crossplane.io
      kind: Role
      jqPathExpressions:
        - '.spec.forProvider.compositeRoles'
```

## Option 2 (forward-looking): use `initProvider` + management policies

Move every reference selector / ref out of `spec.forProvider` into
`spec.initProvider` and disable late-initialisation by setting
`managementPolicies` to everything *except* `LateInitialize`:

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: my-client
spec:
  managementPolicies: ["Observe", "Create", "Update", "Delete"]
  forProvider:
    realmId: master
    clientId: my-client
    name: my-client
  initProvider:
    authenticationFlowBindingOverrides:
      - browserIdSelector:
          matchLabels:
            my-label: my-value
```

`spec.initProvider` is consulted only at *create* time, so the controller
never writes back into it. This requires `--enable-management-policies` on
the Crossplane core (stable since v1.17) and is the recommended
forward-looking pattern across all upjet-based providers.
