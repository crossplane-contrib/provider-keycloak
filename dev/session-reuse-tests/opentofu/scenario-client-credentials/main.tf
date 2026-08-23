# OpenTofu equivalent of ../../crossplane/scenario-client-credentials. See
# ../scenario-password-grant/main.tf for the general notes on how OpenTofu's
# per-apply provider process lifecycle differs from Crossplane's long-running,
# cached client.
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

variable "keycloak_client_id" {
  type    = string
  default = "session-test-client"
}

variable "keycloak_client_secret" {
  type      = string
  default   = "session-test-secret"
  sensitive = true
}

variable "keycloak_realm" {
  type    = string
  default = "master"
}

provider "keycloak" {
  client_id     = var.keycloak_client_id
  client_secret = var.keycloak_client_secret
  url           = var.keycloak_url
  realm         = var.keycloak_realm
}

resource "keycloak_realm" "session_test" {
  realm   = "session-test-cc-realm-tofu"
  enabled = true
}

resource "keycloak_group" "session_test" {
  realm_id = keycloak_realm.session_test.id
  name     = "session-test-cc-group"
}
