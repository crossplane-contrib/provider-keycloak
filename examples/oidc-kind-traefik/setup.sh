#!/usr/bin/env bash
# End-to-end OIDC demo on kind with Traefik OIDC plugin and provider-keycloak
#
# Prerequisites: docker, kind, kubectl, helm
# Usage: ./setup.sh
#
# This script:
#   1. Creates a kind cluster with port mappings
#   2. Deploys Keycloak
#   3. Installs Crossplane + provider-keycloak
#   4. Configures ProviderConfig, Realm, Client, Roles, Users, Groups
#   5. Installs Traefik with the OIDC plugin
#   6. Deploys nginx protected by the OIDC middleware
#
# After completion:
#   - Open http://localhost:8080 → redirected to Keycloak login
#   - Alice (password: password) → allowed (has allowed-role)
#   - Bob   (password: password) → denied  (has only forbidden-role)
#
# Cleanup: kind delete cluster --name oidc-demo

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLUSTER_NAME="oidc-demo"

info() { echo "▶ $*"; }
wait_for() { kubectl wait "$@" --timeout=180s; }

# ─── Step 1: Create kind cluster ────────────────────────────────────────────
info "Creating kind cluster '${CLUSTER_NAME}'..."
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  info "Cluster already exists, skipping creation"
else
  kind create cluster --name "${CLUSTER_NAME}" --config "${SCRIPT_DIR}/kind-config.yaml"
fi

# ─── Step 2: Deploy Keycloak ────────────────────────────────────────────────
info "Deploying Keycloak..."
kubectl apply -f "${SCRIPT_DIR}/keycloak.yaml"
kubectl -n keycloak rollout status deployment/keycloak --timeout=180s
info "Keycloak available at http://localhost:9090"

# ─── Step 3: Install Crossplane ─────────────────────────────────────────────
info "Installing Crossplane..."
helm repo add crossplane https://charts.crossplane.io/stable 2>/dev/null || true
helm repo update
if ! helm status crossplane -n crossplane-system &>/dev/null; then
  helm install crossplane crossplane/crossplane \
    --namespace crossplane-system --create-namespace --wait
fi
kubectl -n crossplane-system rollout status deployment/crossplane --timeout=120s

# ─── Step 4: Install provider-keycloak ───────────────────────────────────────
info "Installing provider-keycloak..."
kubectl apply -f "${SCRIPT_DIR}/provider.yaml"
info "Waiting for provider to become healthy..."
sleep 10
kubectl wait provider.pkg provider-keycloak --for=condition=Healthy --timeout=180s

# ─── Step 5: Configure ProviderConfig ────────────────────────────────────────
info "Applying ProviderConfig..."
kubectl apply -f "${SCRIPT_DIR}/provider-config.yaml"
sleep 5

# ─── Step 6: Create Realm and Client ────────────────────────────────────────
info "Creating Realm and Client..."
kubectl apply -f "${SCRIPT_DIR}/realm-client.yaml"
info "Waiting for Realm to be ready..."
sleep 10
kubectl wait realm.realm.keycloak.crossplane.io/demo --for=condition=Ready --timeout=120s
info "Waiting for Client to be ready..."
kubectl wait client.openidclient.keycloak.crossplane.io/traefik-oidc --for=condition=Ready --timeout=120s

# ─── Step 7: Create Roles and Role Mapper ────────────────────────────────────
info "Creating Roles and Role Mapper..."
kubectl apply -f "${SCRIPT_DIR}/roles.yaml"
kubectl apply -f "${SCRIPT_DIR}/role-mapper.yaml"
sleep 5

# ─── Step 8: Create Users, Groups, Memberships ──────────────────────────────
info "Creating Users, Groups, and Memberships..."
kubectl apply -f "${SCRIPT_DIR}/users.yaml"
kubectl apply -f "${SCRIPT_DIR}/groups.yaml"
sleep 5
kubectl apply -f "${SCRIPT_DIR}/memberships.yaml"
sleep 5

# ─── Step 9: Deploy nginx backend ───────────────────────────────────────────
info "Deploying nginx backend..."
kubectl apply -f "${SCRIPT_DIR}/nginx.yaml"
kubectl -n demo-app rollout status deployment/nginx --timeout=60s

# ─── Step 10: Install Traefik with OIDC plugin ──────────────────────────────
info "Installing Traefik with OIDC plugin..."
helm repo add traefik https://traefik.github.io/charts 2>/dev/null || true
helm repo update
if ! helm status traefik -n traefik &>/dev/null; then
  helm install traefik traefik/traefik \
    --namespace traefik --create-namespace \
    --values "${SCRIPT_DIR}/traefik-values.values" \
    --wait
fi

# ─── Step 11: Retrieve client secret and create middleware ───────────────────
info "Retrieving client secret from Crossplane connection secret..."
CLIENT_SECRET=""
for i in $(seq 1 30); do
  CLIENT_SECRET=$(kubectl get secret traefik-oidc-client-secret \
    -n crossplane-system \
    -o jsonpath='{.data.attribute\.client_secret}' 2>/dev/null | base64 -d) || true
  if [ -n "${CLIENT_SECRET}" ]; then
    break
  fi
  info "  Waiting for client secret to be available (attempt $i/30)..."
  sleep 5
done

if [ -z "${CLIENT_SECRET}" ]; then
  echo "ERROR: Could not retrieve client secret after 150s"
  exit 1
fi

info "Client secret retrieved successfully"

# Generate the middleware manifest with the actual secret
sed "s/\${CLIENT_SECRET}/${CLIENT_SECRET}/" "${SCRIPT_DIR}/middleware-ingress.yaml" | kubectl apply -f -

# ─── Done ────────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════════════════"
echo " ✅ Setup complete!"
echo ""
echo " Open http://localhost:8080 in your browser."
echo " You will be redirected to Keycloak for login."
echo ""
echo " Test users:"
echo "   Alice → username: alice, password: password → ✅ ACCESS GRANTED"
echo "          (has 'allowed-role' → passes claim check)"
echo ""
echo "   Bob   → username: bob,   password: password → ❌ ACCESS DENIED"
echo "          (has only 'forbidden-role' → fails claim check)"
echo ""
echo " Keycloak admin: http://localhost:9090 (admin/admin)"
echo ""
echo " Cleanup: kind delete cluster --name ${CLUSTER_NAME}"
echo "═══════════════════════════════════════════════════════════════════"
