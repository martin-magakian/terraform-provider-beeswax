data "beeswax_roles" "all" {
}

locals {
  roles_by_name = {
    for role in data.beeswax_roles.all.roles :
    role.name => role
  }
}

resource "beeswax_user" "user" {
  super_user         = false
  email              = "martin@org.com"
  first_name         = "Martin"
  last_name          = "Magakian"
  role_id            = local.roles_by_name["Administrator"].id
  account_id         = 1
  active             = true
  all_account_access = false
  account_group_ids  = []
}
