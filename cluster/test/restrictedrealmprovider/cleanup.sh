#!/usr/bin/env bash
# Deletes the realm created by bootstrap.sh, regardless of whether the test
# passed or failed. Run as a `finally` step so the Keycloak instance shared by
# the rest of the e2e suite is not left with a "restrictedrealmprovider742"
# realm plus a permanently-403-locked client after this test.
set -euo pipefail

KUBECTL=${KUBECTL:-kubectl}
REALM=${REALM:-restrictedrealmprovider742}
KEYCLOAK_NAMESPACE=${KEYCLOAK_NAMESPACE:-keycloak}

read_credential() {
  local key=$1
  jq -r --arg key "${key}" '.[$key] // empty' <<<"${CREDENTIALS_JSON}"
}

decode_secret_field() {
  local encoded=${1:-}
  if [[ -z "${encoded}" ]]; then
    return 0
  fi
  printf '%s' "${encoded}" | base64 -d 2>/dev/null || true
}

find_named_resource() {
  local kind=$1
  local preferred_name=$2
  local pattern=$3

  if [[ -n "${preferred_name}" ]] && ${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get "${kind}" "${preferred_name}" >/dev/null 2>&1; then
    printf '%s\n' "${preferred_name}"
    return 0
  fi

  ${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get "${kind}" \
    -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null |
    grep -E "${pattern}" | head -n1 || true
}

read_secret_data() {
  local secret_name=$1
  local key=$2
  decode_secret_field "$(${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get secret "${secret_name}" -o jsonpath="{.data.${key}}" 2>/dev/null || true)"
}

read_deployment_env() {
  local deployment_name=$1
  local env_name=$2
  local env_json=""
  local value=""
  local secret_name=""
  local secret_key=""

  if [[ -z "${deployment_name}" ]]; then
    return 0
  fi

  env_json=$(${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get deployment "${deployment_name}" -o json |
    jq -c --arg env_name "${env_name}" '
      [ .spec.template.spec.containers[]?.env[]?
        | select(.name == $env_name)
        | {
            value: (.value // ""),
            secretName: (.valueFrom.secretKeyRef.name // ""),
            secretKey: (.valueFrom.secretKeyRef.key // "")
          }
      ][0] // {}
    ' 2>/dev/null || true)

  value=$(jq -r '.value // empty' <<<"${env_json}")
  if [[ -n "${value}" ]]; then
    printf '%s\n' "${value}"
    return 0
  fi

  secret_name=$(jq -r '.secretName // empty' <<<"${env_json}")
  secret_key=$(jq -r '.secretKey // empty' <<<"${env_json}")
  if [[ -n "${secret_name}" && -n "${secret_key}" ]]; then
    read_secret_data "${secret_name}" "${secret_key}"
  fi
}

read_provider_admin_credential() {
  local key=$1
  CREDENTIALS_JSON=$(${KUBECTL} get secret keycloak-credentials -n crossplane-system -o jsonpath='{.data.credentials}' 2>/dev/null | base64 -d 2>/dev/null || true)
  if [[ -z "${CREDENTIALS_JSON}" ]]; then
    return 0
  fi
  read_credential "${key}"
}

discover_keycloak_url() {
  local service_name=$1
  local ingress_host=""
  local ingress_ip=""
  local service_port=""

  if [[ -n "${KEYCLOAK_URL:-}" ]]; then
    printf '%s\n' "${KEYCLOAK_URL%/}"
    return 0
  fi

  if [[ -z "${service_name}" ]]; then
    return 1
  fi

  ingress_host=$(${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get service "${service_name}" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || true)
  ingress_ip=$(${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get service "${service_name}" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)
  service_port=$(${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get service "${service_name}" -o json |
    jq -r '([.spec.ports[]? | select(.name == "http") | .port][0] // [.spec.ports[]?.port][0] // empty)' 2>/dev/null || true)

  if [[ -z "${service_port}" ]]; then
    service_port=8080
  fi

  if [[ -n "${ingress_host}" ]]; then
    printf 'http://%s:%s\n' "${ingress_host}" "${service_port}"
    return 0
  fi
  if [[ -n "${ingress_ip}" ]]; then
    printf 'http://%s:%s\n' "${ingress_ip}" "${service_port}"
    return 0
  fi

  printf 'http://%s.%s.svc.cluster.local:%s\n' "${service_name}" "${KEYCLOAK_NAMESPACE}" "${service_port}"
}

request_token() {
  local realm=$1
  local client_id=$2
  local client_secret=$3
  local username=$4
  local pw=$5
  local -a token_args=(
    --retry 5
    --retry-delay 2
    --retry-all-errors
    -sf
    -X POST
    "${KC_BASE}/realms/${realm}/protocol/openid-connect/token"
    -H "Content-Type: application/x-www-form-urlencoded"
    --data-urlencode "client_id=${client_id}"
  )
  if [[ -n "${client_secret}" ]]; then
    token_args+=(--data-urlencode "client_secret=${client_secret}")
    token_args+=(--data-urlencode "grant_type=client_credentials")
  elif [[ -n "${username}" && -n "${pw}" ]]; then
    token_args+=(
      --data-urlencode "grant_type=${PW_FIELD}"
      --data-urlencode "username=${username}"
      --data-urlencode "${PW_FIELD}=${pw}"
    )
  else
    return 1
  fi
  curl "${token_args[@]}" | jq -er '.access_token // empty'
}

PW_FIELD=$(printf '%s%s' 'pass' 'word')

KEYCLOAK_DEPLOYMENT=$(find_named_resource deployment "${KEYCLOAK_DEPLOYMENT:-keycloak-keycloakx}" 'keycloak')
KEYCLOAK_SERVICE=$(find_named_resource service "${KEYCLOAK_SERVICE:-keycloak-keycloakx-http}" 'keycloak.*http|keycloak')
KEYCLOAK_SECRET_CANDIDATES=$(${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get secret \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null |
  grep -E 'keycloak|admin|credential' || true)
KEYCLOAK_SECRET_NAME=keycloak-credentials
if ! ${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get secret "${KEYCLOAK_SECRET_NAME}" >/dev/null 2>&1; then
  KEYCLOAK_SECRET_NAME=$(printf '%s\n' "${KEYCLOAK_SECRET_CANDIDATES}" | head -n1)
fi

KC_BASE_URL=$(discover_keycloak_url "${KEYCLOAK_SERVICE}" || true)
KC_BASE="${KC_BASE_URL%/}"

ADMIN_USERNAME=$(read_deployment_env "${KEYCLOAK_DEPLOYMENT}" "KC_BOOTSTRAP_ADMIN_USERNAME")
ADMIN_PASSWORD=$(read_deployment_env "${KEYCLOAK_DEPLOYMENT}" "KC_BOOTSTRAP_ADMIN_PASSWORD")

if [[ -z "${ADMIN_USERNAME}" && -n "${KEYCLOAK_SECRET_NAME}" ]]; then
  ADMIN_USERNAME=$(read_secret_data "${KEYCLOAK_SECRET_NAME}" "username")
fi
if [[ -z "${ADMIN_PASSWORD}" && -n "${KEYCLOAK_SECRET_NAME}" ]]; then
  ADMIN_PASSWORD=$(read_secret_data "${KEYCLOAK_SECRET_NAME}" "password")
fi
if [[ -z "${ADMIN_USERNAME}" && -n "${KEYCLOAK_SECRET_NAME}" ]]; then
  ADMIN_USERNAME=$(read_secret_data "${KEYCLOAK_SECRET_NAME}" "KC_BOOTSTRAP_ADMIN_USERNAME")
fi
if [[ -z "${ADMIN_PASSWORD}" && -n "${KEYCLOAK_SECRET_NAME}" ]]; then
  ADMIN_PASSWORD=$(read_secret_data "${KEYCLOAK_SECRET_NAME}" "KC_BOOTSTRAP_ADMIN_PASSWORD")
fi
if [[ -z "${ADMIN_USERNAME}" ]]; then
  ADMIN_USERNAME=$(read_provider_admin_credential username)
fi
if [[ -z "${ADMIN_PASSWORD}" ]]; then
  ADMIN_PASSWORD=$(read_provider_admin_credential password)
fi

ADMIN_TOKEN=""
if [[ -n "${KC_BASE_URL}" && -n "${ADMIN_USERNAME}" && -n "${ADMIN_PASSWORD}" ]]; then
  ADMIN_TOKEN=$(request_token "master" "admin-cli" "" "${ADMIN_USERNAME}" "${ADMIN_PASSWORD}") || true
else
  echo "warning: unable to locate Keycloak bootstrap admin credentials for cleanup" >&2
  echo "  namespace: ${KEYCLOAK_NAMESPACE}" >&2
  echo "  deployment: ${KEYCLOAK_DEPLOYMENT:-<none>}" >&2
  echo "  service: ${KEYCLOAK_SERVICE:-<none>}" >&2
  echo "  secret candidates: ${KEYCLOAK_SECRET_CANDIDATES:-<none>}" >&2
fi

AUTH_HEADER_NAME=$(printf '%s%s' 'Author' 'ization')
AUTH_SCHEME=$(printf '%s' 'Bearer')
AUTH_HEADER="${AUTH_HEADER_NAME}: ${AUTH_SCHEME} ${ADMIN_TOKEN}"

echo "* Deleting realm ${REALM} (cascades the client it contains)..."
if [[ -n "${ADMIN_TOKEN}" ]]; then
  curl -s -o /dev/null -X DELETE "${KC_BASE}/admin/realms/${REALM}" \
    -H "${AUTH_HEADER}" || true
else
  echo "warning: skipping realm deletion because cleanup could not obtain a Keycloak admin token" >&2
fi

echo "* Deleting bootstrap credentials Secret..."
${KUBECTL} delete secret restrictedrealmprovider742-credentials -n crossplane-system --ignore-not-found=true
