#!/bin/bash
set -eou pipefail

kill_port_forward() {
  echo "Killing port forward"
  echo "Killing ArgoCD port forward with PID $ARGOCD_PORT_FORWARD_PID"
  kill $ARGOCD_PORT_FORWARD_PID
  echo "Killing Keycloak port forward with PID $KEYCLOAK_PORT_FORWARD_PID"
  kill $KEYCLOAK_PORT_FORWARD_PID
}

# trap ctrl-c and call kill_port_forward()
trap kill_port_forward INT


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
kind create cluster --name fenrir-1 --config kind-config.yaml --kubeconfig $HOME/.kube/fenrir-1 || true

echo "Running some commands to make sure the cluster is ready"
export KUBECONFIG=$HOME/.kube/fenrir-1
kubectl cluster-info
kubectl get nodes

echo "Switching context to fenrir-1"
kubectl config use-context kind-fenrir-1

echo "########### Installing ArgoCD ###########"
kubectl create namespace argocd || true
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
echo "* Waiting for ArgoCD to be ready"
kubectl wait pod --all --for=condition=Ready --namespace argocd --timeout=300s


echo "########### Installing Keycloak ###########"
kubectl apply -f apps/keycloak.yaml
sleep 5
kubectl wait pod --all --for=condition=Ready --namespace keycloak --timeout=300s

export KEYCLOAK_PORT=$(kubectl -n keycloak get svc keycloak-keycloakx-http -o json | jq -r .spec.ports[0].port)
export KEYCLOAK_USER=admin
export KEYCLOAK_PASSWORD=admin

echo "########### Installing Crossplane ###########"

kubectl apply -f apps/crossplane.yaml
sleep 10
kubectl wait pod --all --for=condition=Ready --namespace crossplane-system --timeout=300s

echo "########### Installing Keycloak Provider ###########"
cat apps/keycloak-provider/keycloak-provider-secret.yaml | kubectl apply --namespace crossplane-system  -f -
cat apps/keycloak-provider/keycloak-provider.yaml | kubectl apply --namespace crossplane-system  -f -


# Port forward ArgoCD and Keycloak and save the process id
# check if ports 8888 and 8080 are already in use
if lsof -Pi :8888 -sTCP:LISTEN -t >/dev/null 2>&1; then
  echo "Port 8888 is already in use. Aborting."
  exit 1
fi
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
  echo "Port 8080 is already in use. Aborting."
  exit 1
fi
ARGOCD_PORT_FORWARD_PID=$(kubectl port-forward svc/argocd-server -n argocd 8888:443 > /dev/null 2>&1 & echo $!)
KEYCLOAK_PORT_FORWARD_PID=$(kubectl port-forward svc/keycloak-keycloakx-http -n keycloak 8889:$KEYCLOAK_PORT > /dev/null 2>&1 & echo $!)

echo $ARGOCD_PORT_FORWARD_PID
echo $KEYCLOAK_PORT_FORWARD_PID

echo "#################################################"
echo "You're ready to go!"
echo "ArgoCD is ready at https://127.0.0.1:8888"
echo "ArgoCD login: admin / $(kubectl -n argocd get secrets argocd-initial-admin-secret -o json | jq -r .data.password | base64 -d)"
echo "-------------------------------------------------"
echo "Keycloak is ready at http://127.0.0.1:8889/auth"
echo "Keycloak login: admin / admin"
echo "#################################################"
echo "To delete the cluster run: kind delete cluster --name fenrir-1"
echo "#################################################"
echo "You can now run the provider by executing: make run"
echo "#################################################"
echo "Press Ctrl+C to kill the port forwards and exit."

while true; do sleep 10; done


