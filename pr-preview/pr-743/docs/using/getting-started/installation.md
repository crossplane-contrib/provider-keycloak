# Installation

## Prerequisites

- A Kubernetes cluster with [Crossplane](https://docs.crossplane.io/latest/software/install/) installed
- `kubectl` configured to access your cluster
- A running Keycloak instance

## Install the Provider

Apply the following manifest to install provider-keycloak:

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-keycloak
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-keycloak:<version>
```

Replace `<version>` with the desired release version (e.g., `v2.22.0`). See [GitHub Releases](https://github.com/crossplane-contrib/provider-keycloak/releases) for available versions.

## Verify Installation

```bash
# Check that the provider pod is running
kubectl get pods -n crossplane-system | grep keycloak

# Check that CRDs are installed
kubectl get crd | grep keycloak.crossplane.io
```

## DeploymentRuntimeConfig (Optional)

To customize the provider deployment (e.g., enable external secret stores), create a `DeploymentRuntimeConfig`:

```yaml
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

Reference it in the Provider resource:

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-keycloak
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-keycloak:<version>
  runtimeConfigRef:
    name: runtimeconfig-provider-keycloak
```

## Next Steps

- [Configure credentials](./configuration.md) to connect to your Keycloak instance
- [Create your first realm](./first-realm.md)

