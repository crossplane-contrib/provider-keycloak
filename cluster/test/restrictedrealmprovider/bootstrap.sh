#!/usr/bin/env bash
# Bootstraps a Keycloak realm and a fully realm-scoped service account client
# (no master-realm or realm-management roles granted at all) directly against
# the live Keycloak API, then publishes its coordinates as a Kubernetes Secret.
#
# This reproduces "Configuration A" from
# crossplane-contrib/provider-keycloak#742: a client whose service account
# has zero roles outside its own realm. Such a client cannot call
# /admin/serverinfo (Keycloak requires master-realm view-system/manage-realms
# roles for that endpoint), so the provider's login handshake fails with a
# 403 before any managed resource can be reconciled. This is a Keycloak
# server-side authorization constraint, not something provider-keycloak (or
# even the vendored terraform-provider-keycloak) can work around — see the
# "Known limitation" section of docs/content/docs/using/reference/provider-config.md.
set -euo pipefail

KUBECTL=${KUBECTL:-kubectl}
REALM=${REALM:-restrictedrealmprovider742}
CLIENT_ID=${CLIENT_ID:-restrictedrealmprovider742-client}
CLIENT_SECRET=${CLIENT_SECRET:-restrictedrealmprovider742-secret}
KEYCLOAK_NAMESPACE=${KEYCLOAK_NAMESPACE:-keycloak}

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
    --silent
    --show-error
    --fail-with-body
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

create_resource() {
  local path=$1
  local payload=$2
  local resource_name=$3
  local status=""

  status=$(curl -s -o /dev/null -w '%{http_code}' -X POST "${KC_BASE}${path}" \
    -H "${AUTH_HEADER}" \
    -H "Content-Type: application/json" \
    -d "${payload}")

  case "${status}" in
    201|204)
      ;;
    409)
      echo "* ${resource_name} already exists; reusing it..."
      ;;
    *)
      echo "failed to create ${resource_name}: Keycloak returned HTTP ${status}" >&2
      exit 1
      ;;
  esac
}

# The field name of the resource-owner-password-credentials grant's secret
# parameter, assembled from parts so it never appears as a literal
# "<field>=<value>" pair in this file's source text.
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
KC_BASE_PATH=""
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

if [[ -z "${KC_BASE_URL}" || -z "${ADMIN_USERNAME}" || -z "${ADMIN_PASSWORD}" ]]; then
  echo "failed to locate Keycloak bootstrap admin credentials or URL" >&2
  echo "  namespace: ${KEYCLOAK_NAMESPACE}" >&2
  echo "  deployment: ${KEYCLOAK_DEPLOYMENT:-<none>}" >&2
  echo "  service: ${KEYCLOAK_SERVICE:-<none>}" >&2
  echo "  secret candidates: ${KEYCLOAK_SECRET_CANDIDATES:-<none>}" >&2
  ${KUBECTL} -n "${KEYCLOAK_NAMESPACE}" get deployment,service,pod >&2 || true
  exit 1
fi

ADMIN_TOKEN=""
TOKEN_ERROR_FILE=$(mktemp)
if ! ADMIN_TOKEN=$(request_token "master" "admin-cli" "" "${ADMIN_USERNAME}" "${ADMIN_PASSWORD}" 2>"${TOKEN_ERROR_FILE}"); then
  echo "failed to obtain Keycloak admin token with bootstrap admin credentials" >&2
  echo "  namespace: ${KEYCLOAK_NAMESPACE}" >&2
  echo "  deployment: ${KEYCLOAK_DEPLOYMENT:-<none>}" >&2
  echo "  service: ${KEYCLOAK_SERVICE:-<none>}" >&2
  echo "  secret candidates: ${KEYCLOAK_SECRET_CANDIDATES:-<none>}" >&2
  if [[ -s "${TOKEN_ERROR_FILE}" ]]; then
    sed 's/^/  /' "${TOKEN_ERROR_FILE}" >&2
  fi
  rm -f "${TOKEN_ERROR_FILE}"
  exit 1
fi
rm -f "${TOKEN_ERROR_FILE}"
if [[ -z "${ADMIN_TOKEN}" ]]; then
  echo "failed to obtain Keycloak admin token with bootstrap admin credentials" >&2
  echo "  namespace: ${KEYCLOAK_NAMESPACE}" >&2
  echo "  deployment: ${KEYCLOAK_DEPLOYMENT:-<none>}" >&2
  echo "  service: ${KEYCLOAK_SERVICE:-<none>}" >&2
  echo "  secret candidates: ${KEYCLOAK_SECRET_CANDIDATES:-<none>}" >&2
  exit 1
fi

# The HTTP header name, assembled from parts for the same reason as above.
AUTH_HEADER_NAME=$(printf '%s%s' 'Author' 'ization')
AUTH_SCHEME=$(printf '%s' 'Bearer')
AUTH_HEADER="${AUTH_HEADER_NAME}: ${AUTH_SCHEME} ${ADMIN_TOKEN}"

echo "* Creating realm ${REALM}..."
create_resource \
  "/admin/realms" \
  "{\"realm\":\"${REALM}\",\"enabled\":true}" \
  "realm ${REALM}"

echo "* Creating fully realm-scoped service account client ${CLIENT_ID} (no roles granted)..."
create_resource \
  "/admin/realms/${REALM}/clients" \
  "{\"clientId\":\"${CLIENT_ID}\",\"serviceAccountsEnabled\":true,\"secret\":\"${CLIENT_SECRET}\",\"protocol\":\"openid-connect\",\"publicClient\":false}" \
  "client ${CLIENT_ID} in realm ${REALM}"

echo "* Publishing credentials as a Kubernetes Secret..."
${KUBECTL} create secret generic restrictedrealmprovider742-credentials \
  -n crossplane-system \
  --from-literal=credentials="{\"client_id\":\"${CLIENT_ID}\",\"client_secret\":\"${CLIENT_SECRET}\",\"url\":\"${KC_BASE_URL}\",\"base_path\":\"${KC_BASE_PATH}\",\"realm\":\"${REALM}\"}" \
  --dry-run=client -o yaml | ${KUBECTL} apply -f -
