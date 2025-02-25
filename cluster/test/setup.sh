#!/usr/bin/env bash
set -aeuo pipefail

SCRIPT_DIR=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)
echo "Run Setup..."

${KUBECTL} apply -f ${SCRIPT_DIR}/../../dev/demos/basic/000-init.yaml