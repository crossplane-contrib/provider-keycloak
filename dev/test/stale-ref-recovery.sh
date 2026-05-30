#!/usr/bin/env bash
#
# End-to-end test for stale-reference recovery
# (internal/clients/stalerefs).
#
# Scenario:
#   1. Apply a Realm + a "target" Role + a Role + a RoleMapper that references
#      the target Role via roleIdRef.
#   2. Wait for everything Synced=True.
#   3. Capture the resolved spec.forProvider.roleId on the RoleMapper (this is
#      the Keycloak UUID).
#   4. Delete the target Role directly via the Keycloak Admin API (bypassing
#      the K8s reconciler so the local Role MR's status doesn't auto-update).
#   5. Re-create the target Role with the same name. The new Keycloak Role has
#      a fresh UUID.
#   6. Watch the Role MR's reconciler observe its own external resource and
#      pick up the new UUID into status.atProvider.id.
#   7. Watch the RoleMapper's reconciler hit a 404 from Keycloak, the provider
#      clear spec.forProvider.roleId via stalerefs.MaybeRecover, and the
#      runtime resolver repopulate it with the new UUID. Synced=True returns.
#   8. Assert: the new resolved roleId != the original.
#
# Usage:
#   dev/test/stale-ref-recovery.sh
#
# Requires the dev environment from dev/setup_dev_environment.sh to be
# running. KEYCLOAK_IP, KEYCLOAK_PORT, KEYCLOAK_USER, KEYCLOAK_PASSWORD must
# be exported (the setup script exports them).

set -euo pipefail

: "${KEYCLOAK_IP:?KEYCLOAK_IP not set — source dev/setup_dev_environment.sh first}"
: "${KEYCLOAK_PORT:?KEYCLOAK_PORT not set}"
: "${KEYCLOAK_USER:=admin}"
: "${KEYCLOAK_PASSWORD:=admin}"

KC_BASE="http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}"
REALM="dev"
TARGET_ROLE="stale-ref-target-role"
RM_NAME="stale-ref-rolemapper"
SOURCE_CLIENT="test"

log() { printf "==> %s\n" "$*"; }
die() { printf "FAIL: %s\n" "$*" >&2; exit 1; }

kc_token() {
  curl -fsS -X POST "${KC_BASE}/realms/master/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=${KEYCLOAK_USER}" \
    -d "password=${KEYCLOAK_PASSWORD}" | jq -r '.access_token'
}

kc_role_uuid() {
  local token=$1 name=$2
  curl -fsS -H "Authorization: Bearer ${token}" \
    "${KC_BASE}/admin/realms/${REALM}/roles/${name}" | jq -r '.id'
}

kc_delete_role() {
  local token=$1 name=$2
  curl -fsS -X DELETE -H "Authorization: Bearer ${token}" \
    "${KC_BASE}/admin/realms/${REALM}/roles/${name}"
}

kc_create_role() {
  local token=$1 name=$2
  curl -fsS -X POST -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    "${KC_BASE}/admin/realms/${REALM}/roles" \
    -d "{\"name\":\"${name}\"}"
}

cleanup() {
  log "Cleanup"
  kubectl delete -f - --ignore-not-found <<EOF || true
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: RoleMapper
metadata: {name: ${RM_NAME}}
---
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata: {name: ${TARGET_ROLE}}
EOF
}
trap cleanup EXIT

log "Applying Role + RoleMapper"
kubectl apply -f - <<EOF
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: ${TARGET_ROLE}
spec:
  deletionPolicy: Delete
  forProvider:
    realmId: ${REALM}
    name: ${TARGET_ROLE}
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: client.keycloak.crossplane.io/v1alpha1
kind: RoleMapper
metadata:
  name: ${RM_NAME}
spec:
  deletionPolicy: Delete
  providerConfigRef:
    name: keycloak-provider-config
  forProvider:
    realmId: ${REALM}
    clientIdRef:
      name: ${SOURCE_CLIENT}
    roleIdRef:
      name: ${TARGET_ROLE}
EOF

log "Waiting for Role to be Synced=True"
kubectl wait role.role.keycloak.crossplane.io/${TARGET_ROLE} \
  --for=condition=Synced=True --timeout=120s

log "Waiting for RoleMapper to be Synced=True"
kubectl wait rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
  --for=condition=Synced=True --timeout=120s

ORIGINAL_ROLE_ID=$(kubectl get rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
  -o jsonpath='{.spec.forProvider.roleId}')
log "Original resolved roleId = ${ORIGINAL_ROLE_ID}"

[[ -n "${ORIGINAL_ROLE_ID}" ]] || die "RoleMapper.spec.forProvider.roleId is empty after Synced"

log "Deleting the Keycloak Role directly via Admin API (bypassing crossplane)"
TOKEN=$(kc_token)
kc_delete_role "${TOKEN}" "${TARGET_ROLE}"

log "Recreating the Keycloak Role with the same name (will have a fresh UUID)"
kc_create_role "${TOKEN}" "${TARGET_ROLE}"
NEW_KC_UUID=$(kc_role_uuid "${TOKEN}" "${TARGET_ROLE}")
log "New Keycloak UUID = ${NEW_KC_UUID}"

[[ "${NEW_KC_UUID}" != "${ORIGINAL_ROLE_ID}" ]] \
  || die "expected new Keycloak UUID to differ from original"

log "Waiting up to 3m for the RoleMapper to recover its reference"
deadline=$(( $(date +%s) + 180 ))
while (( $(date +%s) < deadline )); do
  current=$(kubectl get rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
    -o jsonpath='{.spec.forProvider.roleId}' 2>/dev/null || true)
  synced=$(kubectl get rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
    -o jsonpath='{.status.conditions[?(@.type=="Synced")].status}' 2>/dev/null || true)
  if [[ "${current}" == "${NEW_KC_UUID}" && "${synced}" == "True" ]]; then
    log "PASS: roleId rewritten to new UUID and Synced=True"
    log "  before: ${ORIGINAL_ROLE_ID}"
    log "  after:  ${current}"

    anno=$(kubectl get rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
      -o jsonpath='{.metadata.annotations.provider-keycloak\.crossplane\.io/stale-ref-recovery-at-generation}' \
      2>/dev/null || true)
    if [[ -n "${anno}" ]]; then
      log "  recovery annotation set to generation ${anno}"
    fi
    exit 0
  fi
  sleep 5
done

die "RoleMapper did not recover within timeout. Last roleId=${current:-<empty>} synced=${synced:-<unknown>}"
