#!/usr/bin/env bash
# Sets up what the build and e2e flows need but the devcontainer base image
# doesn't already provide. Tool versions live in the Makefile, never here.
set -euo pipefail

BIN=/usr/local/bin

# make, curl, unzip, jq, envsubst, git and bash-completion already ship with the
# pinned base image (devcontainers/base:bookworm), so there is no apt step. Add
# one back here if that image is swapped for a slimmer one.

echo "==> Initializing git submodules (build/ is required by the Makefile)"
git config --global --add safe.directory "$(pwd)" || true
git submodule update --init --recursive

echo "==> Downloading k8s tools via the Makefile (single source of version truth)"
# Read the version-pinned tool paths from the Makefile, let its targets download
# them, then symlink to canonical names so the e2e script finds them on PATH.
# yq stays Makefile-internal ($(YQ)), so it is not symlinked here.
eval "$(make -s build.vars | grep -E '^(KIND|KUBECTL|HELM)=')"
make -s "$KIND" "$KUBECTL" "$HELM"
sudo ln -sf "$KIND" "${BIN}/kind"
sudo ln -sf "$KUBECTL" "${BIN}/kubectl"
sudo ln -sf "$HELM" "${BIN}/helm"

echo "==> Installing goimports (used by make generate / check-diff)"
go install golang.org/x/tools/cmd/goimports@latest

echo "==> Configuring shell completion for kubectl, kind and docker (bash + zsh)"
# The snippets live in .devcontainer/completions/; each rc file just sources the
# matching one. Marker-guarded so container rebuilds don't append it twice.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MARK="# >>> provider-keycloak devcontainer completions >>>"
END_MARK="# <<< provider-keycloak devcontainer completions <<<"

link_completion() {
  local rc="$1" snippet="$2" always="${3:-}"
  # Always set up .bashrc (create if missing); only touch .zshrc if it exists.
  [ "$always" = "always" ] || [ -e "$rc" ] || return 0
  grep -qF "$MARK" "$rc" 2>/dev/null && return 0
  {
    echo ""
    echo "$MARK"
    echo "[ -r \"$snippet\" ] && source \"$snippet\""
    echo "$END_MARK"
  } >>"$rc"
}

link_completion "$HOME/.bashrc" "$SCRIPT_DIR/completions/completions.bash" always
link_completion "$HOME/.zshrc"  "$SCRIPT_DIR/completions/completions.zsh"

echo ""
echo "==> Versions:"
go version
docker --version || true
kind version
kubectl version --client 2>/dev/null | head -1 || true
helm version --short || true
jq --version

cat <<'EOF'

Dev container ready.

Build & codegen:
  make generate        # regenerate CRDs/types (needs network; uses config/schema.json @ v5.8.0)
  make build           # build the provider binary + xpkg image
  make test            # unit tests

End-to-end (kind + Keycloak + Crossplane + provider):
  # One-shot: cluster + Keycloak (Helm) + Crossplane + locally-built provider
  ./dev/setup_dev_environment.sh --direct-helm --deploy-local-provider -k 26.6.2

  # Export the kubeconfig for the kind cluster (fenrir-1) so you can use kubectl
  kind export kubeconfig --name fenrir-1

  # Then run the uptest suite (includes the new RealmLocalization cases)
  make uptest KEYCLOAK_VERSION=26.6.2

  # Tear down
  kind delete cluster --name fenrir-1

Notes:
  * Docker-in-Docker is enabled; kind runs its nodes as containers inside this
    container, so MetalLB LoadBalancer IPs are reachable from here.
  * `make e2e` builds+deploys the provider and runs uptest but does NOT install
    Keycloak — use the setup script above for a full environment.
EOF
