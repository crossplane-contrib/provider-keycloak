/*
 Copyright 2022 Upbound Inc.
*/

package features

import "github.com/crossplane/crossplane-runtime/v2/pkg/feature"

// Feature flags.
const (
	// EnableAlphaExternalSecretStores enables alpha support for
	// External Secret Stores. See the below design for more details.
	// https://github.com/crossplane/crossplane/blob/390ddd/design/design-doc-external-secret-stores.md
	EnableAlphaExternalSecretStores feature.Flag = "EnableAlphaExternalSecretStores"
	EnableBetaManagementPolicies    feature.Flag = "EnableBetaManagementPolicies"
)
