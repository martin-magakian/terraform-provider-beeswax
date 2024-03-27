data "beeswax_user" "example" {
  # Limitation: it's only possible to retrieve the user by ID for now. TODO: allow finding user by email
  id = 42
}
