package common

import (
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	xpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
)

const (
	// SelfPackagePath is the golang path for this package.
	SelfPackagePath = "github.com/crossplane-contrib/provider-keycloak/config/common"

	// PathServiceAccountRoleIDExtractor is the golang path to ARNExtractor function
	// in this package.
	PathServiceAccountRoleIDExtractor = SelfPackagePath + ".ServiceAccountRoleIDExtractor()"
	// PathAuthenticationFlowAliasExtractor is the golang path to ARNExtractor function
	// in this package.
	PathAuthenticationFlowAliasExtractor = SelfPackagePath + ".AuthenticationFlowAliasExtractor()"
	// PathUUIDExtractor is the golang path to UUIDExtractor function
	PathUUIDExtractor = SelfPackagePath + ".UUIDExtractor()"
)

// ServiceAccountRoleIDExtractor returns a reference.ExtractValueFn that can be used to extract the ServiceAccountRoleID from a managed resource.
func ServiceAccountRoleIDExtractor() reference.ExtractValueFn {
	return func(mg xpresource.Managed) string {
		paved, err := fieldpath.PaveObject(mg)
		if err != nil {
			// todo(hasan): should we log this error?
			return ""
		}
		r, err := paved.GetString("status.atProvider.serviceAccountUserId")
		if err != nil {
			// todo(hasan): should we log this error?
			return ""
		}
		return r
	}
}

// AuthenticationFlowAliasExtractor extract Alias from AuthenticationFlow Ref
func AuthenticationFlowAliasExtractor() reference.ExtractValueFn {
	return func(mg xpresource.Managed) string {
		paved, err := fieldpath.PaveObject(mg)
		if err != nil {
			// todo(hasan): should we log this error?
			return ""
		}
		// Caution, this is case-sensitive
		r, err := paved.GetString("status.atProvider.alias")
		if err != nil {
			// todo(hasan): should we log this error?
			return ""
		}
		return r
	}
}

// UUIDExtractor returns a reference.ExtractValueFn that can be used to extract the UUID from a managed resource.
func UUIDExtractor() reference.ExtractValueFn {
	return func(mg xpresource.Managed) string {
		paved, err := fieldpath.PaveObject(mg)
		if err != nil {
			// todo(hasan): should we log this error?
			return ""
		}
		r, err := paved.GetString("status.atProvider.id")
		// split at / and return the last element of there are two parts
		// this is to handle the case where the id is a path realm/uuid
		if err != nil {
			// todo(hasan): should we log this error?
			return ""
		}
		split := strings.Split(r, "/")
		if len(split) == 2 {
			return split[1]
		}
		return r
	}
}
