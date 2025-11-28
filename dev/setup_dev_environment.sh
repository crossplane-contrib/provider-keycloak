#!/bin/bash
set -eo pipefail

SCRIPT_DIR=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)

# Default variable values
CLUSTER_NAME="fenrir-1"
KEYCLOAK_VERSION="26.4.4"
skipmetallb=false
runcloudproviderkind=false
uselocalprovider=false
deploylocalprovider=false
# Function to display script usage
usage() {
 echo "Usage: $0 [OPTIONS]"
 echo "Options:"
 echo " -h, --help                       Display this help message"
 echo " -c, --cluster-name               Name of the Cluster"
 echo " -s, --skip-metal-lb              Do not install MetalLB"
 echo " -p, --start-cloud-provider-kind  Run 'cloud-provider-kind' with sudo as Background task due to rootless docker (metal lb wont work) + mounting user docker socket to root docker socket"
 echo " -l, --use-local-provider         Use local provider (Scales down 'provider-keycloak')"
 echo " -d, --deploy-local-provider      Deploy local provider"
 echo " -k, --keycloak-version           Keycloak Version"
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
      -p | --start-cloud-provider-kind)
        runcloudproviderkind=true
        ;;
      -l | --use-local-provider)
        uselocalprovider=true
        ;;
      -d | --deploy-local-provider)
        deploylocalprovider=true
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
      -k | --keycloak-version*)
        if ! has_argument $@; then
          echo "Keycloakversion not specified." >&2
          usage
          exit 1
        fi

        KEYCLOAK_VERSION=$(extract_argument $@)

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
echo "Keycloak version: $KEYCLOAK_VERSION"

echo "########### Checking dependencies ###########"
command -v docker >/dev/null 2>&1 || { echo >&2 "Docker is required but not installed.  Aborting."; exit 1; }
command -v kind >/dev/null 2>&1 || { echo >&2 "Kind is required but not installed.  Aborting."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo >&2 "Kubectl is required but not installed.  Aborting."; exit 1; }
command -v jq >/dev/null 2>&1 || { echo >&2 "jq is required but not installed.  Aborting."; exit 1; }
command -v envsubst >/dev/null 2>&1 || { echo >&2 "envsubst is required but not installed.  Aborting."; exit 1; }
command -v base64 >/dev/null 2>&1 || { echo >&2 "base64 is required but not installed.  Aborting."; exit 1; }
command -v sed >/dev/null 2>&1 || { echo >&2 "sed is required but not installed.  Aborting."; exit 1; }
if [[ "$runcloudproviderkind" == "true" ]]; then
command -v cloud-provider-kind >/dev/null 2>&1 || { echo >&2 "cloud-provider-kind is required but not installed.  Aborting."; exit 1; }
fi
echo "All dependencies are installed."

# check if user can run docker without sudo, if not create an alias for sudo docker for this session
if ! docker ps >/dev/null 2>&1; then
  echo "Sudo required for docker."
  sudo_prefix='sudo'
# if using rootless podman, wrap call into own cgroup - runRoot: /run/user/
elif [[ "$KIND_EXPERIMENTAL_PROVIDER" = podman ]] && podman info | grep -q 'runRoot: /run/user/'; then
  sudo_prefix="systemd-run --user --scope -p Delegate=yes"
else
  sudo_prefix=''
fi

echo "########### Setup Cluster ###########"
if $sudo_prefix kind get clusters | grep "$CLUSTER_NAME" >/dev/null 2>&1; then
  echo "$CLUSTER_NAME cluster already exists"
else
  echo "Creating cluster"
  old_context=$(kubectl config current-context || echo "notset")
  $sudo_prefix kind create cluster --name $CLUSTER_NAME --config ${SCRIPT_DIR}/kind-config.yaml --kubeconfig $HOME/.kube/$CLUSTER_NAME
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
  cat ${SCRIPT_DIR}/metallb/pools.yaml | envsubst | $kubectl_cmd apply -f - && break  # Break the loop if command succeeds
  echo "** still waiting for metallb resources to be ready"
  sleep 1
done
fi

if [[ "$runcloudproviderkind" == "true" ]]; then
echo "Starting cloud-provider-kind with sudo as BackgroundTask"
export CLOUD_PROVIDER_KIND_LOGS=$(mktemp)
echo "Cloud-Provider-Kind Logs are here: tail -f '$CLOUD_PROVIDER_KIND_LOGS'"
sudo echo -n ""
sudo ln -s /run/user/1000/docker.sock /var/run/docker.sock || true
sudo cloud-provider-kind -v 0  > $CLOUD_PROVIDER_KIND_LOGS 2>&1 &

## Define the cleanup function
cleanup() {
  parentProcess=$$
  echo ""
  echo "Killing children of ${parentProcess} (Cloud-Provider-Kind), which runs in background"
  pkill -SIGTERM -P $$ cloud-provider
  sleep 2
}

## Register the cleanup function to be executed on script interruption
trap cleanup SIGINT SIGTERM
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

echo "########### Installing Keycloak & OpenLdap ###########"
$kubectl_cmd apply -f ${SCRIPT_DIR}/apps/open-ldap.yaml
sed "s/{{keycloak-version}}/${KEYCLOAK_VERSION}/g" ${SCRIPT_DIR}/apps/keycloak.yaml | $kubectl_cmd apply -f -


echo "* Waiting for Keycloak to be ready"
$kubectl_cmd wait applications.argoproj.io  --namespace argocd keycloak --for=create --for=jsonpath='{.status.health.status}'=Healthy --for=jsonpath='{.status.sync.status}'=Synced  --timeout=300s
$kubectl_cmd wait pod --namespace keycloak --selector="app.kubernetes.io/name=keycloakx" --for=condition=Ready --timeout=300s

while [[ -z $($kubectl_cmd get svc -n keycloak keycloak-keycloakx-http -o jsonpath="{.status.loadBalancer.ingress}" 2>/dev/null) ]]; do
  echo "** still waiting for service keycloak/keycloak-keycloakx-http to get ingress"
  sleep 1
done

export KEYCLOAK_IP=$($kubectl_cmd -n keycloak get svc keycloak-keycloakx-http -o json | jq -r .status.loadBalancer.ingress[0].ip)
export KEYCLOAK_PORT=$($kubectl_cmd -n keycloak get svc keycloak-keycloakx-http -o json | jq -r '.spec.ports[] | select(.name == "http").port')
export KEYCLOAK_USER=admin
export KEYCLOAK_PASSWORD=admin


echo "########### Installing Crossplane ###########"
$kubectl_cmd apply -f ${SCRIPT_DIR}/apps/crossplane.yaml

echo "* Waiting for Crossplane to be ready"
$kubectl_cmd wait applications.argoproj.io  --namespace argocd crossplane-system --for=create --for=jsonpath='{.status.health.status}'=Healthy --for=jsonpath='{.status.sync.status}'=Synced --timeout=300s
$kubectl_cmd wait pod --namespace crossplane-system  --selector="app=crossplane" --for=condition=Ready --timeout=300s
$kubectl_cmd wait pod --namespace crossplane-system  --selector="app=crossplane-rbac-manager" --for=condition=Ready --timeout=300s

echo "########### Installing Keycloak Provider ###########"

$kubectl_cmd create namespace dev -o yaml --dry-run=client | $kubectl_cmd apply -f -
if [[ "$deploylocalprovider" == "false" ]]; then
  cat "${SCRIPT_DIR}/apps/keycloak-provider/keycloak-provider-secret.yaml" | envsubst | $kubectl_cmd apply -f -
  $kubectl_cmd apply -f ${SCRIPT_DIR}/apps/keycloak-provider/keycloak-provider.yaml
else
  export OLD_KEYCLOAK_IP=$KEYCLOAK_IP
  export KEYCLOAK_IP=keycloak-keycloakx-http.keycloak.svc.cluster.local
  cat "${SCRIPT_DIR}/apps/keycloak-provider/keycloak-provider-secret.yaml" | envsubst | $kubectl_cmd apply  -f -
  export KEYCLOAK_IP=$OLD_KEYCLOAK_IP

  echo "Deploy local source code as provider 'provider-keycloak'"

  # Hint: crossplane pod´s filesystem based cache for providers is patched with local built provider
  KIND_CLUSTER_NAME=$CLUSTER_NAME make local-deploy-provider
  
  echo "* Restarting provider pod to pick up changes"
  $kubectl_cmd delete pods -n crossplane-system -l pkg.crossplane.io/provider=provider-keycloak --ignore-not-found=true
  sleep 5
fi

echo "* Waiting for Keycloak Provider to be ready"
$kubectl_cmd wait provider.pkg.crossplane.io/provider-keycloak --for=create --timeout=300s
$kubectl_cmd wait provider.pkg.crossplane.io/provider-keycloak --for=condition=Healthy --timeout=300s
$kubectl_cmd wait deployment --all --namespace crossplane-system --for=condition=Available
$kubectl_cmd wait pod --all --namespace crossplane-system --for=condition=Ready --timeout=300s
$kubectl_cmd wait crd/providerconfigs.keycloak.crossplane.io --for condition=established --timeout=60s

$kubectl_cmd apply -f ${SCRIPT_DIR}/apps/keycloak-provider/keycloak-provider-config.yaml

if [[ "$uselocalprovider" == "true" ]]; then
  echo "Scaling down 'provider-keycloak' to use local provider"
  $kubectl_cmd patch DeploymentRuntimeConfig runtimeconfig-provider-keycloak --type='merge' -p '{"spec":{"deploymentTemplate":{"spec":{"replicas":0}}}}'

  echo "* Waiting for Keycloak Provider to be removed"
  $kubectl_cmd wait pod --namespace crossplane-system --selector="pkg.crossplane.io/provider=provider-keycloak" --for=delete --timeout=300s
  $kubectl_cmd wait deployment --all --namespace crossplane-system --for=condition=Available
  $kubectl_cmd apply -f ${SCRIPT_DIR}/../package/crds
fi

echo "#################################################"
echo "You're ready to go!"
echo "ArgoCD is ready at https://$ARGOCD_IP:443"
echo "ArgoCD login: admin / $($kubectl_cmd -n argocd get secrets argocd-initial-admin-secret -o json | jq -r .data.password | base64 -d)"
echo "-------------------------------------------------"
echo "Keycloak is ready at http://$KEYCLOAK_IP:$KEYCLOAK_PORT/"
echo "Keycloak login: admin / admin"
echo "#################################################"

if [[ "$runcloudproviderkind" == "true" ]]; then
echo "Cloud-Provider-Kind Logs are here: tail -f $CLOUD_PROVIDER_KIND_LOGS"

# Do not finish script, so that Cloud-Provider-Kind keeps running!
tail -f /dev/null
fi
