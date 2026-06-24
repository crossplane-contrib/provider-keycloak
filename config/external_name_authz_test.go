package config

import (
	"reflect"
	"testing"

	"github.com/crossplane-contrib/provider-keycloak/config/openidclient"
)

// TestAuthzExternalNameImportLookup locks in the external-name configuration
// for the client authorization resource and permission so they are resolved by
// their identifying properties (realm/resource-server/name) instead of relying
// on a provider-assigned identifier.
//
// This is the regression test for the 409 Conflict bug (issue #459): without a
// lookup the provider could not detect an already existing resource/permission
// in Keycloak and kept POSTing, getting a 409 Conflict back. If a future change
// reverts these entries to config.IdentifierFromProvider, the import/observe
// behaviour breaks again and this test fails.
func TestAuthzExternalNameImportLookup(t *testing.T) {
	cases := map[string]reflect.Value{
		"keycloak_openid_client_authorization_resource":   reflect.ValueOf(openidclient.AuthzResourceIdentifierFromIdentifyingProperties.GetIDFn),
		"keycloak_openid_client_authorization_permission": reflect.ValueOf(openidclient.AuthzPermissionIdentifierFromIdentifyingProperties.GetIDFn),
	}

	for tfName, want := range cases {
		t.Run(tfName, func(t *testing.T) {
			got, ok := ExternalNameConfigs[tfName]
			if !ok {
				t.Fatalf("%s: no external name config registered", tfName)
			}
			if got.GetIDFn == nil {
				t.Fatalf("%s: GetIDFn is nil; expected identifying-properties lookup", tfName)
			}
			if reflect.ValueOf(got.GetIDFn).Pointer() != want.Pointer() {
				t.Errorf("%s: external name config is not wired to the identifying-properties lookup; "+
					"the import/observe fix for issue #459 would be reverted", tfName)
			}
		})
	}
}
