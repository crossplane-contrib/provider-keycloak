# ====================================================================================
# Setup Project

PROJECT_NAME ?= provider-keycloak
PROJECT_REPO ?= github.com/crossplane-contrib/$(PROJECT_NAME)

export TERRAFORM_VERSION ?= 1.5.7

# Do not allow a version of terraform greater than 1.5.x, due to versions 1.6+ being
# licensed under BSL, which is not permitted.
TERRAFORM_VERSION_VALID := $(shell [ "$(TERRAFORM_VERSION)" = "`printf "$(TERRAFORM_VERSION)\n1.6" | sort -V | head -n1`" ] && echo 1 || echo 0)
#TERRAFORM_VERSION_VALID := 1

export TERRAFORM_PROVIDER_SOURCE ?= keycloak/keycloak
export TERRAFORM_PROVIDER_REPO ?= https://github.com/keycloak/terraform-provider-keycloak
# renovate: datasource=github-releases depName=keycloak/terraform-provider-keycloak
export TERRAFORM_PROVIDER_VERSION ?= 5.8.0
export TERRAFORM_PROVIDER_DOWNLOAD_NAME ?= terraform-provider-keycloak
export TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX ?= ${TERRAFORM_PROVIDER_REPO}/releases/download/v$(TERRAFORM_PROVIDER_VERSION)
export TERRAFORM_NATIVE_PROVIDER_BINARY ?= terraform-provider-keycloak_v$(TERRAFORM_PROVIDER_VERSION)
export TERRAFORM_DOCS_PATH ?= docs/resources
export TERRAFORM_FILE_MIRROR ?= .terraform.d/plugins
export TERRAFORM_FILE_MIRROR_REPO ?= ${TERRAFORM_FILE_MIRROR}/registry.terraform.io

export GOLANGCILINT_VERSION ?= 2.7.2

PLATFORMS ?= linux_amd64 linux_arm64

# -include will silently skip missing files, which allows us
# to load those files with a target in the Makefile. If only
# "include" was used, the make command would fail and refuse
# to run a target until the include commands succeeded.
-include build/makelib/common.mk

# ====================================================================================
# Setup Output

-include build/makelib/output.mk

# ====================================================================================
# Setup Go

# Set a sane default so that the nprocs calculation below is less noisy on the initial
# loading of this file
NPROCS ?= 1

# each of our test suites starts a kube-apiserver and running many test suites in
# parallel can lead to high CPU utilization. by default we reduce the parallelism
# to half the number of CPU cores.
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))

GO_REQUIRED_VERSION ?= 1.25
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/provider $(GO_PROJECT)/cmd/generator
GO_LDFLAGS += -X $(GO_PROJECT)/internal/version.Version=$(VERSION)
GO_SUBDIRS += cmd internal apis generate
-include build/makelib/golang.mk

# ====================================================================================
# Setup Kubernetes tools
KUBECTL_VERSION ?= v1.32.2
KIND_VERSION = v0.32.0
UP_VERSION = v0.38.4
UP_CHANNEL = stable
UPTEST_VERSION = v2.2.0
-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup Images

REGISTRY_ORGS ?= xpkg.upbound.io/crossplane-contrib
IMAGES = $(PROJECT_NAME)
-include build/makelib/imagelight.mk

# ====================================================================================
# Setup XPKG

XPKG_REG_ORGS ?= xpkg.upbound.io/crossplane-contrib
# NOTE(hasheddan): skip promoting on xpkg.upbound.io as channel tags are
# inferred.
XPKG_REG_ORGS_NO_PROMOTE ?= xpkg.upbound.io/crossplane-contrib
XPKGS = $(PROJECT_NAME)
-include build/makelib/xpkg.mk

# ====================================================================================
# Fallthrough

# run `make help` to see the targets and options

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make

# NOTE(hasheddan): we force image building to happen prior to xpkg build so that
# we ensure image is present in daemon.
xpkg.build.upjet-provider-template: do.build.images

# NOTE(hasheddan): we ensure up is installed prior to running platform-specific
# build steps in parallel to avoid encountering an installation race condition.
build.init: $(UP) check-terraform-version $(CROSSPLANE_CLI)


# ====================================================================================
# Setup Terraform for fetching provider schema
TERRAFORM := $(TOOLS_HOST_DIR)/terraform-$(TERRAFORM_VERSION)
TERRAFORM_WORKDIR := $(WORK_DIR)/terraform
TERRAFORM_PROVIDER_SCHEMA := config/schema.json

check-terraform-version:
ifneq ($(TERRAFORM_VERSION_VALID),1)
	$(error invalid TERRAFORM_VERSION $(TERRAFORM_VERSION), must be less than 1.6.0 since that version introduced a not permitted BSL license))
endif

$(TERRAFORM): check-terraform-version
	@$(INFO) installing terraform $(HOSTOS)-$(HOSTARCH)
	@mkdir -p $(TOOLS_HOST_DIR)/tmp-terraform
	@curl -fsSL https://releases.hashicorp.com/terraform/$(TERRAFORM_VERSION)/terraform_$(TERRAFORM_VERSION)_$(SAFEHOST_PLATFORM).zip -o $(TOOLS_HOST_DIR)/tmp-terraform/terraform.zip
	@unzip $(TOOLS_HOST_DIR)/tmp-terraform/terraform.zip -d $(TOOLS_HOST_DIR)/tmp-terraform
	@mv $(TOOLS_HOST_DIR)/tmp-terraform/terraform $(TERRAFORM)
	@rm -fr $(TOOLS_HOST_DIR)/tmp-terraform
	@$(OK) installing terraform $(HOSTOS)-$(HOSTARCH)

$(TERRAFORM_PROVIDER_SCHEMA): $(TERRAFORM)
	@$(INFO) generating provider schema for $(TERRAFORM_PROVIDER_SOURCE) $(TERRAFORM_PROVIDER_VERSION)
	@mkdir -p $(TERRAFORM_WORKDIR)
	@$(MAKE) download-tf-provider-platforms
	@echo '{"terraform":[{"required_providers":[{"provider":{"source":"'"$(TERRAFORM_PROVIDER_SOURCE)"'","version":"'"$(TERRAFORM_PROVIDER_VERSION)"'"}}],"required_version":"'"$(TERRAFORM_VERSION)"'"}]}' > $(TERRAFORM_WORKDIR)/main.tf.json
	@echo 'provider_installation { filesystem_mirror { path = "$(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR)" include = ["*/*/*"] } }' > $(TERRAFORM_WORKDIR)/config.tfrc
	@TF_CLI_CONFIG_FILE=$(TERRAFORM_WORKDIR)/config.tfrc $(TERRAFORM) -chdir=$(TERRAFORM_WORKDIR) init -no-color > $(TERRAFORM_WORKDIR)/terraform-logs.txt 2>&1
	@TF_CLI_CONFIG_FILE=$(TERRAFORM_WORKDIR)/config.tfrc $(TERRAFORM) -chdir=$(TERRAFORM_WORKDIR) providers schema -json=true > $(TERRAFORM_PROVIDER_SCHEMA) 2>> $(TERRAFORM_WORKDIR)/terraform-logs.txt
	@$(OK) generating provider schema for $(TERRAFORM_PROVIDER_SOURCE) $(TERRAFORM_PROVIDER_VERSION)

download-tf-provider-platforms: $(foreach p,$(PLATFORMS), download-tf-provider-platform.$(p))

download-tf-provider-platform.%:
	@$(MAKE) download-tf-provider-platform PLATFORM=$*

download-tf-provider-platform:
	@PLATFORM=$*
	@$(INFO) downloading provider for platform $(PLATFORM)
	@mkdir -p $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}
	@curl -fsSL ${TERRAFORM_PROVIDER_DOWNLOAD_URL_PREFIX}/${TERRAFORM_PROVIDER_DOWNLOAD_NAME}_${TERRAFORM_PROVIDER_VERSION}_${PLATFORM}.zip -o $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}/terraform.zip
	@unzip -o -qq $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}/terraform.zip -d $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}/
	@rm $(TERRAFORM_WORKDIR)/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(TERRAFORM_PROVIDER_VERSION)/${PLATFORM}/terraform.zip

pull-docs:
	@if [ ! -d "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)" ]; then \
  		mkdir -p "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)" && \
		git clone -c advice.detachedHead=false --depth 1 --filter=blob:none --branch "v$(TERRAFORM_PROVIDER_VERSION)" --sparse "$(TERRAFORM_PROVIDER_REPO)" "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)"; \
	fi
	@git -C "$(WORK_DIR)/$(TERRAFORM_PROVIDER_SOURCE)" sparse-checkout set "$(TERRAFORM_DOCS_PATH)"

generate.init: $(TERRAFORM_PROVIDER_SCHEMA) pull-docs

.PHONY: $(TERRAFORM_PROVIDER_SCHEMA) pull-docs check-terraform-version
# ====================================================================================
# Targets

# NOTE: the build submodule currently overrides XDG_CACHE_HOME in order to
# force the Helm 3 to use the .work/helm directory. This causes Go on Linux
# machines to use that directory as the build cache as well. We should adjust
# this behavior in the build submodule because it is also causing Linux users
# to duplicate their build cache, but for now we just make it easier to identify
# its location in CI so that we cache between builds.
go.cachedir:
	@go env GOCACHE

# Generate a coverage report for cobertura applying exclusions on
# - generated file
cobertura:
	@cat $(GO_TEST_OUTPUT)/coverage.txt | \
		grep -v zz_ | \
		$(GOCOVER_COBERTURA) > $(GO_TEST_OUTPUT)/cobertura-coverage.xml

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# This is for running out-of-cluster locally, and is for convenience. Running
# this make target will print out the command which was used. For more control,
# try running the binary directly with different arguments.
run: go.build
	@$(INFO) Running Crossplane locally out-of-cluster . . .
	@# To see other arguments that can be provided, run the command with --help instead
	UPBOUND_CONTEXT="local" $(GO_OUT_DIR)/provider --debug

# ====================================================================================
# End to End Testing
CHAINSAW_VERSION = 0.2.12
CROSSPLANE_VERSION = 2.0.2
CROSSPLANE_CLI_VERSION = v2.0.2
CROSSPLANE_NAMESPACE = crossplane-system
CROSSPLANE_CHART_DIR := $(TOOLS_HOST_DIR)/crossplane-chart-$(CROSSPLANE_VERSION)
CROSSPLANE_CHART := $(CROSSPLANE_CHART_DIR)/Chart.yaml
-include build/makelib/local.xpkg.mk
-include build/makelib/controlplane.mk

$(CROSSPLANE_CLI):
	@$(INFO) installing Crossplane CLI $(CROSSPLANE_CLI_VERSION)
	@rm -rf $(TOOLS_HOST_DIR)/tmp-crossplane-cli
	@mkdir -p $(dir $(CROSSPLANE_CLI)) $(TOOLS_HOST_DIR)/tmp-crossplane-cli
	@GOBIN=$(TOOLS_HOST_DIR)/tmp-crossplane-cli go install github.com/crossplane/crossplane/v2/cmd/crank@$(CROSSPLANE_CLI_VERSION)
	@mv $(TOOLS_HOST_DIR)/tmp-crossplane-cli/crank $(CROSSPLANE_CLI)
	@rm -rf $(TOOLS_HOST_DIR)/tmp-crossplane-cli
	@$(OK) installing Crossplane CLI $(CROSSPLANE_CLI_VERSION)

$(CROSSPLANE_CHART):
	@$(INFO) downloading Crossplane chart $(CROSSPLANE_VERSION)
	@rm -rf $(CROSSPLANE_CHART_DIR) $(TOOLS_HOST_DIR)/tmp-crossplane-chart
	@mkdir -p $(CROSSPLANE_CHART_DIR) $(TOOLS_HOST_DIR)/tmp-crossplane-chart
	@curl -fsSL https://github.com/crossplane/crossplane/archive/refs/tags/v$(CROSSPLANE_VERSION).tar.gz | tar -xz -C $(TOOLS_HOST_DIR)/tmp-crossplane-chart
	@cp -R $(TOOLS_HOST_DIR)/tmp-crossplane-chart/*/cluster/charts/crossplane/. $(CROSSPLANE_CHART_DIR)/
	@rm -rf $(TOOLS_HOST_DIR)/tmp-crossplane-chart
	@$(OK) downloading Crossplane chart $(CROSSPLANE_VERSION)

controlplane.up: $(HELM) $(KUBECTL) $(KIND) $(CROSSPLANE_CHART)
	@$(INFO) setting up controlplane
	@$(KIND) get kubeconfig --name $(KIND_CLUSTER_NAME) >/dev/null 2>&1 || $(KIND) create cluster --name=$(KIND_CLUSTER_NAME)
	@$(INFO) "setting kubectl context to kind-$(KIND_CLUSTER_NAME)"
	@$(KUBECTL) config use-context "kind-$(KIND_CLUSTER_NAME)"
	@if ! $(HELM) get notes -n $(CROSSPLANE_NAMESPACE) crossplane >/dev/null 2>&1; then \
		if [ -z "$(CROSSPLANE_ARGS)" ]; then \
			$(HELM) install crossplane --create-namespace --namespace=$(CROSSPLANE_NAMESPACE) --set image.tag=v$(CROSSPLANE_VERSION) $(CROSSPLANE_CHART_DIR); \
		else \
			$(HELM) install crossplane --create-namespace --namespace=$(CROSSPLANE_NAMESPACE) --set image.tag=v$(CROSSPLANE_VERSION) --set "args={$(CROSSPLANE_ARGS)}" $(CROSSPLANE_CHART_DIR); \
		fi; \
	fi

# Define a variable for the optional flags
RENDER_ONLY_FLAG :=

# Check if the render-only flag is set
ifeq ($(RENDER_ONLY), true)
    RENDER_ONLY_FLAG := --render-only
endif

UPTEST_EXAMPLE_LIST := $(shell grep -v '^\#' cluster/test/cases.txt | paste -sd ',' -)

KEYCLOAK_VERSION ?=
MIN_KC_VERSION_26_4 := 26.4
MIN_KC_VERSION_26_5 := 26.5
MIN_KC_VERSION_ORGS := 26.6

ifneq ($(KEYCLOAK_VERSION),)
KC_VERSION_GE_26_4 := $(shell printf '%s\n%s' "$(MIN_KC_VERSION_26_4)" "$(KEYCLOAK_VERSION)" | sort -V | head -n1 | grep -q "$(MIN_KC_VERSION_26_4)" && echo true || echo false)
ifeq ($(KC_VERSION_GE_26_4),true)
UPTEST_EXAMPLE_LIST := $(shell grep -v '^\#' cluster/test/cases-kc-26.4.txt | paste -sd ',' -),$(UPTEST_EXAMPLE_LIST)
endif
KC_VERSION_GE_26_5 := $(shell printf '%s\n%s' "$(MIN_KC_VERSION_26_5)" "$(KEYCLOAK_VERSION)" | sort -V | head -n1 | grep -q "$(MIN_KC_VERSION_26_5)" && echo true || echo false)
ifeq ($(KC_VERSION_GE_26_5),true)
UPTEST_EXAMPLE_LIST := $(shell grep -v '^\#' cluster/test/cases-kc-26.5.txt | paste -sd ',' -),$(UPTEST_EXAMPLE_LIST)
endif
KC_VERSION_GE_ORGS := $(shell printf '%s\n%s' "$(MIN_KC_VERSION_ORGS)" "$(KEYCLOAK_VERSION)" | sort -V | head -n1 | grep -q "$(MIN_KC_VERSION_ORGS)" && echo true || echo false)
ifeq ($(KC_VERSION_GE_ORGS),true)
UPTEST_EXAMPLE_LIST := $(UPTEST_EXAMPLE_LIST),$(shell grep -v '^\#' cluster/test/cases-orgs.txt | paste -sd ',' -)
endif
endif

# This target requires the following environment variables to be set:
# - UPTEST_EXAMPLE_LIST, a comma-separated list of examples to test
#   To ensure the proper functioning of the end-to-end test resource pre-deletion hook, it is crucial to arrange your resources appropriately.
#   You can check the basic implementation here: https://github.com/crossplane/uptest/blob/main/internal/templates/03-delete.yaml.tmpl.
# - UPTEST_CLOUD_CREDENTIALS (optional), multiple sets of AWS IAM User credentials specified as key=value pairs.
#   The support keys are currently `DEFAULT` and `PEER`. So, an example for the value of this env. variable is:
#   DEFAULT='[default]
#   aws_access_key_id = REDACTED
#   aws_secret_access_key = REDACTED'
#   PEER='[default]
#   aws_access_key_id = REDACTED
#   aws_secret_access_key = REDACTED'
#   The associated `ProviderConfig`s will be named as `default` and `peer`.
# - UPTEST_DATASOURCE_PATH (optional), please see https://github.com/crossplane/uptest#injecting-dynamic-values-and-datasource
# - Render and inspect the generated chainsaws test cases by using uptest --render-only flag and checking the output directory.
uptest: $(UPTEST) $(KUBECTL) $(CHAINSAW) $(CROSSPLANE_CLI)
	@$(INFO) running automated tests
	@KUBECTL=$(KUBECTL) CHAINSAW=$(CHAINSAW) CROSSPLANE_CLI=$(CROSSPLANE_CLI) CROSSPLANE_NAMESPACE=$(CROSSPLANE_NAMESPACE) $(UPTEST) e2e "$(UPTEST_EXAMPLE_LIST)" $(RENDER_ONLY_FLAG) --data-source="${UPTEST_DATASOURCE_PATH}" --setup-script=cluster/test/setup.sh --default-conditions="Test" --default-timeout=2400s || $(FAIL)
	@$(OK) running automated tests

local-deploy: build controlplane.up local.xpkg.deploy.provider.$(PROJECT_NAME)
	@$(INFO) running locally built provider
	@$(KUBECTL) wait crd providers.pkg.crossplane.io --for=create --timeout 5m
	@$(KUBECTL) wait provider.pkg $(PROJECT_NAME) --for condition=Healthy --for condition=Installed --for=create --timeout 5m
	@$(OK) running locally built provider

local-deploy-provider: build local.xpkg.deploy.provider.$(PROJECT_NAME)
	@$(INFO) running locally built provider
	@$(KUBECTL) wait crd providers.pkg.crossplane.io --for=create --timeout 5m
	@$(KUBECTL) wait provider.pkg $(PROJECT_NAME) --for condition=Healthy --for condition=Installed --for=create --timeout 5m
	@$(OK) running locally built provider

local-deploy-provider-prebuilt: local.xpkg.deploy.provider.$(PROJECT_NAME)
	@$(INFO) running pre-built provider
	@$(KUBECTL) wait crd providers.pkg.crossplane.io --for=create --timeout 5m
	@$(KUBECTL) wait provider.pkg $(PROJECT_NAME) --for condition=Healthy --for condition=Installed --for=create --timeout 5m
	@$(OK) running pre-built provider

e2e: local-deploy uptest

crddiff: $(UPTEST)
	@$(INFO) Checking breaking CRD schema changes
	@for crd in $${MODIFIED_CRD_LIST}; do \
		if ! git cat-file -e "$${GITHUB_BASE_REF}:$${crd}" 2>/dev/null; then \
			echo "CRD $${crd} does not exist in the $${GITHUB_BASE_REF} branch. Skipping..." ; \
			continue ; \
		fi ; \
		echo "Checking $${crd} for breaking API changes..." ; \
		changes_detected=$$($(UPTEST) crddiff revision <(git cat-file -p "$${GITHUB_BASE_REF}:$${crd}") "$${crd}" 2>&1) ; \
		if [[ $$? != 0 ]] ; then \
			printf "\033[31m"; echo "Breaking change detected!"; printf "\033[0m" ; \
			echo "$${changes_detected}" ; \
			echo ; \
		fi ; \
	done
	@$(OK) Checking breaking CRD schema changes

schema-version-diff:
	@$(INFO) Checking for native state schema version changes
	@export PREV_PROVIDER_VERSION=$$(git cat-file -p "${GITHUB_BASE_REF}:Makefile" | sed -nr 's/^export[[:space:]]*TERRAFORM_PROVIDER_VERSION[[:space:]]*:=[[:space:]]*(.+)/\1/p'); \
	echo Detected previous Terraform provider version: $${PREV_PROVIDER_VERSION}; \
	echo Current Terraform provider version: $${TERRAFORM_PROVIDER_VERSION}; \
	mkdir -p $(WORK_DIR); \
	git cat-file -p "$${GITHUB_BASE_REF}:config/schema.json" > "$(WORK_DIR)/schema.json.$${PREV_PROVIDER_VERSION}"; \
	./scripts/version_diff.py config/generated.lst "$(WORK_DIR)/schema.json.$${PREV_PROVIDER_VERSION}" config/schema.json
	@$(OK) Checking for native state schema version changes

# Compare the current schema.json against a schema from a specific provider version.
# Downloads the old provider binary, generates its schema, and diffs the two.
# Usage:
#   make schema-diff OLD_PROVIDER_VERSION=5.6.0
schema-diff: $(TERRAFORM)
	@if [ -z "$(OLD_PROVIDER_VERSION)" ]; then \
		echo "Error: OLD_PROVIDER_VERSION is required. Usage: make schema-diff OLD_PROVIDER_VERSION=5.6.0"; \
		exit 1; \
	fi
	@$(INFO) Comparing provider schema $(OLD_PROVIDER_VERSION) vs $(TERRAFORM_PROVIDER_VERSION)
	@DIFF_OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	DIFF_ARCH=$$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/'); \
	DIFF_PLATFORM="$${DIFF_OS}_$${DIFF_ARCH}"; \
	mkdir -p $(WORK_DIR)/schema-diff/old/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(OLD_PROVIDER_VERSION)/$${DIFF_PLATFORM}; \
	curl -fsSL $(TERRAFORM_PROVIDER_REPO)/releases/download/v$(OLD_PROVIDER_VERSION)/$(TERRAFORM_PROVIDER_DOWNLOAD_NAME)_$(OLD_PROVIDER_VERSION)_$${DIFF_PLATFORM}.zip \
		-o $(WORK_DIR)/schema-diff/old/terraform-provider.zip; \
	unzip -o -qq $(WORK_DIR)/schema-diff/old/terraform-provider.zip \
		-d $(WORK_DIR)/schema-diff/old/$(TERRAFORM_FILE_MIRROR_REPO)/$(TERRAFORM_PROVIDER_SOURCE)/$(OLD_PROVIDER_VERSION)/$${DIFF_PLATFORM}/; \
	rm -f $(WORK_DIR)/schema-diff/old/terraform-provider.zip; \
	echo '{"terraform":[{"required_providers":[{"provider":{"source":"'"$(TERRAFORM_PROVIDER_SOURCE)"'","version":"'"$(OLD_PROVIDER_VERSION)"'"}}],"required_version":"'"$(TERRAFORM_VERSION)"'"}]}' > $(WORK_DIR)/schema-diff/old/main.tf.json; \
	echo 'provider_installation { filesystem_mirror { path = "$(WORK_DIR)/schema-diff/old/$(TERRAFORM_FILE_MIRROR)" include = ["*/*/*"] } }' > $(WORK_DIR)/schema-diff/old/config.tfrc; \
	TF_CLI_CONFIG_FILE=$(WORK_DIR)/schema-diff/old/config.tfrc $(TERRAFORM) -chdir=$(WORK_DIR)/schema-diff/old init -no-color > $(WORK_DIR)/schema-diff/old/terraform-logs.txt 2>&1; \
	TF_CLI_CONFIG_FILE=$(WORK_DIR)/schema-diff/old/config.tfrc $(TERRAFORM) -chdir=$(WORK_DIR)/schema-diff/old providers schema -json=true > $(WORK_DIR)/schema-diff/old-schema.json 2>> $(WORK_DIR)/schema-diff/old/terraform-logs.txt; \
	echo ""; \
	echo "Comparing schema v$(OLD_PROVIDER_VERSION) -> v$(TERRAFORM_PROVIDER_VERSION):"; \
	echo ""; \
	./scripts/version_diff.py config/generated.lst $(WORK_DIR)/schema-diff/old-schema.json config/schema.json || true
	@$(OK) Comparing provider schema $(OLD_PROVIDER_VERSION) vs $(TERRAFORM_PROVIDER_VERSION)

# test.race runs the concurrency-safety unit tests with the Go race detector.
# It is separate from `make test` because the full suite starts kube-apiservers
# that do not need the race detector.
test.race:
	@$(INFO) running race detector tests
	@CGO_ENABLED=1 go test -race -count=1 ./internal/tfconcurrency/... ./internal/clients/... ./config/...
	@$(OK) running race detector tests

.PHONY: cobertura submodules fallthrough run crds.clean schema-diff schema-version-diff test.race

# ====================================================================================
# Special Targets

define CROSSPLANE_MAKE_HELP
Crossplane Targets:
    cobertura             Generate a coverage report for cobertura applying exclusions on generated files.
    submodules            Update the submodules, such as the common build scripts.
    run                   Run crossplane locally, out-of-cluster. Useful for development.
    schema-diff           Compare provider schema between versions. Usage: make schema-diff OLD_PROVIDER_VERSION=5.6.0
    schema-version-diff   Check for schema version changes against the base branch (CI).
    test.race             Run the concurrency-safety unit tests with the Go race detector.

endef
# The reason CROSSPLANE_MAKE_HELP is used instead of CROSSPLANE_HELP is because the crossplane
# binary will try to use CROSSPLANE_HELP if it is set, and this is for something different.
export CROSSPLANE_MAKE_HELP

crossplane.help:
	@echo "$$CROSSPLANE_MAKE_HELP"

help-special: crossplane.help

.PHONY: crossplane.help help-special

# TODO(negz): Update CI to use these targets.
vendor: modules.download
vendor.check: modules.check

# ====================================================================================
# Documentation targets

# Regenerate docs/static/llms.txt and docs/static/llms-full.txt from the
# Hugo content tree.
docs-gen:
	@echo "==> generating llms.txt and llms-full.txt"
	@bash docs/scripts/gen-llms.sh
	@echo "==> done"

# Verify that docs/static/llms.txt and llms-full.txt are up to date with the
# Hugo content tree. Exits non-zero when the files are stale. Run docs-gen to
# fix.
docs-freshness-check:
	@echo "==> checking llms.txt freshness"
	@bash docs/scripts/gen-llms.sh --check

.PHONY: docs-gen docs-freshness-check
