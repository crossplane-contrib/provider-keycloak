#!/usr/bin/env bash
set -euox pipefail

read -r -p "Lower case provider name (ex. github): " PROVIDER_NAME_LOWER
read -r -p "Normal case provider name (ex. GitHub): " PROVIDER_NAME_NORMAL
read -r -p "Organization (ex. upbound, my-org-name): " ORGANIZATION_NAME

REPLACE_FILES='./* ./.github :!build/** :!go.* :!hack/prepare.sh'
# shellcheck disable=SC2086
git grep -l 'template' -- ${REPLACE_FILES} | xargs sed -i.bak "s/upjet-provider-template/provider-${PROVIDER_NAME_LOWER}/g"
# shellcheck disable=SC2086
git grep -l 'template' -- ${REPLACE_FILES} | xargs sed -i.bak "s/template/${PROVIDER_NAME_LOWER}/g"
# shellcheck disable=SC2086
git grep -l "upbound/provider-${PROVIDER_NAME_LOWER}" -- ${REPLACE_FILES} | xargs sed -i.bak "s|upbound/provider-${PROVIDER_NAME_LOWER}|${ORGANIZATION_NAME}/provider-${PROVIDER_NAME_LOWER}|g"
# shellcheck disable=SC2086
git grep -l 'Template' -- ${REPLACE_FILES} | xargs sed -i.bak "s/Template/${PROVIDER_NAME_NORMAL}/g"
# We need to be careful while replacing "template" keyword in go.mod as it could tamper
# some imported packages under require section.
sed -i.bak "s|upbound/upjet-provider-template|${ORGANIZATION_NAME}/provider-${PROVIDER_NAME_LOWER}|g" go.mod
sed -i.bak "s|PROJECT_REPO := github.com/upbound/|PROJECT_REPO := github.com/${ORGANIZATION_NAME}/|g" Makefile

# Clean up the .bak files created by sed
git clean -fd

git mv "internal/clients/template.go" "internal/clients/${PROVIDER_NAME_LOWER}.go"
git mv "cluster/images/upjet-provider-template" "cluster/images/provider-${PROVIDER_NAME_LOWER}"

# We need to remove this api folder otherwise first `make generate` fails with
# the following error probably due to some optimizations in go generate with v1.17:
# generate: open /Users/hasanturken/Workspace/crossplane-contrib/upjet-provider-template/apis/null/v1alpha1/zz_generated.deepcopy.go: no such file or directory
rm -rf apis/null