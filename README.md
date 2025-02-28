# Provider keycloak

`provider-keycloak` is a [Crossplane](https://crossplane.io/) provider that
is built using [Upjet](https://github.com/crossplane/upjet) code
generation tools and exposes XRM-conformant managed resources for the
keycloak API.

Check out the examples in the `examples` directory for more information on how to use this provider.

## Usage 


### Installation

To install the provider, use the following resource definition:

```yaml
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-keycloak
  namespace: crossplane-system
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-keycloak:v1.5.0
``` 

This will install the provider in the `crossplane-system` namespace and install CRDs and controllers for the provider.

#### DeploymentRuntimeConfig

We also support DeploymentRuntimeConfig to enable additional features in the provider.

```yaml
--- 
apiVersion: pkg.crossplane.io/v1beta1
kind: DeploymentRuntimeConfig
metadata:
  name: runtimeconfig-provider-keycloak
spec:
  deploymentTemplate:
    spec:
      selector: {}
      template:
        spec:
          containers:
            - name: package-runtime
              args:
                - --enable-external-secret-stores
```

which can be used in the provider resource as follows:

```diff
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: keycloak-provider
  namespace: crossplane-system
  annotations:
    argocd.argoproj.io/sync-options: SkipDryRunOnMissingResource=true
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-keycloak:v1.5.0
+ runtimeConfigRef:
+   name: runtimeconfig-provider-keycloak
```
(Without the + signs of course)



### Configuration 

- For each keycloak instance you need one or more `ProviderConfig` resources.
- The `ProviderConfig` resource is used to store the keycloak API server URL, credentials, and other configuration details that are required to connect to the keycloak API server.
- Here is an example of a `ProviderConfig` resource:

```yaml
---
apiVersion: keycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: keycloak-provider-config
spec:
  credentials:
    source: Secret
    secretRef:
      name: keycloak-credentials
      key: credentials
      namespace: crossplane-system
---
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-credentials
  namespace: crossplane-system
  labels: 
    type: provider-credentials
type: Opaque
stringData:
  credentials: |
    {
      "client_id":"admin-cli",
      "username": "admin",
      "password": "admin",
      "url": "https://keycloak.example.com",
      "base_path": "/auth",
      "realm": "master"
    }
```

The secret `keycloak-credentials` contains the keycloak API server URL, credentials, and other configuration details that are required to connect to the keycloak API server. **It supports the same fields as the [terraform provider configuration](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs#argument-reference)**

As an alternative to using the embedded JSON format shown above, you can also place settings in a plain Kubernetes secret like this:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-credentials
  namespace: crossplane-system
  labels:
    type: provider-credentials
type: Opaque
stringData:
  client_id: "admin-cli"
  username: "admin"
  password: "admin"
  url: "https://keycloak.example.com"
  base_path: "/auth"
  realm: "master"
```


### Custom Resource Definitions

You can explore the available custom resources: 
- [Upbound marketplace site](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/)
- `kubectl get crd | grep keycloak.crossplane.io` to list all the CRDs provided by the provider
- `kubectl explain <CRD_NAME>` for docs on the CLI
- You can also see the CRDs in the `package/crds` directory


### Functions and Compositions: 

- [function-keycloak-builtin-objects](https://gitlab.com/corewire/images/crossplane/function-keycloak-builtin-objects) - The function is used to import the builtin objects of a keycloak, e.g. clients and roles. Since v3.0 it also offers the possibility to adapt some default config. Everything you need to know is in the README of the repository.  



## Developing

Run code-generation pipeline:
```console
go install golang.org/x/tools/cmd/goimports@latest
go run cmd/generator/main.go "$(pwd)"
```

Checkout sub-repositories:

```console
make submodules
```

Execute code generation:

```console
make generate
```

Run against a Kubernetes cluster:

```console
make run
```

Build, push, and install:

```console
make all
```

Build binary:

```console
make build
```

### Build from custom Terraform Provider

If you want to build this crossplane provider on top of a forked `terraform-provider-keycloak` follow these instructions:

1. Execute code generation:
```
TERRAFORM_PROVIDER_REPO=https://github.com/<owner>/terraform-provider-keycloak \
TERRAFORM_PROVIDER_VERSION=1.0.0 \
make generate
```
**Hint:** `TERRAFORM_PROVIDER_VERSION` must be a Release. Releases can be found here: `https://github.com/<owner>/terraform-provider-keycloak/releases`. 
Every ReleaseName should have the prefix "v" (i.e 'v1.0.0'). But if you specify the `TERRAFORM_PROVIDER_VERSION` you need to 
skip that prefix (i.e. '1.0.0')

2. Use forked repo as go dependency:
```
go mod edit -replace="github.com/keycloak/terraform-provider-keycloak@v0.0.0-20241206084240-f87470c95855=github.com/<owner>/terraform-provider-keycloak@v1.0.0"
go mod tidy
```
**Hint:** You can also specify the version as `github.com/<owner>/terraform-provider-keycloak@v0.0.0-<timestamp>-<commitHash>`

### Build and publish to custom repo

Install up cli: https://docs.upbound.io/reference/cli/

Git tag with the version that should be published:
```console
git tag v<VersionNumber>
```

Create a release branch with git:
```console
git checkout -b release-v<VersionNumber>
```

Ensure that you ran `make generate` and `make build`

**Hint:** If you want to build a specific platform you can do this with:
```console
PLATFORMS=linux_amd64 make build
```

Login
```console
up login -t <TOKEN>
```

Publish
```console
PLATFORMS=linux_amd64 \
XPKG_REG_ORGS=xpkg.upbound.io/<owner> \
XPKG_REG_ORGS_NO_PROMOTE=xpkg.upbound.io/<owner> \
make publish
```

### Local Environment 
Execute setup script which creates a KIND Cluster
and installs crossplane, keycloak and the official crossplane provider
via ArgoCD (for more options run script with `--help`)
```console
./dev/setup_dev_environment.sh 
```
**Hint**: If you are using rootless docker you can add the flags `--skip-metal-lb` 
and `--start-cloud-provider-kind` (how to install cloud-provider-kind [see here](https://github.com/kubernetes-sigs/cloud-provider-kind))

Use created file from KIND as kubeconfig `~/.kube/<clustername>` 

For debugging local source code you can run the script with `--use-local-provider` flag
this will scale down the crossplane provider which is running in the cluster
and then start your local crossplane provider instance.

### Alternative Local Environment

This make target creates a KIND Cluster
and installs crossplane and the crossplane provider 
from current sources. But no keycloak deployment is stared.
```console
make local-deploy
```

## Regression Tests

### Run Tests
Follow the following steps to run end to end tests:


Start local dev cluster
```console
./dev/setup_dev_environment.sh --deploy-local-provider
```
**Hint**: If you are using rootless docker you can add the flags `--skip-metal-lb`
and `--start-cloud-provider-kind` (how to install cloud-provider-kind [see here](https://github.com/kubernetes-sigs/cloud-provider-kind))

Use created file from KIND as kubeconfig `~/.kube/<clustername>`

Run tests
```console
make uptest
```

### Add Tests
New TestCases are added to this file `cluster/test/cases.txt`.
Every resource that is necessary (i.e. Secrets) but no ManagedResource has to be created within this file `dev/demos/basic/000-init.yaml`

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane-contrib/provider-keycloak/issues).


