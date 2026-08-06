provider "keycloak" {
  client_id = "admin-cli"
  url       = "http://localhost:8080"
}

resource "keycloak_realm" "this" {
  realm        = "my-realm"
  enabled      = true
  display_name = "My Realm"

  smtp_server {
    host = "smtp.example.com"
    port = "587"
    from = "admin@example.com"
  }

  attributes = {
    custom = "value"
  }
}

resource "keycloak_group" "child_group" {
  realm_id  = keycloak_realm.this.id
  name      = "Child Group"
  parent_id = "some-static-parent"

  attributes = {
    department = "engineering"
  }
}

resource "keycloak_openid_client" "app" {
  realm_id  = keycloak_realm.this.id
  client_id = "my-app"
  name      = "My App"
  enabled   = true

  valid_redirect_uris = [
    "https://app.example.com/*",
  ]
}
