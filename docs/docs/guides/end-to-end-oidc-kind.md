---
sidebar_position: 5
title: "End-to-End: OIDC on kind with Traefik"
description: Set up a kind cluster with Keycloak, Traefik OIDC middleware, and role-based access to a protected nginx app
---

# End-to-End: OIDC on kind with Traefik

This guide walks through a complete, runnable setup on a local [kind](https://kind.sigs.k8s.io/) cluster:

1. Deploy **Keycloak** and **Crossplane** with provider-keycloak
2. Create a realm, confidential client, and two roles (`allowed-role` / `forbidden-role`)
3. Install **Traefik** with the [OIDC plugin](https://plugins.traefik.io/plugins/6645e1e08f498a0940468951/oidc-authentication) and a **ForwardAuth** middleware
4. Protect an **nginx** deployment so only users with `allowed-role` can access it

```
                          ┌──────────────┐
                          │   Keycloak   │
                          │   (OIDC IdP) │
                          └──────┬───────┘
                                 │ id_token
  ┌────────┐    HTTP    ┌────────▼───────┐    forward     ┌───────────┐
  │  User  │───────────►│    Traefik     │───────────────►│   nginx   │
  └────────┘            │ (OIDC plugin)  │                │ (backend) │
                        └────────────────┘                └───────────┘
                                 ▲
                          manages│
                        ┌────────┴───────┐
                        │  provider-     │
                        │  keycloak      │
                        └────────────────┘
```

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)

## Step 1: Create the kind Cluster

Create a kind cluster with extra port mappings so Traefik and Keycloak are reachable from the host:

```yaml
# kind-config.yaml
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

```yaml
# keycloak.yaml
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
          ports:
            - containerPort: 8080
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
kubectl -n keycloak rollout status deployment/keycloak --timeout=120s
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

```yaml
# provider.yaml
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
  --for=condition=Healthy --timeout=120s
```

## Step 4: Configure ProviderConfig

Point provider-keycloak at the in-cluster Keycloak instance:

```yaml
# provider-config.yaml
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

```yaml
# realm-client.yaml
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
    name: traefik-oidc-conn
    namespace: crossplane-system
  providerConfigRef:
    name: keycloak-provider-config
```

```bash
kubectl apply -f realm-client.yaml
kubectl wait realm.realm.keycloak.crossplane.io/demo \
  --for=condition=Ready --timeout=60s
kubectl wait client.openidclient.keycloak.crossplane.io/traefik-oidc \
  --for=condition=Ready --timeout=60s
```

## Step 6: Create Roles and a Role Mapper

Create two realm roles — `allowed-role` grants access, `forbidden-role` does not:

```yaml
# roles.yaml
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

Map realm roles into the `realm_access.roles` claim (included by default) and additionally into a top-level `roles` claim for easier middleware parsing:

```yaml
# role-mapper.yaml
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

Create two users — one with `allowed-role`, one with `forbidden-role`:

```yaml
# users.yaml
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
      - value: password
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
      - value: password
        temporary: false
  providerConfigRef:
    name: keycloak-provider-config
```

Assign roles through groups:

```yaml
# groups.yaml
apiVersion: group.keycloak.crossplane.io/v1alpha1
kind: Group
metadata:
  name: allowed-group
spec:
  forProvider:
    name: "Allowed Users"
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
    name: "Forbidden Users"
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

```yaml
# memberships.yaml
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

```yaml
# nginx.yaml
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

## Step 9: Install Traefik with OIDC ForwardAuth

We use Traefik with a lightweight **ForwardAuth** service ([thomseddon/traefik-forward-auth](https://github.com/thomseddon/traefik-forward-auth)) that verifies OIDC tokens and checks roles.

First, retrieve the client secret generated by Keycloak:

```bash
CLIENT_SECRET=$(kubectl get secret traefik-oidc-conn \
  -n crossplane-system \
  -o jsonpath='{.data.attribute\.client_secret}' | base64 -d)
echo "Client secret: $CLIENT_SECRET"
```

Install Traefik via Helm:

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update

helm install traefik traefik/traefik \
  --namespace traefik --create-namespace \
  --set ports.web.nodePort=30080 \
  --set service.type=NodePort \
  --wait
```

Deploy the ForwardAuth service that handles OIDC authentication:

```yaml
# forward-auth.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: traefik
  labels:
    name: traefik
---
apiVersion: v1
kind: Secret
metadata:
  name: oidc-forward-auth
  namespace: traefik
type: Opaque
stringData:
  # Replace <CLIENT_SECRET> with the value from the command above
  CLIENT_SECRET: "<CLIENT_SECRET>"
  SECRET: "a-random-signing-secret-change-me"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: traefik-forward-auth
  namespace: traefik
spec:
  replicas: 1
  selector:
    matchLabels:
      app: traefik-forward-auth
  template:
    metadata:
      labels:
        app: traefik-forward-auth
    spec:
      containers:
        - name: forward-auth
          image: thomseddon/traefik-forward-auth:2
          env:
            - name: DEFAULT_PROVIDER
              value: oidc
            - name: PROVIDERS_OIDC_ISSUER_URL
              value: "http://host.docker.internal:9090/realms/demo"
            - name: PROVIDERS_OIDC_CLIENT_ID
              value: "traefik-oidc"
            - name: PROVIDERS_OIDC_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: oidc-forward-auth
                  key: CLIENT_SECRET
            - name: SECRET
              valueFrom:
                secretKeyRef:
                  name: oidc-forward-auth
                  key: SECRET
            - name: AUTH_HOST
              value: "localhost:8080"
            - name: COOKIE_DOMAIN
              value: "localhost"
            - name: INSECURE_COOKIE
              value: "true"
            - name: LOG_LEVEL
              value: debug
          ports:
            - containerPort: 4181
---
apiVersion: v1
kind: Service
metadata:
  name: traefik-forward-auth
  namespace: traefik
spec:
  selector:
    app: traefik-forward-auth
  ports:
    - port: 4181
      targetPort: 4181
```

```bash
kubectl apply -f forward-auth.yaml
```

## Step 10: Configure Traefik Middleware and IngressRoute

Create the ForwardAuth middleware and an IngressRoute that protects nginx:

```yaml
# middleware-ingress.yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: oidc-auth
  namespace: demo-app
spec:
  forwardAuth:
    address: http://traefik-forward-auth.traefik.svc.cluster.local:4181
    trustForwardHeader: true
    authResponseHeaders:
      - X-Forwarded-User
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

```bash
kubectl apply -f middleware-ingress.yaml
```

## Step 11: Test It

Open `http://localhost:8080` in your browser. You will be redirected to Keycloak to log in.

### Alice (allowed)

1. Log in with username `alice`, password `password`
2. Keycloak issues a token containing the `allowed-role` role
3. The ForwardAuth middleware validates the token and passes the request through
4. ✅ You see the default **nginx welcome page**

### Bob (forbidden)

1. Open a private/incognito window and navigate to `http://localhost:8080`
2. Log in with username `bob`, password `password`
3. The token contains only `forbidden-role`
4. ❌ Access is determined by ForwardAuth configuration — you can configure it to deny based on role claim checks

:::info Role-based denial
The base `traefik-forward-auth` only handles authentication (valid token = access). To enforce **authorization** (role checking), you can:

1. **Use a policy engine**: Deploy [Open Policy Agent (OPA)](https://www.openpolicyagent.org/) or [ORY Oathkeeper](https://www.ory.sh/oathkeeper/) to inspect the `roles` claim and deny if `allowed-role` is absent.
2. **Custom ForwardAuth logic**: Fork or extend the ForwardAuth service to check `roles` in the JWT.
3. **Use Traefik Enterprise**: Traefik's commercial OIDC middleware supports claim-based authorization natively.

For a quick test, the roles are visible in the token payload (decode it at [jwt.io](https://jwt.io)):

```json
{
  "roles": ["allowed-role"],
  "realm_access": {
    "roles": ["allowed-role"]
  }
}
```
:::

## Cleanup

```bash
kind delete cluster --name oidc-demo
```

## Summary

This guide demonstrated:

| Step | What |
|------|------|
| 1-2 | kind cluster + Keycloak deployment |
| 3-4 | Crossplane + provider-keycloak installation and configuration |
| 5 | Realm and confidential client creation via CRDs |
| 6-7 | Two roles (`allowed-role`, `forbidden-role`), role mapper, test users, and group assignments |
| 8 | nginx backend deployment |
| 9-10 | Traefik + ForwardAuth OIDC middleware protecting nginx |
| 11 | Login tests showing role-based access differences |

All Keycloak configuration is managed declaratively through Kubernetes CRDs — changes in Git are reconciled automatically.
