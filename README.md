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
  name: enable-ess
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
+   name: enable-ess
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
go run cmd/generator/main.go "$PWD"
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

## Regression Tests 
TODO: Add regression test docs

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane-contrib/provider-keycloak/issues).


