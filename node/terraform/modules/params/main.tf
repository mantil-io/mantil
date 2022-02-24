locals {
  params = { for param in var.params : param.name => merge(
    param,
    {
      secure : coalesce(param.secure, false),
    }) if param.value != ""
  }
}

terraform {
  experiments = [module_variable_optional_attrs]
}

resource "aws_ssm_parameter" "param" {
  for_each = local.params
  name     = "${var.path_prefix}/${each.value.name}"
  type     = each.value.secure ? "SecureString" : "String"
  value    = each.value.value
  lifecycle {
    ignore_changes = [value]
  }
}
