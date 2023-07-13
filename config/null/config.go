/*
Copyright 2021 Upbound Inc.
*/

package null

import (
	ujconfig "github.com/upbound/upjet/pkg/config"
)

// Configure configures the null group
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("null_resource", func(r *ujconfig.Resource) {
		r.Kind = "Resource"
		// And other overrides.
	})
}
