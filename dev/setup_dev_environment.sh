#!/bin/bash
set -eo pipefail

# Default variable values
CLUSTER_NAME="fenrir-1"
skipmetallb=false

# Function to display script usage
usage() {
 echo "Usage: $0 [OPTIONS]"
 echo "Options:"
 echo " -h, --help           Display this help message"
 echo " -c, --cluster-name   Name of the Cluster"
 echo " -s, --skip-metal-lb  Do not install MetalLB"
}

has_argument() {
    [[ ("$1" == *=* && -n ${1#*=}) || ( ! -z "$2" && "$2" != -*)  ]];
}

extract_argument() {
  echo "${2:-${1#*=}}"
}

# Function to handle options and arguments
handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      -h | --help)
        usage
        exit 0
        ;;
      -s | --skip-metal-lb)
        skipmetallb=true
        ;;
      -c | --cluster-name*)
        if ! has_argument $@; then
          echo "Clustername not specified." >&2
          usage
          exit 1
        fi

        CLUSTER_NAME=$(extract_argument $@)

        shift
        ;;
      *)
        echo "Invalid option: $1" >&2
        usage
        exit 1
        ;;
    esac
    shift
  done
}

# Main script execution
handle_options "$@"

echo "Cluster name: $CLUSTER_NAME"

echo "########### Checking dependencies ###########"
command -v docker >/dev/null 2>&1 || { echo >&2 "Docker is required but not installed.  Aborting."; exit 1; }
command -v kind >/dev/null 2>&1 || { echo >&2 "Kind is required but not installed.  Aborting."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo >&2 "Kubectl is required but not installed.  Aborting."; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq is required but not installed.  Aborting."; exit 1; }
command -v envsubst >/dev/null 2>&1 || { echo >&2 "envsubst is required but not installed.  Aborting."; exit 1; }
command -v base64 >/dev/null 2>&1 || { echo >&2 "base64 is required but not installed.  Aborting."; exit 1; }
command -v sed >/dev/null 2>&1 || { echo >&2 "sed is required but not installed.  Aborting."; exit 1; }
echo "All dependencies are installed."

# check if user can run docker without sudo, if not create an alias for sudo docker for this session
if ! docker ps >/dev/null 2>&1; then
  echo "Sudo required for docker."
  sudo_prefix='sudo'
else
  sudo_prefix=''
fi

echo "########### Setup Cluster ###########"
if $sudo_prefix kind get clusters | grep "$CLUSTER_NAME" >/dev/null 2>&1; then
  echo "$CLUSTER_NAME cluster already exists"
else
  echo "Creating cluster"
  old_context=$(kubectl config current-context || echo "notset")
  $sudo_prefix kind create cluster --name $CLUSTER_NAME --config kind-config.yaml --kubeconfig $HOME/.kube/$CLUSTER_NAME
  $sudo_prefix chown $USER:$USER $HOME/.kube/$CLUSTER_NAME
  if [[ ! "$old_context" == "notset" ]]; then
    echo "Restore old context $old_context"
    kubectl config use-context $old_context
  fi
fi

echo "Running some commands to make sure the cluster is ready"
export KUBECONFIG=$HOME/.kube/$CLUSTER_NAME
kubectl_cmd="kubectl --context=kind-$CLUSTER_NAME"
$kubectl_cmd cluster-info
$kubectl_cmd get nodes

if [[ "$skipmetallb" == "true" ]]; then
echo "Skipping MetalLB"
else
echo "########### Setup MetalLB ###########"
echo "* Installing MetalLB"
$kubectl_cmd apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml
echo "* Waiting for MetalLB to be ready"
$kubectl_cmd wait --namespace metallb-system --for=condition=ready --all pod --selector=app=metallb  --timeout=300s

echo "* Get IPAM config: "
$sudo_prefix docker network inspect kind
echo "* Available subnets (IPv4 required): "
$sudo_prefix docker network inspect kind | jq -r .[].IPAM.Config[].Subnet

export IP_PREFIX=$($sudo_prefix docker network inspect kind | jq -r .[].IPAM.Config[].Subnet | grep -E "([0-9]+\.){3}0/[0-9]+" | sed -r 's|\.0/[0-9]+||g')
echo "* Found IP Prefix: $IP_PREFIX"
# if CLUSTER_NAME == "fenrir-1" then IP_PREFIX == "172.18.0" else IP_PREFIX == "172.19.0"
if [[ $CLUSTER_NAME == "fenrir-1" ]]; then
  export IP_RANGE_START="$IP_PREFIX.30"
  export IP_RANGE_END="$IP_PREFIX.55"
else
  export IP_RANGE_START="$IP_PREFIX.56"
  export IP_RANGE_END="$IP_PREFIX.80"
fi

echo "* Set IP Range: $IP_RANGE_START - $IP_RANGE_END"
while true; do
  cat metallb/pools.yaml | envsubst | $kubectl_cmd apply -f - && break  # Break the loop if command succeeds
  echo "** still waiting for metallb resources to be ready"
  sleep 1
done
fi

echo "########### Installing ArgoCD ###########"
$kubectl_cmd create namespace argocd || true
$kubectl_cmd apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
echo "* Exposing ArgoCD"
$kubectl_cmd patch svc argocd-server -n argocd -p '{"spec": {"type": "LoadBalancer"}}'

while [[ -z $($kubectl_cmd get svc -n argocd argocd-server -o jsonpath="{.status.loadBalancer.ingress}" 2>/dev/null) ]]; do
  echo "** still waiting for argocd/argocd-server to get ingress"
  sleep 1
done
echo "* argocd/argocd-server now has ingress."

export ARGOCD_IP=$($kubectl_cmd -n argocd get svc argocd-server -o json | jq -r .status.loadBalancer.ingress[0].ip)

echo "* Waiting for ArgoCD to be ready"
$kubectl_cmd wait pod --all --for=condition=Ready --namespace argocd --timeout=300s


echo "########### Installing Keycloak ###########"
if $kubectl_cmd diff -f apps/keycloak.yaml >/dev/null 2>&1; then
  echo "Keycloak up-to-date."
else
  $kubectl_cmd apply -f apps/keycloak.yaml
  sleep 5
  $kubectl_cmd wait pod --all --for=condition=Ready --namespace keycloak --timeout=300s
fi

while [[ -z $($kubectl_cmd get svc -n keycloak keycloak-keycloakx-http -o jsonpath="{.status.loadBalancer.ingress}" 2>/dev/null) ]]; do
  echo "** still waiting for service keycloak/keycloak-keycloakx-http to get ingress"
  sleep 5
done

export KEYCLOAK_IP=$($kubectl_cmd -n keycloak get svc keycloak-keycloakx-http -o json | jq -r .status.loadBalancer.ingress[0].ip)
export KEYCLOAK_PORT=$($kubectl_cmd -n keycloak get svc keycloak-keycloakx-http -o json | jq -r .spec.ports[0].port)
export KEYCLOAK_USER=admin
export KEYCLOAK_PASSWORD=admin


echo "########### Installing Crossplane ###########"
if $kubectl_cmd diff -f apps/crossplane.yaml >/dev/null 2>&1; then
  echo "Crossplane up-to-date."
else
  $kubectl_cmd apply -f apps/crossplane.yaml
  sleep 10
  $kubectl_cmd wait pod --all --for=condition=Ready --namespace crossplane-system --timeout=300s
  sleep 10
  $kubectl_cmd wait --for condition=established --timeout=60s crd/providerconfigs.keycloak.crossplane.io
fi

echo "########### Installing Keycloak Provider secret ###########"
cat ./apps/keycloak-provider/keycloak-provider-secret.yaml | envsubst | $kubectl_cmd apply --namespace crossplane-system  -f -
$kubectl_cmd apply -f ./apps/keycloak-provider/keycloak-provider.yaml


echo "#################################################"
echo "You're ready to go!"
echo "ArgoCD is ready at https://$ARGOCD_IP:443"
echo "ArgoCD login: admin / $($kubectl_cmd -n argocd get secrets argocd-initial-admin-secret -o json | jq -r .data.password | base64 -d)"
echo "-------------------------------------------------"
echo "Keycloak is ready at http://$KEYCLOAK_IP:$KEYCLOAK_PORT/auth"
echo "Keycloak login: admin / admin"
echo "#################################################"
