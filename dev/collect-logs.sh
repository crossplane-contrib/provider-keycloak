#!/bin/bash

mkdir -p logs/

# Get all CRDs containing "crossplane" in their names
for crd in $(kubectl get crd -o jsonpath='{.items[*].metadata.name}'); do
    if [[ $crd == *crossplane* ]]; then
      # Get the API version for the CRD
      apiVersion=$(kubectl get crd $crd -o jsonpath='{.spec.group}/{.spec.versions[0].name}')
      echo  $apiVersion
      # Iterate over each object of this CRD type
      for obj in $(kubectl get $crd --all-namespaces -o jsonpath='{.items[*].metadata.name}'); do
          echo "->" $obj

          touch  logs/$obj.$crd.yaml
          # Get the YAML for the specific object
          kubectl get $crd $obj --all-namespaces -o yaml >> logs/$obj.$crd.yaml
          echo "---" >>  logs/$obj.$crd.yaml
          # Get events for the specific object
          kubectl get events --all-namespaces --field-selector involvedObject.name="$obj",involvedObject.apiVersion="$apiVersion" -o yaml >>  logs/$obj.$crd.yaml
          echo "---" >>  logs/$obj.$crd.yaml
      done
    fi
done

kubectl logs statefulset/keycloak-keycloakx -n keycloak > logs/keycloak.log
kubectl logs deployment/openldap-deployment -n keycloak > logs/openldap.log
kubectl logs deployment/crossplane -n crossplane-system crossplane > logs/crossplane.log
kubectl logs deployment/crossplane -n crossplane-system dev > logs/crossplane-dev.log
kubectl logs deployment/crossplane -n crossplane-system crossplane-init > logs/crossplane-init.log
provider=$(kubectl get deployment -n crossplane-system -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}'| grep keycloak-)
kubectl logs deployment/${provider} -n crossplane-system > logs/crossplane-keycloak-provider.log

