#!/usr/bin/env bash
set -aeuo pipefail

SCRIPT_DIR=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)
echo "Run Setup..."

${KUBECTL} apply -f ${SCRIPT_DIR}/../../dev/demos/basic/000-init.yaml

# Uptest creates by default these chainsaw test files in following folder /tmp/uptest-e2e/case/
# * 00-apply.yaml
# * 02-import.yaml
# * 03-delete.yaml

# We want to add more import tests that:
# 1. test if it finds the resource if the external name is set to a different value
# 2. test if it finds the resource if the external name is not set

cp ${SCRIPT_DIR}/hack/patchIncorrectExternalName.sh /tmp/patchIncorrectExternalName.sh
sed "s/patch.sh/patchIncorrectExternalName.sh/g" 02-import.yaml | sed "s/curl/#curl/g" > 02-import-IncorrectExtName.yaml

cp ${SCRIPT_DIR}/hack/patchRemoveExternalName.sh /tmp/patchRemoveExternalName.sh
sed "s/patch.sh/patchRemoveExternalName.sh/g" 02-import.yaml | sed "s/curl/#curl/g" > 02-import-NoExtName.yaml

echo "---" >> 02-import.yaml
cat 02-import-IncorrectExtName.yaml >> 02-import.yaml
echo "---" >> 02-import.yaml
cat 02-import-NoExtName.yaml >> 02-import.yaml

rm 02-import-IncorrectExtName.yaml
rm 02-import-NoExtName.yaml
