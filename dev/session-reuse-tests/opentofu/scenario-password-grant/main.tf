# OpenTofu equivalent of ../../crossplane/scenario-password-grant, using the
# same terraform-provider-keycloak dependency this provider vendors. Used to
# compare/reproduce the session-growth behavior from
# https://github.com/crossplane-contrib/provider-keycloak/issues/309 without
# Crossplane in the loop.
#
# NOTE: unlike the long-running Crossplane provider process (which caches one
# client per ProviderConfig, see ../../../../docs/content/docs/using/reference/provider-config.md),
# each `tofu apply` invocation starts a fresh provider plugin process and
# therefore performs its own fresh login. Use run-apply-loop.sh in this
# directory to simulate repeated reconciliation and observe how sessions
# accumulate per-apply regardless of grant type, then contrast that with the
# Crossplane scenario where the client (and its login) is cached across
# reconciles.
terraform {
  required_providers {
    keycloak = {
      source  = "keycloak/keycloak"
      version = ">= 5.0.0"
    }
  }
}

variable "keycloak_url" {
  type        = string
  description = "Base URL of the Keycloak server, e.g. http://127.0.0.1:8080"
}

variable "keycloak_username" {
  type    = string
  default = "admin"
}

variable "keycloak_password" {
  type      = string
  default   = "admin"
  sensitive = true
}

variable "keycloak_realm" {
  type    = string
  default = "master"
}

provider "keycloak" {
  client_id = "admin-cli"
  username  = var.keycloak_username
  password  = var.keycloak_password
  url       = var.keycloak_url
  realm     = var.keycloak_realm
}

resource "keycloak_realm" "session_test" {
  realm   = "session-test-realm-tofu"
  enabled = true
}

resource "keycloak_group" "session_test" {
  realm_id = keycloak_realm.session_test.id
  name     = "session-test-group"
}
