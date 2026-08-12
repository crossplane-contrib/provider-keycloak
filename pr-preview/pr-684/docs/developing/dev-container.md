# Dev Container

# Dev Container

The repository ships a [Dev Container](https://containers.dev/) under
`.devcontainer/` that provides a ready-to-use environment to **build the
provider** and **run the end-to-end (uptest) suite** — Go, Docker-in-Docker and
kind included. Use it to avoid installing the toolchain on your host.

## What's inside

| Tool | Version | Source |
|------|---------|--------|
| Go | 1.25 | devcontainer feature |
| Docker (DinD) | latest | devcontainer feature |
| kind, kubectl, helm | pinned in the Makefile | `make` (via `post-create.sh`) |
| make, curl, unzip, jq, envsubst, git, bash-completion | — | provided by the base image |
| goimports | latest | `post-create.sh` |

`kind`, `kubectl` and `helm` are **not** version-pinned in the container.
`post-create.sh` downloads them through the repository's existing Makefile
targets (`build/makelib/k8s_tools.mk`) into `.cache/tools/` and symlinks them
onto `PATH`, so the Makefile stays the single source of truth for their
versions. `yq`, `chainsaw`, `uptest`, `crossplane` and `terraform` are consumed
only by the Makefile and are fetched by it on demand, so they are not placed on
`PATH` here.

## Getting started

Open the folder in VS Code and **"Reopen in Container"**, or use the
[`devcontainer` CLI](https://github.com/devcontainers/cli):

```bash
devcontainer up --workspace-folder .
devcontainer exec --workspace-folder . make generate
```

The first build runs `post-create.sh`, which initializes the `build/` git
submodule, fetches the k8s CLIs via the Makefile, installs `goimports`, and wires
up shell completion. Everything else comes from the base image and the
Go / Docker-in-Docker features.

> Recommended host resources: **4 CPUs / 8 GB RAM / 32 GB disk**. The e2e stack
> (2-node kind cluster + Keycloak + Crossplane + provider) is memory hungry;
> less than this risks OOM-killed pods.

## Build and code generation

```bash
make generate   # regenerate CRDs, API types, controllers, examples
make build      # build provider binary and xpkg package
make test       # unit tests
```

`make generate` uses the committed `config/schema.json` (Keycloak Terraform
provider **v5.8.0**); it does not upgrade the provider dependency.

## End-to-end tests

`make e2e` (= `local-deploy` + `uptest`) builds and deploys the provider into a
fresh kind cluster but **does not install Keycloak**. For a complete, runnable
environment use the dev setup script, then run uptest:

```bash
# 1. kind cluster + Keycloak (via Helm) + Crossplane + locally-built provider
./dev/setup_dev_environment.sh --direct-helm --deploy-local-provider -k 26.6.2

# 2. point kubectl at the kind cluster (fenrir-1)
kind export kubeconfig --name fenrir-1

# 3. run the uptest suite
make uptest KEYCLOAK_VERSION=26.6.2

# 4. clean up
kind delete cluster --name fenrir-1
```

The `--direct-helm` flag installs Keycloak straight from the Helm chart (faster,
no ArgoCD). Drop `--deploy-local-provider` to test the published provider image
instead of your local build.

Render the test cases without a cluster:

```bash
make uptest RENDER_ONLY=true KEYCLOAK_VERSION=26.6.2
```

## Shell completion

`post-create.sh` enables command completion for **kubectl**, **kind** and
**docker** in both `bash` and `zsh`, and adds a `k` alias for `kubectl`. The
configuration is appended to `~/.bashrc` / `~/.zshrc` behind a marker, so
container rebuilds don't duplicate it. It takes effect in any new shell — in your
current one, run `source ~/.bashrc` (or `source ~/.zshrc`).

## Networking note

The setup script exposes Keycloak through a `LoadBalancer` service. By default it
uses **MetalLB** to assign an IP from the kind Docker network; Docker-in-Docker
runs the Docker daemon inside the container, so that IP is reachable and the
script can `curl` Keycloak directly.

Alternatively, run it with `--skip-metal-lb --start-cloud-provider-kind` to use
[`cloud-provider-kind`](https://github.com/kubernetes-sigs/cloud-provider-kind) —
the LoadBalancer implementation for kind clusters — instead of MetalLB. That
binary is not part of the container image; install it separately (for example
`go install sigs.k8s.io/cloud-provider-kind@latest`) if you take this path.

