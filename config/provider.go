/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/upbound/upjet/pkg/config"

	"github.com/corewire/config/group"
	"github.com/corewire/config/openidclient"
	"github.com/corewire/config/realm"
	"github.com/corewire/config/role"
	"github.com/corewire/config/mapper"
)

const (
	resourcePrefix = "keycloak"
	modulePath     = "github.com/corewire/provider-keycloak"
	rootGroup      = "keycloak.crossplane.io"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithDefaultResourceOptions(ExternalNameConfigurations()),
		ujconfig.WithRootGroup(rootGroup))

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
		realm.Configure,
		group.Configure,
		role.Configure,
		openidclient.Configure,
		mapper.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
