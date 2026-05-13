# SSO with ArgoCD — Example Manifests

This directory contains all the Keycloak CRDs and ArgoCD configuration needed
to set up single sign-on for ArgoCD using Keycloak, managed by provider-keycloak.

## Prerequisites

- Keycloak running and reachable by the provider
- A `ProviderConfig` named `keycloak-provider-config`
- ArgoCD installed in the `argocd` namespace

## Apply in order

```bash
# Keycloak resources (via Crossplane)
kubectl apply -f 01-realm.yaml
kubectl apply -f 02-client.yaml
kubectl apply -f 03-roles.yaml
kubectl apply -f 04-protocol-mapper.yaml
kubectl apply -f 05-groups.yaml
kubectl apply -f 06-users.yaml
kubectl apply -f 07-memberships.yaml

# Wait for the client secret to be created
kubectl wait client.openidclient.keycloak.crossplane.io/argocd-client \
  --for=condition=Ready --timeout=120s

# Extract client secret and store in ArgoCD
CLIENT_SECRET=$(kubectl get secret argocd-keycloak-client-secret \
  -n crossplane-system \
  -o jsonpath='{.data.attribute\.client_secret}' | base64 -d)

kubectl -n argocd patch secret argocd-secret --type merge -p \
  "{\"stringData\": {\"oidc.keycloak.clientSecret\": \"${CLIENT_SECRET}\"}}"

# ArgoCD configuration — edit URLs in these files first!
kubectl apply -f 08-argocd-cm.yaml
kubectl apply -f 09-argocd-rbac-cm.yaml

# Restart ArgoCD to pick up changes
kubectl -n argocd rollout restart deployment argocd-server
```

## Test users

| User | Password | Role |
|------|----------|------|
| admin-user | changeme | argocd-admin → ArgoCD admin |
| viewer-user | changeme | argocd-readonly → ArgoCD readonly |

See the full guide: [SSO with ArgoCD](../../docs/docs/guides/sso-with-argocd.md)
