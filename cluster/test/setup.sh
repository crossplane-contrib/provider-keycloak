#!/usr/bin/env bash
set -aeuo pipefail

SCRIPT_DIR=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)
echo "Run Setup..."

${KUBECTL} apply -f ${SCRIPT_DIR}/../../dev/demos/basic/000-init.yaml
${KUBECTL} apply -f ${SCRIPT_DIR}/../../dev/demos/namespaced/000-init.yaml

# Uptest creates by default these chainsaw test files in following folder /tmp/uptest-e2e/case/
# * 00-apply.yaml
# * 02-import.yaml
# * 03-delete.yaml

# Increase timeouts
sed "s/exec: 20m0s/exec: 60m0s/g" 00-apply.yaml | sed "s/apply: 20m0s/apply: 60m0s/g" | sed "s/assert: 20m0s/assert: 60m0s/g"  > 00-apply-new.yaml
rm 00-apply.yaml
mv 00-apply-new.yaml  00-apply.yaml

sed "s/exec: 20m0s/exec: 60m0s/g" 02-import.yaml | sed "s/apply: 20m0s/apply: 60m0s/g" | sed "s/assert: 20m0s/assert: 60m0s/g" > 02-import-new.yaml
rm 02-import.yaml
mv 02-import-new.yaml  02-import.yaml

sed "s/exec: 20m0s/exec: 60m0s/g" 03-delete.yaml > 03-delete-new.yaml
rm 03-delete.yaml
mv 03-delete-new.yaml  03-delete.yaml


# We want to add more import tests that:
# 1. test if it finds the resource if the external name is set to a different value
# 2. test if it finds the resource if the external name is not set

cp ${SCRIPT_DIR}/hack/patchIncorrectExternalName.sh /tmp/patchIncorrectExternalName.sh
cp ${SCRIPT_DIR}/hack/patchIncorrectExternalName-ns.sh /tmp/patchIncorrectExternalName-ns.sh
sed "s/patch.sh/patchIncorrectExternalName.sh/g" 02-import.yaml | sed "s/patch-ns.sh/patchIncorrectExternalName-ns.sh/g" | sed "s/curl/#curl/g" > 02-import-IncorrectExtName.yaml

cp ${SCRIPT_DIR}/hack/patchRemoveExternalName.sh /tmp/patchRemoveExternalName.sh
cp ${SCRIPT_DIR}/hack/patchRemoveExternalName-ns.sh /tmp/patchRemoveExternalName-ns.sh
sed "s/patch.sh/patchRemoveExternalName.sh/g" 02-import.yaml | sed "s/patch-ns.sh/patchRemoveExternalName-ns.sh/g" | sed "s/curl/#curl/g" > 02-import-NoExtName.yaml

echo "---" >> 02-import.yaml
cat 02-import-IncorrectExtName.yaml >> 02-import.yaml
echo "---" >> 02-import.yaml
cat 02-import-NoExtName.yaml >> 02-import.yaml

rm 02-import-IncorrectExtName.yaml
rm 02-import-NoExtName.yaml

cp ${SCRIPT_DIR}/hack/deleteOrdered.sh /tmp/deleteOrdered.sh
sed 's/retry_kubectl "/eval "\/tmp\/deleteOrdered.sh /g' 03-delete.yaml > 03-delete-new.yaml
rm 03-delete.yaml
mv 03-delete-new.yaml  03-delete.yaml

