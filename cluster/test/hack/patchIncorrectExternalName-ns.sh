#!/bin/bash
# Original Script: https://raw.githubusercontent.com/crossplane/uptest/refs/heads/main/hack/patch-ns.sh
# Added: set annotation crossplane.io/external-name to Incorrect
function patch {
    kindgroup=$1;
    name=$2;
    namespace=$3;
    if ${KUBECTL} --subresource=status patch --namespace "$namespace" "$kindgroup/$name" --type=merge -p '{"status":{"conditions":[]}}' ; then
        return 0;
    else
        return 1;
    fi;
};


kindgroup=$1;
name=$2;
namespace=$3;
attempt=1;
max_attempts=10;
while [[ $attempt -le $max_attempts ]]; do
    if patch "$kindgroup" "$name" "$namespace"; then
        echo "Successfully patched $kindgroup/$name";
        ${KUBECTL} annotate --namespace "$namespace" "$kindgroup/$name" uptest-old-id=$(${KUBECTL} get --namespace "$namespace" "$kindgroup/$name" -o=jsonpath='{.status.atProvider.id}') --overwrite;
        ${KUBECTL} annotate --namespace "$namespace" "$kindgroup/$name" crossplane.io/external-name=Incorrect --overwrite;
        break;
    else
        printf "Retrying... (%d/%d) for %s/%s/%s\n" "$attempt" "$max_attempts" "$kindgroup" "$name" "$namespace" >&2;
    fi;
    ((attempt++));
    sleep 5;
done;
if [[ $attempt -gt $max_attempts ]]; then
    echo "Failed to patch $kindgroup/$name after $max_attempts attempts";
    exit 1;
fi;
exit 0;