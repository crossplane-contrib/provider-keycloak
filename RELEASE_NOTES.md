# Release Notes — provider-keycloak v2.20.0

These notes supplement the auto-generated changelog.

## New Features

### 1. Workflow CRD

Manage Keycloak realm automation workflows (e.g., user onboarding, offboarding) declaratively.

```yaml
apiVersion: workflow.keycloak.crossplane.io/v1alpha1
kind: Workflow
metadata:
  name: onboarding
spec:
  forProvider:
    enabled: true
    name: onboarding-new-users
    "on": user_created
    realmRef:
      name: my-realm
    step:
    - config:
        message: |
          <p>Dear ${user.firstName} ${user.lastName}, </p>
          <p>Welcome to ${realm.displayName}!</p>
          <p>Best regards,<br/>The Keycloak Team</p>
      uses: notify-user
    - after: "2592000000"
      config:
        action: UPDATE_PASSWORD
      uses: set-user-required-action
  providerConfigRef:
    name: keycloak-provider-config
```

---

### 2. SPIFFE Identity Provider CRD

Federate workload identity via SPIFFE trust domains.

```yaml
apiVersion: identityprovider.keycloak.crossplane.io/v1alpha1
kind: SpiffeIdentityProvider
metadata:
  name: my-spiffe-idp
spec:
  forProvider:
    alias: my-spiffe-idp
    bundleEndpoint: https://spiffe-bundle.example.com/bundle
    trustDomain: spiffe://my-trust-domain
    realmRef:
      name: my-realm
  providerConfigRef:
    name: keycloak-provider-config
```

---

### 3. OIDC OpenShift v4 Identity Provider CRD

Integrate Keycloak with OpenShift 4 clusters as an OIDC identity provider.

```yaml
apiVersion: identityprovider.keycloak.crossplane.io/v1alpha1
kind: OidcOpenShiftV4IdentityProvider
metadata:
  name: openshift-v4
spec:
  forProvider:
    baseUrl: https://openshift.example.com:8443
    clientId: my-openshift-client
    clientSecretSecretRef:
      key: client-secret
      name: openshift-credentials
      namespace: crossplane-system
    defaultScopes: user:full
    syncMode: IMPORT
    trustEmail: true
    realmRef:
      name: my-realm
  providerConfigRef:
    name: keycloak-provider-config
```

---

### 4. Client Regex Policy CRD

Define authorization policies based on regex patterns against token claims.

```yaml
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: ClientRegexPolicy
metadata:
  name: email-domain-policy
spec:
  forProvider:
    name: email-domain-policy
    decisionStrategy: UNANIMOUS
    logic: POSITIVE
    pattern: "^.+@example\\.com$"
    targetClaim: email
    realmIdRef:
      name: my-realm
    resourceServerIdRef:
      name: my-client
  providerConfigRef:
    name: keycloak-provider-config
```

---

## Bug Fixes

- **fix:** Consider `parent_id` when resolving group by identifying properties (#544)
- **fix:** Remove `ClientServiceAccountRole` validation wrapper — fixes reconcile errors on service account role assignments
- **fix:** Empty identifier guard for `ClientServiceAccountRole`

## Dependency Updates

- Bumped upstream `terraform-provider-keycloak` to **v5.8.0**
- Updated `k8s.io` dependencies to **v0.35.4**

---

## Upgrade Notes

No breaking changes. New CRDs are additive. Install the updated provider package and the new CRDs will be available automatically.
