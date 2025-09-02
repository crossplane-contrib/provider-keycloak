/*
Copyright 2021 Upbound Inc.
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crossplane/upjet/v2/pkg/pipeline"

	"github.com/crossplane-contrib/provider-keycloak/config"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" {
		panic("root directory is required to be given as argument")
	}
	rootDir := os.Args[1]
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		panic(fmt.Sprintf("cannot calculate the absolute path with %s", rootDir))
	}
	provider, err := config.GetProvider(true)
	if err != nil {
		panic(fmt.Sprintf("cannot get cluster provider configuration: %s", err))
	}
	providerNs, err := config.GetProviderNamespaced(true)
	if err != nil {
		panic(fmt.Sprintf("cannot get namespaced provider configuration: %s", err))
	}
	pipeline.Run(provider, providerNs, absRootDir)
}
