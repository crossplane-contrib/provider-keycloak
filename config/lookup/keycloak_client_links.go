package lookup

import (
	"context"

	_ "unsafe"

	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

// This needs to be removed in the future. See comments on GetComponents method
//
//go:linkname keycloakClientGet github.com/keycloak/terraform-provider-keycloak/keycloak.(*KeycloakClient).get
func keycloakClientGet(*keycloak.KeycloakClient, context.Context, string, interface{}, map[string]string) error
