locals {
  integrations = { for value in var.integrations : "${value.route}" => merge(
    value,
    {
      enable_auth : var.authorizer != null && coalesce(value.enable_auth, false),
    }
  ) }
  default_integration = try(
    [for k, v in local.integrations : v if coalesce(v.is_default, false) == true][0],
    values(local.integrations)[0]
  )
}

terraform {
  experiments = [module_variable_optional_attrs]
}
