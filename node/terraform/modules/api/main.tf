locals {
  integrations = { for value in var.integrations : "${value.route}/${value.method}" => merge(
    value,
    {
      enable_auth : var.authorizer != null && coalesce(value.enable_auth, false)
    }
  ) }
}

terraform {
  experiments = [module_variable_optional_attrs]
}

resource "aws_api_gateway_account" "api" {
  cloudwatch_role_arn = aws_iam_role.cloudwatch.arn
}
