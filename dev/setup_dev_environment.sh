#!/bin/bash

# check if user can run docker without sudo, if not create an alias for sudo docker for this session
if ! docker ps >/dev/null 2>&1; then
  echo "You need to be able to run docker without sudo. Adding alias for this session."
  alias docker='sudo docker'
fi


echo "Checking dependencies"
command -v docker >/dev/null 2>&1 || { echo >&2 "Docker is required but not installed.  Aborting."; exit 1; }
command -v kind >/dev/null 2>&1 || { echo >&2 "Kind is required but not installed.  Aborting."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo >&2 "Kubectl is required but not installed.  Aborting."; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq is required but not installed.  Aborting."; exit 1; }
command -v envsubst >/dev/null 2>&1 || { echo >&2 "envsubst is required but not installed.  Aborting."; exit 1; }
command -v base64 >/dev/null 2>&1 || { echo >&2 "base64 is required but not installed.  Aborting."; exit 1; }
command -v sed >/dev/null 2>&1 || { echo >&2 "sed is required but not installed.  Aborting."; exit 1; }
echo "All dependencies are installed."

echo "Creating cluster"
kind create cluster --name fenrir-1 --config kind-config.yaml --kubeconfig $HOME/.kube/fenrir-1

echo "Running some commands to make sure the cluster is ready"
export KUBECONFIG=$HOME/.kube/fenrir-1
kubectl cluster-info
kubectl get nodes

echo "Switching context to fenrir-1"
kubectl config use-context kind-fenrir-1

echo "########### Setup MetalLB ###########"
echo "* Installing MetalLB"
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml
echo "* Waiting for MetalLB to be ready"
kubectl wait --namespace metallb-system --for=condition=ready pod --selector=app=metallb

echo "* Get IPAM config: "
docker network inspect kind
export IP_PREFIX=$(docker network inspect kind | jq -r .[].IPAM.Config[0].Subnet | sed -r 's|\.0/[0-9]+||g')
echo "* IP Prefix: $IP_PREFIX"
export IP_RANGE_START="$IP_PREFIX.30"
export IP_RANGE_END="$IP_PREFIX.80"
echo "* IP Range: $IP_RANGE_START - $IP_RANGE_END"
cat metallb/pools.yaml | envsubst | kubectl apply -f -

echo "########### Installing ArgoCD ###########"
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
echo "* Exposing ArgoCD"
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "LoadBalancer"}}'

while [[ -z $(kubectl get svc -n argocd argocd-server -o jsonpath="{.status.loadBalancer.ingress}" 2>/dev/null) ]]; do
  echo "** still waiting for argocd/argocd-server to get ingress"
  sleep 1
done
echo "* argocd/argocd-server now has ingress."

export ARGOCD_IP=$(kubectl -n argocd get svc argocd-server -o json | jq -r .status.loadBalancer.ingress[0].ip)

echo "* Waiting for ArgoCD to be ready"
kubectl wait pod --all --for=condition=Ready --namespace argocd --timeout=300s


echo "########### Installing Keycloak ###########"
kubectl apply -f apps/keycloak.yaml
sleep 5
kubectl wait pod --all --for=condition=Ready --namespace keycloak --timeout=300s

while [[ -z $(kubectl get svc -n keycloak keycloak-keycloakx-http -o jsonpath="{.status.loadBalancer.ingress}" 2>/dev/null) ]]; do
  echo "** still waiting for service keycloak/keycloak-keycloakx-http to get ingress"
  sleep 1
done

export KEYCLOAK_IP=$(kubectl -n keycloak get svc keycloak-keycloakx-http -o json | jq -r .status.loadBalancer.ingress[0].ip)
export KEYCLOAK_PORT=$(kubectl -n keycloak get svc keycloak-keycloakx-http -o json | jq -r .spec.ports[0].port)
export KEYCLOAK_USER=admin
export KEYCLOAK_PASSWORD=admin


echo "########### Installing Crossplane ###########"

kubectl apply -f apps/crossplane.yaml
sleep 10
kubectl wait pod --all --for=condition=Ready --namespace crossplane-system --timeout=300s


echo "########### Installing Keycloak Provider secret ###########"
cat iam/keycloak-provider-secret.yaml | envsubst | kubectl apply --namespace crossplane-system  -f -

echo "#################################################"
echo "You're ready to go!"
echo "ArgoCD is ready at https://$ARGOCD_IP:443"
echo "ArgoCD login: admin / $(kubectl -n argocd get secrets argocd-initial-admin-secret -o json | jq -r .data.password | base64 -d)"
echo "-------------------------------------------------"
echo "Keycloak is ready at http://$KEYCLOAK_IP:$KEYCLOAK_PORT/auth"
echo "Keycloak login: admin / admin"
echo "#################################################"
echo "To delete the cluster run: kind delete cluster --name fenrir-1"
echo "#################################################"
echo "You can now run the provider by executing: make run"
echo "#################################################"

