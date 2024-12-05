
prereqs: ctlptl, kind, kubectl

## Keycloak up and running

This is an express installation of keycloak on new kind cluster.

``` sh
ctlptl apply -f kind-kustomize/cluster/cluster.yaml

kubectl apply -f kind-kustomize/keycloak/keycloak.yaml

kubectl port-forward -n keycloak svc/keycloak 8080:80
```


```
❯ ctlptl apply -f kind-kustomize/cluster/cluster.yaml
No kind clusters found.
Creating cluster "provider-keycloak-cluster" ...
 ✓ Ensuring node image (kindest/node:v1.31.0) 🖼
 ✓ Preparing nodes 📦 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Joining worker nodes 🚜
Set kubectl context to "kind-provider-keycloak-cluster"
You can now use your cluster with:

kubectl cluster-info --context kind-provider-keycloak-cluster

Have a question, bug, or feature request? Let us know! https://kind.sigs.k8s.io/#community 🙂
Switched to context "kind-provider-keycloak-cluster".
 🔌 Connected cluster kind-provider-keycloak-cluster to registry ctlptl-registry at localhost:52145
 👐 Push images to the cluster like 'docker push localhost:52145/alpine'
cluster.ctlptl.dev/kind-provider-keycloak-cluster created

> kubectl config get-contexts
CURRENT   NAME                             CLUSTER                          AUTHINFO                         NAMESPACE
*         kind-provider-keycloak-cluster   kind-provider-keycloak-cluster   kind-provider-keycloak-cluster

```

```
> kubectl apply -f kind-kustomize/keycloak/keycloak.yaml
namespace/keycloak created
configmap/keycloak-cm created
service/keycloak created
deployment.apps/kc created

```

```
> kubectl port-forward -n keycloak svc/keycloak 8080:80
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
Handling connection for 8080
```

When surfing into the keycloak UI at http://localhost:8080 you can logon as admin/admin. You are then prompted to replace the temporary admin account with a permanent one. For the purpose of demonstrating or getting started with this crossplane provider you can skip this step. Make sure the new user can log on and has the correct access (typically the admin role) before deleting the temporary user.

![An orange banner at the top urging the temporary user to be replaced](replace-user-banner.png)

Refer to the keycloak documentation on how to best harden security for your setup of keycloak and consider using an external database. https://www.keycloak.org/docs/latest/server_admin/#proc-creating-user_server_administration_guide


## Installing crossplane

This is described in better detail on the crossplane main repository, and on their webpage: https://docs.crossplane.io/latest/software/install/

``` sh
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm install crossplane --namespace crossplane-system --create-namespace crossplane-stable/crossplane 

``` 

This will bootstrap a working crossplane client in the master realm with the admin role. You should consider learning to set up a similar client through the UI or through the API in a manner which fits your security practices.

``` sh
# creates a config map with the script to run
kubectl create configmap client-script -n keycloak --from-file=kind-kustomize/crossplane/create-client.sh

# creates a job to run the script from within kubernetes.
kubectl apply -f kind-kustomize/crossplane/create-client.yaml

```

This will create the keycloak-provider and configure it to use a client withing the master realm to perform actions there.

The settings for the client will also make it appear as a service-account user in the realm.

![Displays the crossplane service-account user](assets/service-account-crossplane.png)
