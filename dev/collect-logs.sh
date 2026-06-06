#!/bin/bash

mkdir -p logs/

kubectl get pods --all-namespaces -o wide > logs/pods.txt 2>&1 || true
kubectl get applications.argoproj.io --namespace argocd -o yaml > logs/argocd-applications.yaml 2>&1 || true
kubectl get events --all-namespaces --sort-by=.lastTimestamp > logs/events.txt 2>&1 || true

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
          if [[ $crd == *.m.crossplane* ]]; then
            kubectl get $crd $obj -n dev-ns -o yaml >> logs/$obj.$crd.yaml 2>&1 || true
          else
            kubectl get $crd $obj --all-namespaces -o yaml >> logs/$obj.$crd.yaml 2>&1 || true
          fi
          echo "---" >>  logs/$obj.$crd.yaml
          # Get events for the specific object
          kubectl get events --all-namespaces --field-selector involvedObject.name="$obj",involvedObject.apiVersion="$apiVersion" -o yaml >>  logs/$obj.$crd.yaml 2>&1 || true
          echo "---" >>  logs/$obj.$crd.yaml
      done
    fi
done

kubectl logs statefulset/keycloak-keycloakx -n keycloak > logs/keycloak.log 2>&1 || true
kubectl logs deployment/openldap-deployment -n keycloak > logs/openldap.log 2>&1 || true
kubectl logs deployment/crossplane -n crossplane-system crossplane > logs/crossplane.log 2>&1 || true
kubectl logs deployment/crossplane -n crossplane-system dev > logs/crossplane-dev.log 2>&1 || true
kubectl logs deployment/crossplane -n crossplane-system crossplane-init > logs/crossplane-init.log 2>&1 || true
provider=$(kubectl get deployment -n crossplane-system -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null | grep keycloak- || true)
if [[ -n "$provider" ]]; then
  kubectl logs deployment/${provider} -n crossplane-system > logs/crossplane-keycloak-provider.log 2>&1 || true
fi
