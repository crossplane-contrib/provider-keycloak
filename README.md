# Provider keycloak

`provider-keycloak` is a [Crossplane](https://crossplane.io/) provider that
is built using [Upjet](https://github.com/upbound/upjet) code
generation tools and exposes XRM-conformant managed resources for the
keycloak API.

Check out the examples in the `examples` directory for more information on how to use this provider.

## Current state
- Currently the provider is built here: https://gitlab.com/corewire/images/provider-keycloak and pushed to the gitlab container registry https://gitlab.com/corewire/images/provider-keycloak/container_registry/4538877
- Soon this provider will be moved to crossplane-contrib and published to upbound. 


## Install

To install the provider, use the following resource definition:

```yaml
---
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: keycloak-provider
  namespace: crossplane-system
  annotations:
    # Only if you use argocd
    argocd.argoproj.io/sync-options: SkipDryRunOnMissingResource=true
spec:
  # currently stored here:  https://gitlab.com/corewire/images/provider-keycloak/container_registry/4538877
  package: registry.gitlab.com/corewire/images/provider-keycloak:v0.0.0-14.gb43e0c4
``` 


## Developing

Run code-generation pipeline:
```console
go run cmd/generator/main.go "$PWD"
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
open an [issue](https://github.com/corewire/provider-keycloak/issues).
