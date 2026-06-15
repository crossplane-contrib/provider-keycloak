//go:build generate
// +build generate

/*
Copyright 2021 Upbound Inc.
*/

// NOTE: See the below link for details on what is happening here.
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

// Remove existing CRDs
//go:generate rm -rf ../package/crds

// Remove generated files
//go:generate bash -c "find ../apis -iname 'zz_*' ! -iname 'zz_generated.managed*.go' -delete"
//go:generate bash -c "find ../apis -type d -empty -delete"
//go:generate bash -c "find ../internal/controller -iname 'zz_*' -delete"
//go:generate bash -c "find ../internal/controller -type d -empty -delete"
//go:generate rm -rf ../examples-generated
//go:generate bash -c "find ../cmd/provider -name 'zz_*' -type f -delete"
//go:generate bash -c "find ../cmd/provider -type d -maxdepth 1 -mindepth 1 -empty -delete"

// Generate documentation from Terraform docs.
//go:generate go run github.com/crossplane/upjet/v2/cmd/scraper -n ${TERRAFORM_PROVIDER_SOURCE} -r ../.work/${TERRAFORM_PROVIDER_SOURCE}/${TERRAFORM_DOCS_PATH} -o ../config/provider-metadata.yaml --prelude-xpath "//text()[contains(., \"page_title\")]"

// Run Upjet generator
//go:generate go run ../cmd/generator/main.go ..

// Generate deepcopy methodsets and CRD manifests
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile=../hack/boilerplate.go.txt paths=../apis/... crd:allowDangerousTypes=true,crdVersions=v1 output:artifacts:config=../package/crds


// Generate crossplane-runtime methodsets (resource.Claim, etc)
//go:generate go run -tags generate github.com/crossplane/crossplane-tools/cmd/angryjet generate-methodsets --header-file=../hack/boilerplate.go.txt ../apis/...

// Run upjet's transformer for the generated resolvers to get rid of the cross
// API-group imports and to prevent import cycles
//go:generate go run github.com/crossplane/upjet/v2/cmd/resolver -g keycloak.crossplane.io -a github.com/crossplane-contrib/provider-keycloak/internal/apis -s -p ../apis/cluster/...
//go:generate go run github.com/crossplane/upjet/v2/cmd/resolver -g keycloak.m.crossplane.io -a github.com/crossplane-contrib/provider-keycloak/internal/apis -s -p ../apis/namespaced/...


package generate

import (
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen" //nolint:typecheck

	_ "github.com/crossplane/crossplane-tools/cmd/angryjet" //nolint:typecheck

	_ "github.com/crossplane/upjet/v2/cmd/scraper"

	_ "github.com/crossplane/upjet/v2/cmd/resolver"
)
