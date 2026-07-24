resource "keycloak_group" "with_var" {
  realm_id  = var.realm_id
  name      = "Group ${var.suffix}"
  parent_id = local.parent
}

resource "keycloak_not_a_real_resource" "x" {
  foo = "bar"
}

data "keycloak_realm" "existing" {
  realm = "master"
}
