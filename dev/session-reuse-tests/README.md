# Keycloak Session Reuse Test Scenarios

Local test scenarios for validating and comparing the Keycloak
authentication-session behavior described in
[`docs/content/docs/using/reference/provider-config.md#authentication-sessions`](../../docs/content/docs/using/reference/provider-config.md#authentication-sessions)
and [issue #309](https://github.com/crossplane-contrib/provider-keycloak/issues/309).

These scenarios let you compare, side by side, how many Keycloak sessions
accumulate for:

1. **This provider (Crossplane)** vs. **plain OpenTofu** using the same
   `terraform-provider-keycloak` dependency.
2. The **resource-owner password grant** (`username`/`password`) vs. the
   **client credentials grant** (`client_id`/`client_secret`).

## Layout

```
dev/session-reuse-tests/
├── crossplane/
│   ├── scenario-password-grant/       # ProviderConfig + resources, password grant
│   └── scenario-client-credentials/   # ProviderConfig + resources, client_credentials grant
├── opentofu/
│   ├── scenario-password-grant/       # Equivalent HCL, password grant
│   └── scenario-client-credentials/   # Equivalent HCL, client_credentials grant
└── scripts/
    ├── check-sessions.sh                     # Query active Keycloak sessions for a user/client
    ├── run-apply-loop.sh                     # Repeat `tofu apply` to simulate reconciliation
    └── setup-client-credentials-client.sh    # One-time: create the service-account client
```

## Prerequisites

- A running Keycloak instance reachable at `$KEYCLOAK_IP:$KEYCLOAK_PORT` (the
  [`dev/setup_dev_environment.sh`](../setup_dev_environment.sh) script sets
  one up, including exporting those two variables) — or any other Keycloak
  instance you can reach.
- `curl` and `jq` (used by `check-sessions.sh`).
- [OpenTofu](https://opentofu.org/) (`tofu` CLI) for the `opentofu/` scenarios.
- For the client-credentials scenarios: a confidential client with a service
  account granted the `realm-admin` role of the `realm-management` client.
  You can reuse the `provider-e2e-client` created by
  `dev/setup_dev_environment.sh` (in the `provider-e2e-realm` realm), or run
  the bundled setup helper to create a dedicated one
  (`session-test-client` / `session-test-secret`) in the `master` realm:

  ```bash
  KEYCLOAK_IP=... KEYCLOAK_PORT=... ./scripts/setup-client-credentials-client.sh
  ```

## Scenario 1: Crossplane, password grant (reproduces #309)

```bash
export KEYCLOAK_IP=... KEYCLOAK_PORT=...
cat crossplane/scenario-password-grant/provider-config.yaml | envsubst | kubectl apply -f -
kubectl apply -f crossplane/scenario-password-grant/resources.yaml

# In a separate terminal, watch the session count for the "admin" user while
# the provider reconciles (Ctrl+C to stop):
KC_BASE_URL="http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}" \
  ./scripts/check-sessions.sh -u admin -w 15
```

Expected (current behavior): the provider only logs in once per
`ProviderConfig` at startup, but every access-token expiry triggers a
password re-authentication rather than a token refresh, so the session count
for `admin` keeps growing over time (see the linked issue for the root
cause upstream in `terraform-provider-keycloak`).

## Scenario 2: Crossplane, client credentials grant (control group)

```bash
cat crossplane/scenario-client-credentials/provider-config.yaml | envsubst | kubectl apply -f -
kubectl apply -f crossplane/scenario-client-credentials/resources.yaml

KC_BASE_URL="http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}" \
  ./scripts/check-sessions.sh -c session-test-client -w 15
```

Expected: the session count for the `session-test-client` service account
stays flat (0 or 1) since this grant type does not go through the
password-grant re-authentication path.

## Scenario 3: OpenTofu, password grant (baseline for comparison)

```bash
cd opentofu/scenario-password-grant
tofu init
../../scripts/run-apply-loop.sh -d . -n 20 -i 10 \
  -- -var "keycloak_url=http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}"

KC_BASE_URL="http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}" \
  ../../scripts/check-sessions.sh -u admin -j
```

Note: unlike the long-running Crossplane provider process, each `tofu apply`
invocation starts a fresh provider plugin process and therefore performs its
own fresh login regardless of grant type. This scenario establishes the
"one session per apply" baseline so you can see how much worse the
password-grant re-authentication-per-token-refresh behavior is for a
long-running consumer like this provider, where the process is never
restarted between reconciles.

## Scenario 4: OpenTofu, client credentials grant (baseline for comparison)

```bash
cd opentofu/scenario-client-credentials
tofu init
../../scripts/run-apply-loop.sh -d . -n 20 -i 10 \
  -- -var "keycloak_url=http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}"

KC_BASE_URL="http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}" \
  ../../scripts/check-sessions.sh -c session-test-client -j
```

## Tooling: `check-sessions.sh`

Queries the Keycloak Admin REST API
(`GET /admin/realms/{realm}/users/{id}/sessions`) for a given username or
service-account client ID and prints the active session count. Supports a
`-w/--watch SECONDS` flag to poll repeatedly (useful while a scenario is
reconciling) and a `-j/--json` flag to print the raw session list (id,
`ipAddress`, `start`, `lastAccess`, `clients`).

```bash
./scripts/check-sessions.sh --help
```

## Cleanup

```bash
kubectl delete -f crossplane/scenario-password-grant/resources.yaml --ignore-not-found
kubectl delete -f crossplane/scenario-client-credentials/resources.yaml --ignore-not-found
kubectl delete providerconfigs.keycloak.crossplane.io session-test-password-grant session-test-client-credentials --ignore-not-found
kubectl delete secret -n crossplane-system session-test-password-grant session-test-client-credentials --ignore-not-found

(cd opentofu/scenario-password-grant && tofu destroy -auto-approve -var "keycloak_url=http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}")
(cd opentofu/scenario-client-credentials && tofu destroy -auto-approve -var "keycloak_url=http://${KEYCLOAK_IP}:${KEYCLOAK_PORT}")
```
