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

# A ConversionReview POST is idempotent, so transport-level failures (webhook
# pod still rolling out after CI setup, endpoints not yet updated) are retried.
# Conversion failures are HTTP 200 responses and never retried.
payload=$(cat)
attempts=24
for i in $(seq 1 "${attempts}"); do
  if out=$(printf '%s' "${payload}" | ${KUBECTL} create --raw "/api/v1/namespaces/${ns}/services/https:${svc}:${port}/proxy/convert" -f - 2>/tmp/convert-err.txt); then
    printf '%s\n' "${out}"
    exit 0
  fi
  echo "attempt ${i}/${attempts}: webhook not reachable yet: $(cat /tmp/convert-err.txt)" >&2
  sleep 5
done
echo "conversion webhook unreachable after ${attempts} attempts" >&2
exit 1
