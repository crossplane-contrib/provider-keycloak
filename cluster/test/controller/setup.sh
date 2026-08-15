#!/usr/bin/env bash
# Setup script for controller-focused chainsaw e2e tests.
#
# This runs in the same kind cluster as the provider (fenrir-1) after the
# regular uptest setup has already applied the init manifests and the provider
# is running. It installs the infrastructure that the controller tests probe:
#
#   * Envoy Gateway (or Envoy Proxy via Helm) for validating that the
#     connection secret produced by the connectionsecrettransform controller
#     really works as an Envoy OIDC SecurityPolicy source.
#
#   * Traefik with the forward-auth middleware plugin as a second OIDC
#     integration point (different authentication flow, same secret format).
#
# Both are deployed as lightweight in-cluster services; no external load
# balancer is needed. The chainsaw tests assert readiness of these services
# and then configure them to authenticate via the Keycloak realm and client
# set up by the test itself.
#
# Usage:
#   KUBECTL=kubectl KEYCLOAK_URL=http://... NAMESPACE=dev ./cluster/test/controller/setup.sh
#
# Required environment:
#   KUBECTL        - path to kubectl (default: kubectl)
#   KEYCLOAK_URL   - base URL of the Keycloak instance (default: http://keycloak.keycloak.svc.cluster.local)
#   NAMESPACE      - namespace the test resources live in (default: dev)
#
# Optional (for customising images/versions):
#   ENVOY_GATEWAY_VERSION    - Envoy Gateway Helm chart version (default: v1.4.1)
#   TRAEFIK_VERSION          - Traefik Helm chart version (default: 34.3.0)
set -euo pipefail

KUBECTL="${KUBECTL:-kubectl}"
KEYCLOAK_URL="${KEYCLOAK_URL:-http://keycloak.keycloak.svc.cluster.local}"
NAMESPACE="${NAMESPACE:-dev}"

ENVOY_GATEWAY_VERSION="${ENVOY_GATEWAY_VERSION:-v1.4.1}"
TRAEFIK_VERSION="${TRAEFIK_VERSION:-34.3.0}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

info()  { echo "==> $*"; }
fatal() { echo "ERROR: $*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Helm
# ---------------------------------------------------------------------------
require_helm() {
  if ! command -v helm &>/dev/null; then
    info "Installing Helm..."
    curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  fi
}

# ---------------------------------------------------------------------------
# Envoy Gateway
# ---------------------------------------------------------------------------
install_envoy_gateway() {
  info "Installing Envoy Gateway ${ENVOY_GATEWAY_VERSION}..."
  helm upgrade --install eg oci://docker.io/envoyproxy/gateway-helm \
    --version "${ENVOY_GATEWAY_VERSION}" \
    --namespace envoy-gateway-system \
    --create-namespace \
    --wait --timeout=3m \
    --set deployment.envoyGateway.image.tag="${ENVOY_GATEWAY_VERSION}" \
    || fatal "Envoy Gateway installation failed"

  info "Waiting for Envoy Gateway controller pod..."
  "${KUBECTL}" wait pod \
    --namespace envoy-gateway-system \
    --selector app.kubernetes.io/name=gateway \
    --for=condition=Ready \
    --timeout=3m
  info "Envoy Gateway ready."
}

# ---------------------------------------------------------------------------
# Traefik
# ---------------------------------------------------------------------------
install_traefik() {
  info "Installing Traefik ${TRAEFIK_VERSION}..."
  helm repo add traefik https://traefik.github.io/charts 2>/dev/null || true
  helm repo update traefik
  helm upgrade --install traefik traefik/traefik \
    --version "${TRAEFIK_VERSION}" \
    --namespace traefik \
    --create-namespace \
    --wait --timeout=3m \
    --set service.type=ClusterIP \
    --set ports.web.expose.default=true \
    || fatal "Traefik installation failed"

  info "Waiting for Traefik pod..."
  "${KUBECTL}" wait pod \
    --namespace traefik \
    --selector app.kubernetes.io/name=traefik \
    --for=condition=Ready \
    --timeout=3m
  info "Traefik ready."
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
require_helm
install_envoy_gateway
install_traefik

info "Controller e2e setup complete."
