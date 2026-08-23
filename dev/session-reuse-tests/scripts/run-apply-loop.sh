#!/usr/bin/env bash
# Repeatedly runs `tofu apply` in a scenario directory to simulate the
# reconciliation frequency of a long-running Crossplane provider, so the
# resulting Keycloak session count can be compared against the Crossplane
# scenarios using check-sessions.sh.
#
# Usage:
#   ./run-apply-loop.sh -d DIR [-n ITERATIONS] [-i INTERVAL_SECONDS] [-- TF_VAR args...]
#
# Example:
#   cd dev/session-reuse-tests/opentofu/scenario-password-grant
#   tofu init
#   ../../scripts/run-apply-loop.sh -d . -n 20 -i 5 \
#     -- -var keycloak_url=http://127.0.0.1:8080
set -euo pipefail

DIR="."
ITERATIONS=10
INTERVAL=5

usage() {
  sed -n '2,15p' "$0" | sed 's/^# \{0,1\}//'
}

while [ $# -gt 0 ]; do
  case "$1" in
    -d|--dir) DIR="$2"; shift 2 ;;
    -n|--iterations) ITERATIONS="$2"; shift 2 ;;
    -i|--interval) INTERVAL="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    --) shift; break ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

TF_ARGS=("$@")

if ! command -v tofu >/dev/null 2>&1; then
  echo "error: 'tofu' (OpenTofu CLI) not found in PATH" >&2
  exit 1
fi

pushd "$DIR" >/dev/null

for i in $(seq 1 "$ITERATIONS"); do
  echo "==> apply $i/$ITERATIONS ($(date -u +%Y-%m-%dT%H:%M:%SZ))"
  tofu apply -auto-approve "${TF_ARGS[@]}"
  if [ "$i" -lt "$ITERATIONS" ]; then
    sleep "$INTERVAL"
  fi
done

popd >/dev/null
