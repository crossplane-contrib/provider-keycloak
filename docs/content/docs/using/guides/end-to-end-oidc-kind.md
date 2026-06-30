---
sidebar_position: 5
title: "End-to-End: OIDC on kind with Traefik"
description: Set up a kind cluster with Keycloak, Traefik OIDC plugin, and role-based access to a protected nginx app
---

# End-to-End: OIDC on kind with Traefik

This guide walks through a complete, runnable setup on a local [kind](https://kind.sigs.k8s.io/) cluster:

1. Deploy **Keycloak** and **Crossplane** with provider-keycloak
2. Create a realm, confidential client, and two roles (`allowed-role` / `forbidden-role`)
3. Install **Traefik** with the [`traefik-oidc-auth`](https://github.com/sevensolutions/traefik-oidc-auth) plugin
4. Protect an **nginx** deployment so only users with `allowed-role` can access it

```
                          ┌──────────────┐
                          │   Keycloak   │
                          │   (OIDC IdP) │
                          └──────┬───────┘
                                 │ id_token
  ┌────────┐    HTTP    ┌────────▼───────┐    proxy      ┌───────────┐
  │  User  │───────────►│    Traefik     │──────────────►│   nginx   │
  └────────┘            │ (OIDC plugin)  │               │ (backend) │
                        └────────────────┘               └───────────┘
                                 ▲
                          manages│
                        ┌────────┴───────┐
                        │  provider-     │
                        │  keycloak      │
                        └────────────────┘
```

{{< callout type="info" >}}
**Quick start:** All manifests and an automated setup script are available in [`examples/oidc-kind-traefik/`](https://github.com/crossplane-contrib/provider-keycloak/tree/main/examples/oidc-kind-traefik). Run `./setup.sh` to deploy everything automatically.
{{< /callout >}}

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)

## Step 1: Create the kind Cluster

Create a kind cluster with extra port mappings so Traefik and Keycloak are reachable from the host:

```yaml title="examples/oidc-kind-traefik/kind-config.yaml"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30080
        hostPort: 8080      # Traefik HTTP
      - containerPort: 30443
        hostPort: 8443      # Traefik HTTPS
      - containerPort: 30090
        hostPort: 9090      # Keycloak HTTP
```

```bash
kind create cluster --name oidc-demo --config kind-config.yaml
```

## Step 2: Deploy Keycloak

Install Keycloak with a NodePort so it is accessible from the host:

```yaml title="examples/oidc-kind-traefik/keycloak.yaml"
apiVersion: v1
kind: Namespace
metadata:
  name: keycloak
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: keycloak
  namespace: keycloak
spec:
  replicas: 1
  selector:
    matchLabels:
      app: keycloak
  template:
    metadata:
      labels:
        app: keycloak
    spec:
      containers:
        - name: keycloak
          image: quay.io/keycloak/keycloak:24.0
          args: ["start-dev"]
          env:
            - name: KEYCLOAK_ADMIN
              value: admin
            - name: KEYCLOAK_ADMIN_PASSWORD
              value: admin
            - name: KC_HTTP_PORT
              value: "8080"
            - name: KC_HOSTNAME_STRICT
              value: "false"
            - name: KC_PROXY
              value: "edge"
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet:
              path: /realms/master
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: keycloak
  namespace: keycloak
spec:
  type: NodePort
  selector:
    app: keycloak
  ports:
    - port: 8080
      targetPort: 8080
      nodePort: 30090
```

```bash
kubectl apply -f keycloak.yaml
kubectl -n keycloak rollout status deployment/keycloak --timeout=180s
```

Keycloak is now reachable at `http://localhost:9090`.

## Step 3: Install Crossplane and provider-keycloak

```bash
helm repo add crossplane https://charts.crossplane.io/stable
helm repo update
helm install crossplane crossplane/crossplane \
  --namespace crossplane-system --create-namespace --wait
```

Install provider-keycloak:

```yaml title="examples/oidc-kind-traefik/provider.yaml"
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-keycloak
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-keycloak:v1.9.0
```

```bash
kubectl apply -f provider.yaml
kubectl wait provider.pkg provider-keycloak \
  --for=condition=Healthy --timeout=180s
```

## Step 4: Configure ProviderConfig

Point provider-keycloak at the in-cluster Keycloak instance:

```yaml title="examples/oidc-kind-traefik/provider-config.yaml"
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: |
    {
      "client_id": "admin-cli",
      "username": "admin",
      "password": "admin",
      "url": "http://keycloak.keycloak.svc.cluster.local:8080",
      "realm": "master"
    }
---
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-provider-config
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-credentials
      key: credentials
      namespace: crossplane-system
```

```bash
kubectl apply -f provider-config.yaml
```

## Step 5: Create the Realm and Client

```yaml title="examples/oidc-kind-traefik/realm-client.yaml"
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: demo
spec:
  forProvider:
    realm: demo
    enabled: true
    displayName: "Demo Realm"
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: traefik-oidc
spec:
  forProvider:
    clientId: traefik-oidc
    name: traefik-oidc
    realmId: demo
    accessType: CONFIDENTIAL
    standardFlowEnabled: true
    directAccessGrantsEnabled: true
    validRedirectUris:
      - "http://localhost:8080/*"
    validPostLogoutRedirectUris:
      - "http://localhost:8080/*"
    webOrigins:
      - "http://localhost:8080"
  writeConnectionSecretToRef:
    name: traefik-oidc-client-secret
    namespace: crossplane-system
  providerConfigRef:
    name: keycloak-provider-config
```

```bash
kubectl apply -f realm-client.yaml
kubectl wait realm.realm.keycloak.crossplane.io/demo \
  --for=condition=Ready --timeout=120s
kubectl wait client.openidclient.keycloak.crossplane.io/traefik-oidc \
  --for=condition=Ready --timeout=120s
```

## Step 6: Create Roles and a Role Mapper

Create two realm roles — `allowed-role` grants access, `forbidden-role` does not:

```yaml title="examples/oidc-kind-traefik/roles.yaml"
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: allowed-role
spec:
  forProvider:
    realmId: demo
    name: allowed-role
    description: "Users with this role may access the protected app"
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: forbidden-role
spec:
  forProvider:
    realmId: demo
    name: forbidden-role
    description: "Users with only this role are denied access"
  providerConfigRef:
    name: keycloak-provider-config
```

Map realm roles into a top-level `roles` claim in the JWT:

```yaml title="examples/oidc-kind-traefik/role-mapper.yaml"
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: ProtocolMapper
metadata:
  name: traefik-roles-mapper
spec:
  forProvider:
    clientIdRef:
      name: traefik-oidc
    realmId: demo
    protocol: openid-connect
    protocolMapper: oidc-usermodel-realm-role-mapper
    name: roles
    config:
      "claim.name": "roles"
      "multivalued": "true"
      "id.token.claim": "true"
      "access.token.claim": "true"
      "userinfo.token.claim": "true"
  providerConfigRef:
    name: keycloak-provider-config
```

```bash
kubectl apply -f roles.yaml
kubectl apply -f role-mapper.yaml
```

## Step 7: Create Test Users

Create two users — one with `allowed-role`, one with `forbidden-role`.

First, create password secrets for the users:

```bash
kubectl create secret generic user-alice-password \
  --namespace crossplane-system \
  --from-literal=******
kubectl create secret generic user-bob-password \
  --namespace crossplane-system \
  --from-literal=******
```

Then create the User resources:

```yaml title="examples/oidc-kind-traefik/users.yaml"
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: alice
spec:
  forProvider:
    realmId: demo
    username: alice
    email: alice@example.com
    enabled: true
    initialPassword:
      - valueSecretRef:
          name: user-alice-password
          namespace: crossplane-system
          key: password
        temporary: false
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: user.keycloak.crossplane.io/v1alpha1
kind: User
metadata:
  name: bob
spec:
  forProvider:
    realmId: demo
    username: bob
    email: bob@example.com
    enabled: true
    initialPassword:
      - valueSecretRef:
          name: user-bob-password
          namespace: crossplane-system
          key: password
        temporary: false
  providerConfigRef:
    name: keycloak-provider-config
```

Assign roles through groups:

```yaml title="examples/oidc-kind-traefik/groups.yaml"
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: allowed-group
spec:
  forProvider:
    name: allowed-group
    realmId: demo
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: forbidden-group
spec:
  forProvider:
    name: forbidden-group
    realmId: demo
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: allowed-group-roles
spec:
  forProvider:
    groupIdRef:
      name: allowed-group
    roleIdsRefs:
      - name: allowed-role
    realmId: demo
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Roles
metadata:
  name: forbidden-group-roles
spec:
  forProvider:
    groupIdRef:
      name: forbidden-group
    roleIdsRefs:
      - name: forbidden-role
    realmId: demo
  providerConfigRef:
    name: keycloak-provider-config
```

Assign users to groups declaratively using the `Memberships` CRD:

```yaml title="examples/oidc-kind-traefik/memberships.yaml"
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Memberships
metadata:
  name: allowed-group-members
spec:
  forProvider:
    realmId: demo
    groupIdRef:
      name: allowed-group
    members:
      - alice
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Memberships
metadata:
  name: forbidden-group-members
spec:
  forProvider:
    realmId: demo
    groupIdRef:
      name: forbidden-group
    members:
      - bob
  providerConfigRef:
    name: keycloak-provider-config
```

```bash
kubectl apply -f users.yaml
kubectl apply -f groups.yaml
kubectl apply -f memberships.yaml
```

## Step 8: Deploy nginx (Protected Backend)

```yaml title="examples/oidc-kind-traefik/nginx.yaml"
apiVersion: v1
kind: Namespace
metadata:
  name: demo-app
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: demo-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
  namespace: demo-app
spec:
  selector:
    app: nginx
  ports:
    - port: 80
      targetPort: 80
```

```bash
kubectl apply -f nginx.yaml
```

## Step 9: Install Traefik with OIDC Plugin

The [`traefik-oidc-auth`](https://github.com/sevensolutions/traefik-oidc-auth) plugin handles OIDC authentication **and** claim-based authorization natively — no ForwardAuth sidecar needed.

Enable the plugin in the Traefik Helm values:

```yaml title="examples/oidc-kind-traefik/traefik-values.values"
experimental:
  plugins:
    traefik-oidc-auth:
      moduleName: github.com/sevensolutions/traefik-oidc-auth
      version: v0.19.0

service:
  type: NodePort

ports:
  web:
    nodePort: 30080

providers:
  kubernetesCRD:
    enabled: true
    allowCrossNamespace: true
```

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update

helm install traefik traefik/traefik \
  --namespace traefik --create-namespace \
  --values traefik-values.values \
  --wait
```

## Step 10: Configure OIDC Middleware and IngressRoute

First, retrieve the client secret generated by provider-keycloak:

```bash
CLIENT_SECRET=$(kubectl get secret traefik-oidc-client-secret \
  -n crossplane-system \
  -o jsonpath='{.data.attribute\.client_secret}' | base64 -d)
echo "Client secret: $CLIENT_SECRET"
```

Create the OIDC middleware and IngressRoute. The `Authorization.AssertClaims` block enforces that only tokens with `allowed-role` in the `roles` claim are granted access:

```yaml title="examples/oidc-kind-traefik/middleware-ingress.yaml"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: oidc-auth
  namespace: demo-app
spec:
  plugin:
    traefik-oidc-auth:
      Provider:
        Url: "http://host.docker.internal:9090/realms/demo"
        ClientId: "traefik-oidc"
        ClientSecret: "<CLIENT_SECRET>"   # replaced by setup.sh or manually
      CallbackUri: /oauth2/callback
      Scopes:
        - openid
        - profile
        - email
      Authorization:
        AssertClaims:
          - Name: "roles"
            AnyOf:
              - "allowed-role"
---
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: nginx-protected
  namespace: demo-app
spec:
  entryPoints:
    - web
  routes:
    - match: PathPrefix(`/`)
      kind: Rule
      middlewares:
        - name: oidc-auth
      services:
        - name: nginx
          port: 80
```

Apply the manifest, substituting the actual client secret:

```bash
sed "s/\${CLIENT_SECRET}/${CLIENT_SECRET}/" middleware-ingress.yaml | kubectl apply -f -
```

{{< callout type="info" >}}
The `setup.sh` script automates this substitution. If running manually, replace `<CLIENT_SECRET>` with the value from the previous command.
{{< /callout >}}

## Step 11: Test It

Open `http://localhost:8080` in your browser. You will be redirected to Keycloak to log in.

### Alice (allowed)

1. Log in with username `alice`, password `password`
2. Keycloak issues a token containing the `allowed-role` role
3. The OIDC plugin validates the token and checks the `roles` claim for `allowed-role`
4. ✅ The claim check passes — you see the default **nginx welcome page**

### Bob (forbidden)

1. Open a private/incognito window and navigate to `http://localhost:8080`
2. Log in with username `bob`, password `password`
3. The token contains only `forbidden-role`
4. ❌ The OIDC plugin rejects the request because `allowed-role` is not present in the `roles` claim

{{< callout type="info" >}}
**How the authorization works:**
The `traefik-oidc-auth` plugin's `Authorization.AssertClaims` feature inspects the JWT claims directly. The configuration:

```yaml
Authorization:
  AssertClaims:
    - Name: "roles"
      AnyOf:
        - "allowed-role"
```

requires that the `roles` array in the token contains at least `allowed-role`. This is a built-in feature of the plugin — no external policy engine or ForwardAuth sidecar is needed.

You can verify the token contents at [jwt.io](https://jwt.io):

```json
{
  "roles": ["allowed-role"],
  "realm_access": {
    "roles": ["allowed-role"]
  }
}
```
{{< /callout >}}

## Cleanup

```bash
kind delete cluster --name oidc-demo
```

## Summary

| Step | What |
|------|------|
| 1-2 | kind cluster + Keycloak deployment |
| 3-4 | Crossplane + provider-keycloak installation and configuration |
| 5 | Realm and confidential client creation via CRDs |
| 6-7 | Two roles (`allowed-role`, `forbidden-role`), role mapper, test users, and group assignments |
| 8 | nginx backend deployment |
| 9-10 | Traefik with OIDC plugin middleware protecting nginx (claim-based authorization) |
| 11 | Login tests — Alice (allowed) vs Bob (denied) |

All Keycloak configuration is managed declaratively through Kubernetes CRDs — changes in Git are reconciled automatically.
