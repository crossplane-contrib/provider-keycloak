#!/usr/bin/env bash
# Posts a ConversionReview (stdin) to the provider's live /convert endpoint
# through the API server's service proxy and prints the response JSON. The
# service coordinates and caBundle are injected into the CRD by the Crossplane
# package manager, so reading them from the CRD also proves that wiring.
set -euo pipefail

KUBECTL=${KUBECTL:-kubectl}
crd=$1

ns=$(${KUBECTL} get crd "${crd}" -o jsonpath='{.spec.conversion.webhook.clientConfig.service.namespace}')
svc=$(${KUBECTL} get crd "${crd}" -o jsonpath='{.spec.conversion.webhook.clientConfig.service.name}')
port=$(${KUBECTL} get crd "${crd}" -o jsonpath='{.spec.conversion.webhook.clientConfig.service.port}')

if [[ -z "${ns}" || -z "${svc}" || -z "${port}" ]]; then
  echo "CRD ${crd} has no conversion webhook service configured" >&2
  exit 1
fi

${KUBECTL} create --raw "/api/v1/namespaces/${ns}/services/https:${svc}:${port}/proxy/convert" -f -
