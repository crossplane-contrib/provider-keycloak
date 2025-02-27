#!/bin/bash
# Original Script: https://raw.githubusercontent.com/crossplane/uptest/refs/heads/main/hack/patch.sh
# Added: remove annotation crossplane.io/external-name
function patch {
    kindgroup=$1;
    name=$2;
    if ${KUBECTL} --subresource=status patch "$kindgroup/$name" --type=merge -p '{"status":{"conditions":[]}}' ; then
        return 0;
    else
        return 1;
    fi;
};


kindgroup=$1;
name=$2;
attempt=1;
max_attempts=10;
while [[ $attempt -le $max_attempts ]]; do
    if patch "$kindgroup" "$name"; then
        echo "Successfully patched $kindgroup/$name";
        ${KUBECTL} annotate "$kindgroup/$name" uptest-old-id=$(${KUBECTL} get "$kindgroup/$name" -o=jsonpath='{.status.atProvider.id}') --overwrite;
        ${KUBECTL} annotate "$kindgroup/$name" crossplane.io/external-name- --overwrite;
        break;
    else
        printf "Retrying... (%d/%d) for %s/%s\n" "$attempt" "$max_attempts" "$kindgroup" "$name" >&2;
    fi;
    ((attempt++));
    sleep 5;
done;
if [[ $attempt -gt $max_attempts ]]; then
    echo "Failed to patch $kindgroup/$name after $max_attempts attempts";
    exit 1;
fi;
exit 0;