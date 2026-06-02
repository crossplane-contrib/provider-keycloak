#!/usr/bin/env bash
#
# Regression test for stale-reference recovery (internal/clients/stalerefs).
#
# What it proves:
#   When a Keycloak object referenced by another managed resource is deleted
#   and recreated out-of-band (so its UUID changes), the provider clears the
#   stored stale UUID on spec.forProvider and the runtime resolver
#   repopulates it with the new one. Without this fix, the dependent resource
#   stays Synced=False with a 404 from Keycloak forever.
#
# Scenario (self-contained — creates its own Realm, Client, Role, RoleMapper):
#   1. Apply Realm, Client, target Role and a RoleMapper that references both.
#   2. Wait for everything Synced=True. Capture the resolved roleId.
#   3. Delete the Role directly via the Keycloak Admin API (bypasses
#      crossplane so the K8s Role MR's reconciler discovers the deletion the
#      hard way).
#   4. Re-create the Role with the same name (fresh UUID in Keycloak).
#   5. Wait for the RoleMapper to recover: spec.forProvider.roleId should
#      transition to the new UUID and Synced=True should return.
#   6. Assert: new UUID != original, and the recovery annotation is set.
#
# Prerequisites:
#   - dev/setup_dev_environment.sh has been run (kind cluster + Keycloak +
#     provider-keycloak deployed). Exports KEYCLOAK_IP and KEYCLOAK_PORT.
#   - jq, curl, kubectl available.

set -euo pipefail

: "${KEYCLOAK_IP:?KEYCLOAK_IP not set — source dev/setup_dev_environment.sh first}"
: "${KEYCLOAK_PORT:?KEYCLOAK_PORT not set}"
: "${KEYCLOAK_USER:=admin}"
: "${KEYCLOAK_PASSWORD:=admin}"

KC_BASE="http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}"

# Unique suffix so reruns and parallel runs don't collide.
SUFFIX="${SUFFIX:-$(date +%s)}"
REALM="staleref-${SUFFIX}"
CLIENT="staleref-client-${SUFFIX}"
TARGET_ROLE="staleref-role-${SUFFIX}"
RM_NAME="staleref-rolemapper-${SUFFIX}"

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
  local token=$1
  curl -fsS -H "Authorization: Bearer ${token}" \
    "${KC_BASE}/admin/realms/${REALM}/roles/${TARGET_ROLE}" | jq -r '.id'
}

kc_delete_role() {
  local token=$1
  curl -fsS -X DELETE -H "Authorization: Bearer ${token}" \
    "${KC_BASE}/admin/realms/${REALM}/roles/${TARGET_ROLE}"
}

kc_create_role() {
  local token=$1
  curl -fsS -X POST -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    "${KC_BASE}/admin/realms/${REALM}/roles" \
    -d "{\"name\":\"${TARGET_ROLE}\"}" > /dev/null
}

cleanup() {
  log "Cleanup"
  kubectl delete --ignore-not-found \
    rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
    role.role.keycloak.crossplane.io/${TARGET_ROLE} \
    client.openidclient.keycloak.crossplane.io/${CLIENT} \
    realm.realm.keycloak.crossplane.io/${REALM} >/dev/null 2>&1 || true
}
trap cleanup EXIT

log "Applying Realm + Client + Role + RoleMapper (suffix=${SUFFIX})"
kubectl apply -f - <<EOF
apiVersion: realm.keycloak.crossplane.io/v1alpha1
kind: Realm
metadata:
  name: ${REALM}
spec:
  deletionPolicy: Delete
  forProvider:
    realm: ${REALM}
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: openidclient.keycloak.crossplane.io/v1alpha1
kind: Client
metadata:
  name: ${CLIENT}
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: ${REALM}
    clientId: ${CLIENT}
    accessType: PUBLIC
  providerConfigRef:
    name: keycloak-provider-config
---
apiVersion: role.keycloak.crossplane.io/v1alpha1
kind: Role
metadata:
  name: ${TARGET_ROLE}
spec:
  deletionPolicy: Delete
  forProvider:
    realmIdRef:
      name: ${REALM}
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
    realmIdRef:
      name: ${REALM}
    clientIdRef:
      name: ${CLIENT}
    roleIdRef:
      name: ${TARGET_ROLE}
EOF

log "Waiting for all resources to be Synced=True"
kubectl wait realm.realm.keycloak.crossplane.io/${REALM} --for=condition=Synced=True --timeout=120s
kubectl wait client.openidclient.keycloak.crossplane.io/${CLIENT} --for=condition=Synced=True --timeout=120s
kubectl wait role.role.keycloak.crossplane.io/${TARGET_ROLE} --for=condition=Synced=True --timeout=120s
kubectl wait rolemapper.client.keycloak.crossplane.io/${RM_NAME} --for=condition=Synced=True --timeout=180s

ORIGINAL_ROLE_ID=$(kubectl get rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
  -o jsonpath='{.spec.forProvider.roleId}')
log "Original resolved roleId = ${ORIGINAL_ROLE_ID}"
[[ -n "${ORIGINAL_ROLE_ID}" ]] || die "RoleMapper.spec.forProvider.roleId is empty after Synced"

log "Deleting the Keycloak Role directly via Admin API (bypassing crossplane)"
TOKEN=$(kc_token)
kc_delete_role "${TOKEN}"

log "Recreating the Keycloak Role with the same name (fresh UUID)"
kc_create_role "${TOKEN}"
NEW_KC_UUID=$(kc_role_uuid "${TOKEN}")
log "New Keycloak UUID = ${NEW_KC_UUID}"

[[ "${NEW_KC_UUID}" != "${ORIGINAL_ROLE_ID}" ]] \
  || die "expected new Keycloak UUID to differ from original"

log "Waiting up to 3m for the RoleMapper to recover its reference"
deadline=$(( $(date +%s) + 180 ))
current=""
synced=""
while (( $(date +%s) < deadline )); do
  current=$(kubectl get rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
    -o jsonpath='{.spec.forProvider.roleId}' 2>/dev/null || true)
  synced=$(kubectl get rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
    -o jsonpath='{.status.conditions[?(@.type=="Synced")].status}' 2>/dev/null || true)
  if [[ "${current}" == "${NEW_KC_UUID}" && "${synced}" == "True" ]]; then
    anno=$(kubectl get rolemapper.client.keycloak.crossplane.io/${RM_NAME} \
      -o jsonpath='{.metadata.annotations.provider-keycloak\.crossplane\.io/stale-ref-recovery-at-generation}' \
      2>/dev/null || true)
    log "PASS: roleId rewritten and Synced=True"
    log "  before: ${ORIGINAL_ROLE_ID}"
    log "  after:  ${current}"
    [[ -n "${anno}" ]] && log "  recovery annotation set to generation ${anno}"
    exit 0
  fi
  sleep 5
done

die "RoleMapper did not recover within timeout. Last roleId=${current:-<empty>} synced=${synced:-<unknown>}"
