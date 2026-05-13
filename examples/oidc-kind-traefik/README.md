# End-to-End OIDC Demo on kind with Traefik OIDC Plugin

This example deploys a complete OIDC-protected application on a local kind cluster using:

- **Keycloak** as the Identity Provider
- **Crossplane + provider-keycloak** to manage Keycloak configuration declaratively
- **Traefik** with the [`traefik-oidc-auth`](https://github.com/sevensolutions/traefik-oidc-auth) plugin for authentication and role-based authorization
- **nginx** as the protected backend

## Architecture

```
                        ┌──────────────┐
                        │   Keycloak   │
                        │   (OIDC IdP) │
                        └──────┬───────┘
                               │ id_token
┌────────┐    HTTP    ┌────────▼───────┐    proxy     ┌───────────┐
│  User  │───────────►│    Traefik     │─────────────►│   nginx   │
└────────┘            │ (OIDC plugin)  │              │ (backend) │
                      └────────────────┘              └───────────┘
                               ▲
                        manages│
                      ┌────────┴───────┐
                      │  provider-     │
                      │  keycloak      │
                      └────────────────┘
```

## Quick Start

```bash
./setup.sh
```

The script will:
1. Create a kind cluster with port mappings
2. Deploy Keycloak (accessible at http://localhost:9090)
3. Install Crossplane and provider-keycloak
4. Create a realm, confidential client, roles, users, and groups via CRDs
5. Install Traefik with the OIDC plugin
6. Deploy nginx protected by the OIDC middleware

## Testing

After the script completes, open http://localhost:8080 in your browser.

| User  | Password | Role           | Result         |
|-------|----------|----------------|----------------|
| alice | password | allowed-role   | ✅ Access granted |
| bob   | password | forbidden-role | ❌ Access denied  |

## Files

| File | Description |
|------|-------------|
| `kind-config.yaml` | kind cluster configuration with port mappings |
| `keycloak.yaml` | Keycloak deployment and NodePort service |
| `provider.yaml` | provider-keycloak Provider resource |
| `provider-config.yaml` | ProviderConfig + credentials Secret |
| `realm-client.yaml` | Realm and confidential Client CRDs |
| `roles.yaml` | Realm roles (allowed-role, forbidden-role) |
| `role-mapper.yaml` | Protocol mapper to expose roles in JWT |
| `users.yaml` | Test users (alice, bob) |
| `groups.yaml` | Groups and role assignments |
| `memberships.yaml` | User-to-group memberships |
| `nginx.yaml` | nginx backend deployment |
| `traefik-values.yaml` | Traefik Helm values with OIDC plugin enabled |
| `middleware-ingress.yaml` | OIDC middleware + IngressRoute |
| `setup.sh` | Automated setup script |

## Cleanup

```bash
kind delete cluster --name oidc-demo
```
