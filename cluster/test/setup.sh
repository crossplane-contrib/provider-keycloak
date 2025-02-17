#!/usr/bin/env bash
set -aeuo pipefail

#echo "Waiting until provider is healthy..."
#${KUBECTL} wait provider.pkg --all --for condition=Healthy --timeout 5m

#echo "Waiting for all pods to come online..."
#${KUBECTL} -n crossplane-system wait --for=condition=Available deployment --all --timeout=5m

echo "Creating a default provider config..."
cat <<EOF | ${KUBECTL} apply -f -
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-credentials
      namespace: crossplane-system
      key: credentials
