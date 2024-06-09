# Provider keycloak

`provider-keycloak` is a [Crossplane](https://crossplane.io/) provider that
is built using [Upjet](https://github.com/crossplane/upjet) code
generation tools and exposes XRM-conformant managed resources for the
keycloak API.

Check out the examples in the `examples` directory for more information on how to use this provider.

## Install

To install the provider, use the following resource definition:

```yaml
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-keycloak
  namespace: crossplane-system
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-keycloak:v0.0.1
``` 


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

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane-contrib/provider-keycloak/issues).
