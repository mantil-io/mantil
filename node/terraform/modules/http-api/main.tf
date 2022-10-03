locals {
  integrations = { for value in var.integrations : "${value.route}" => merge(
    value,
    {
      enable_auth : var.authorizer != null && coalesce(value.enable_auth, false),
    }
  ) }
}
