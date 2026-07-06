# Identity Providers

# Identity Providers

Use identity providers when users should sign in to Keycloak with an external identity system instead of local usernames and passwords. This is the right fit for social login, corporate SAML or OIDC federation, Kubernetes or OpenShift workload identity, SPIFFE-based trust, and controlled token exchange between clients and external providers.

## API Reference

| Kind | API Group | Terraform | CRD Explorer |
|------|-----------|-----------|---|
| `IdentityProvider` | `oidc.keycloak.crossplane.io/v1alpha1` | [`keycloak_oidc_identity_provider`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/oidc_identity_provider) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/oidc.keycloak.crossplane.io/IdentityProvider/v1alpha1) |
| `GoogleIdentityProvider` | `oidc.keycloak.crossplane.io/v1alpha1` | [`keycloak_oidc_google_identity_provider`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/oidc_google_identity_provider) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/oidc.keycloak.crossplane.io/GoogleIdentityProvider/v1alpha1) |
| `IdentityProvider` | `saml.keycloak.crossplane.io/v1alpha1` | [`keycloak_saml_identity_provider`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/saml_identity_provider) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/saml.keycloak.crossplane.io/IdentityProvider/v1alpha1) |
| `IdentityProviderMapper` | `identityprovider.keycloak.crossplane.io/v1alpha1` | [`keycloak_custom_identity_provider_mapper`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/custom_identity_provider_mapper) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/identityprovider.keycloak.crossplane.io/IdentityProviderMapper/v1alpha1) |
| `KubernetesIdentityProvider` | `identityprovider.keycloak.crossplane.io/v1alpha1` | [`keycloak_kubernetes_identity_provider`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/kubernetes_identity_provider) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/identityprovider.keycloak.crossplane.io/KubernetesIdentityProvider/v1alpha1) |
| `OidcOpenShiftV4IdentityProvider` | `identityprovider.keycloak.crossplane.io/v1alpha1` | [`keycloak_oidc_openshift_v4_identity_provider`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/oidc_openshift_v4_identity_provider) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/identityprovider.keycloak.crossplane.io/OidcOpenShiftV4IdentityProvider/v1alpha1) |
| `SpiffeIdentityProvider` | `identityprovider.keycloak.crossplane.io/v1alpha1` | [`keycloak_spiffe_identity_provider`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/spiffe_identity_provider) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/identityprovider.keycloak.crossplane.io/SpiffeIdentityProvider/v1alpha1) |
| `ProviderTokenExchangeScopePermission` | `identityprovider.keycloak.crossplane.io/v1alpha1` | [`keycloak_identity_provider_token_exchange_scope_permission`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/identity_provider_token_exchange_scope_permission) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/identityprovider.keycloak.crossplane.io/ProviderTokenExchangeScopePermission/v1alpha1) |

## Working YAML examples

### OIDC Identity Provider

Use this resource for a generic OpenID Connect identity provider when you need explicit authorization and token endpoints.

```yaml
apiVersion: oidc.keycloak.crossplane.io/v1alpha1
kind: IdentityProvider
metadata:
  name: oidc-identity-provider
spec:
  deletionPolicy: Delete
  forProvider:
    alias: my-idp
    authorizationUrl: https://authorizationurl.com
    clientIdSecretRef:
      key: client-id
      name: client-secret
      namespace: dev
    clientSecretSecretRef:
      key: client-secret
      name: client-secret
      namespace: dev
    extraConfig:
      clientAuthMethod: client_secret_post
    realmRef:
      name: "dev"
      policy:
        resolve: Always
    tokenUrl: https://tokenurl.com
  providerConfigRef:
    name: "keycloak-provider-config"
```

### OIDC Identity Provider with organization binding

Use organization binding when the external IdP should route users into a specific Keycloak organization.

```yaml
apiVersion: oidc.keycloak.crossplane.io/v1alpha1
kind: IdentityProvider
metadata:
  name: org-provider
spec:
  deletionPolicy: Delete
  forProvider:
    alias: my-idp
    authorizationUrl: https://authorizationurl.com
    clientIdSecretRef:
      key: client-id
      name: client-secret
      namespace: dev
    clientSecretSecretRef:
      key: client-secret
      name: client-secret
      namespace: dev
    extraConfig:
      clientAuthMethod: client_secret_post
    realmRef:
      name: "orgs"
      policy:
        resolve: Always
    tokenUrl: https://tokenurl.com
    orgDomain: example.com
    orgRedirectModeEmailMatches: true
    organizationIdRef:
      name: example
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Google Identity Provider

Use the Google-specific CRD when you want the provider defaults and options exposed by `keycloak_oidc_google_identity_provider`.

```yaml
apiVersion: oidc.keycloak.crossplane.io/v1alpha1
kind: GoogleIdentityProvider
metadata:
  name: google
spec:
  forProvider:
    alias: google-idp
    clientIdSecretRef:
      key: client-id
      name: client-secret
      namespace: dev
    clientSecretSecretRef:
      key: client-secret
      name: client-secret
      namespace: dev
    hostedDomain: example.com
    realmRef:
      name: "dev"
      policy:
        resolve: Always
    syncMode: IMPORT
    trustEmail: true
  providerConfigRef:
    name: "keycloak-provider-config"
```

### SAML Identity Provider

Use the SAML CRD for enterprise identity providers that publish SAML metadata and SSO/SLO endpoints.

```yaml
apiVersion: saml.keycloak.crossplane.io/v1alpha1
kind: IdentityProvider
metadata:
  name: saml-identity-provider
spec:
  deletionPolicy: Delete
  forProvider:
    alias: my-saml-idp
    backchannelSupported: true
    entityId: https://domain.com/entity_id
    forceAuthn: true
    postBindingAuthnRequest: true
    postBindingLogout: true
    postBindingResponse: true
    realmRef:
      name: "dev"
      policy:
        resolve: Always
    singleLogoutServiceUrl: https://domain.com/adfs/ls/?wa=wsignout1.0
    singleSignOnServiceUrl: https://domain.com/adfs/ls/
    storeToken: false
    trustEmail: true
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Identity Provider Mapper

Use mappers to transform claims or assertions from the external identity provider into Keycloak user attributes.

```yaml
apiVersion: identityprovider.keycloak.crossplane.io/v1alpha1
kind: IdentityProviderMapper
metadata:
  name: oidc-identity-provider-mapper
spec:
  deletionPolicy: Delete
  forProvider:
    extraConfig:
      Claim: my-email-claim
      UserAttribute: email
      syncMode: INHERIT
    identityProviderAlias: my-idp
    identityProviderMapper: '%s-user-attribute-idp-mapper'
    name: email-attribute-importer
    realmRef:
      name: "dev"
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Kubernetes Identity Provider

Use this resource when Kubernetes-issued tokens should be accepted as an external identity source.

```yaml
apiVersion: identityprovider.keycloak.crossplane.io/v1alpha1
kind: KubernetesIdentityProvider
metadata:
  name: k8s-federated-idp
spec:
  deletionPolicy: Delete
  forProvider:
    alias: k8s-federated
    realmRef:
      name: "orgs"
      policy:
        resolve: Always
    organizationIdRef:
      name: example
      policy:
        resolve: Always
    issuer: https://kubernetes.default.svc
    trustEmail: true
    syncMode: FORCE
    enabled: true
  providerConfigRef:
    name: "keycloak-provider-config"
```

### OpenShift V4 Identity Provider

Use this resource to federate with OpenShift 4 clusters through the provider's purpose-built OIDC integration.

```yaml
apiVersion: identityprovider.keycloak.crossplane.io/v1alpha1
kind: OidcOpenShiftV4IdentityProvider
metadata:
  name: openshift-v4-identity-provider
spec:
  deletionPolicy: Delete
  forProvider:
    alias: openshift-v4
    baseUrl: https://openshift.example.com:8443
    clientId: openshift-client
    clientSecretSecretRef:
      key: client-secret
      name: client-secret
      namespace: dev
    defaultScopes: user:full
    realmRef:
      name: "dev"
      policy:
        resolve: Always
    syncMode: IMPORT
    trustEmail: true
  providerConfigRef:
    name: "keycloak-provider-config"
```

### SPIFFE Identity Provider

Use this resource with Keycloak 26.5+ when workload identities should be validated through a SPIFFE trust domain and bundle endpoint.

```yaml
apiVersion: identityprovider.keycloak.crossplane.io/v1alpha1
kind: SpiffeIdentityProvider
metadata:
  name: spiffe-identity-provider
spec:
  deletionPolicy: Delete
  forProvider:
    alias: spiffe-idp
    bundleEndpoint: https://example.com/spiffe/bundle
    realmRef:
      name: "dev"
      policy:
        resolve: Always
    trustDomain: spiffe://test-domain.example
  providerConfigRef:
    name: "keycloak-provider-config"
```

### Provider Token Exchange Scope Permission

Use token exchange scope permissions when selected clients are allowed to exchange tokens against an external identity provider.

```yaml
apiVersion: identityprovider.keycloak.crossplane.io/v1alpha1
kind: ProviderTokenExchangeScopePermission
metadata:
  name: token-exchange-permission
spec:
  deletionPolicy: Delete
  forProvider:
    clientsRefs:
      - name: token-exchange-test-client
        policy:
          resolve: Always
    policyType: client
    providerAliasRef:
      name: token-exchange-test-idp
      policy:
        resolve: Always
    realmIdRef:
      name: dev
      policy:
        resolve: Always
  providerConfigRef:
    name: "keycloak-provider-config"
```

## Key fields

### Common identity provider fields

| Field | Applies to | Why it matters |
|-------|------------|----------------|
| `alias` | Most provider CRDs | Unique Keycloak alias used by login screens, mappers, and token exchange rules. |
| `realmRef` / `realmIdRef` | All CRDs on this page | Selects the realm that owns the provider or permission. |
| `providerConfigRef` | All resources | Points at the Crossplane provider configuration used to talk to Keycloak. |
| `enabled` | Provider CRDs | Controls whether the provider is active for authentication. |
| `syncMode` | OIDC, Google, Kubernetes, OpenShift, mappers | Decides how external identities are synchronized into Keycloak users. |
| `trustEmail` | Google, SAML, Kubernetes, OpenShift | Marks externally supplied email addresses as trusted. |
| `organizationIdRef` | OIDC org binding, Kubernetes | Binds the provider to a Keycloak organization. |

### Protocol-specific fields

| Field | Resource | Why it matters |
|-------|----------|----------------|
| `authorizationUrl` | OIDC `IdentityProvider` | Authorization endpoint for the external OIDC provider. |
| `tokenUrl` | OIDC `IdentityProvider` | Token endpoint used by Keycloak to exchange authorization codes. |
| `clientIdSecretRef` / `clientSecretSecretRef` | OIDC, Google, OpenShift | Reads client credentials from Kubernetes secrets instead of embedding them in manifests. |
| `entityId` | SAML `IdentityProvider` | Declares the remote SAML IdP entity identifier. |
| `singleSignOnServiceUrl` | SAML `IdentityProvider` | Remote SAML SSO entrypoint. |
| `singleLogoutServiceUrl` | SAML `IdentityProvider` | Remote SAML logout endpoint. |
| `identityProviderAlias` | `IdentityProviderMapper` | Attaches the mapper to a specific provider alias. |
| `extraConfig` | OIDC providers and mappers | Holds provider- or mapper-specific settings such as client auth method or claim mapping. |
| `issuer` | `KubernetesIdentityProvider` | Expected token issuer for Kubernetes service account tokens. |
| `baseUrl` | `OidcOpenShiftV4IdentityProvider` | Base URL for the OpenShift cluster's OIDC endpoints. |
| `bundleEndpoint` | `SpiffeIdentityProvider` | URL that publishes the SPIFFE bundle used for trust validation. |
| `trustDomain` | `SpiffeIdentityProvider` | SPIFFE trust domain accepted by the identity provider. |
| `providerAliasRef` | `ProviderTokenExchangeScopePermission` | Refers to the external provider whose tokens may be exchanged. |
| `clientsRefs` | `ProviderTokenExchangeScopePermission` | Lists the Keycloak clients that receive token exchange permission. |

## Related Resources

- **[Organizations](./organizations.md)** — Bind identity providers to Keycloak organizations.
- **[Authentication Flows](./authentication-flows.md)** — Redirect users through identity providers as part of custom login flows.
- **[Users](./users.md)** — Understand how federated identities map to Keycloak user records.


