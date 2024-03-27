resource "beeswax_role" "example" {
  name           = "my_role"
  parent_role_id = 1
  permissions = [
    {
      object_type = "account"
      permission  = 13
    },
    {
      object_type = "creative"
      permission  = 13
    },
    {
      object_type = "static"
      permission  = 1
    },
  ]
}