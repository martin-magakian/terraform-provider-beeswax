resource "beeswax_user" "example" {
  super_user         = false
  email              = "myemail@myorg.com"
  first_name         = "Martin"
  last_name          = "Magakian"
  role_id            = 1
  account_id         = 1
  active             = true
  all_account_access = true
  account_group_ids  = []
}