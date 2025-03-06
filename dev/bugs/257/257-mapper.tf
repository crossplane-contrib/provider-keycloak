terraform {
  required_providers {
    keycloak = {
      source = "keycloak/keycloak"
      version = "5.1.1"
    }
  }
}

provider "keycloak" {
    url = "http://172.18.0.31"
    base_path = "/auth"
    client_id     = "admin-cli"
    username = "admin"
    password = "admin"
}

resource "keycloak_realm" "realm" {
  realm   = "my-realm"
  enabled = true
}

resource "keycloak_oidc_identity_provider" "oidc" {
  realm             = keycloak_realm.realm.id
  alias             = "oidc"
  authorization_url = "https://example.com/auth"
  token_url         = "https://example.com/token"
  client_id         = "example_id"
  client_secret     = "example_token"
  default_scopes    = "openid random profile"
}

resource "keycloak_custom_identity_provider_mapper" "oidc" {
  realm                    = keycloak_realm.realm.id
  name                     = "email-attribute-importer"
  identity_provider_alias  = keycloak_oidc_identity_provider.oidc.alias
  identity_provider_mapper = "oidc-user-attribute-idp-mapper"

  # extra_config with syncMode is required in Keycloak 10+
  extra_config = {
    syncMode      = "INHERIT"
    Claim         = "my-email-claim"
    UserAttribute = "email"
  }
}