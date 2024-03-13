package common

import (
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	xpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
)

const (
	// SelfPackagePath is the golang path for this package.
	SelfPackagePath = "github.com/crossplane-contrib/provider-keycloak/config/common"

	// PathARNExtractor is the golang path to ARNExtractor function
	// in this package.
	PathServiceAccountRoleIDExtractor = SelfPackagePath + ".ServiceAccountRoleIDExtractor()"
)

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
