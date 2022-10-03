locals {
  domains = {
    http = trim("${var.custom_domain.http_subdomain}.${var.custom_domain.domain_name}", ".")
    ws   = trim("${var.custom_domain.ws_subdomain}.${var.custom_domain.domain_name}", ".")
  }
}

module "custom_domain" {
  for_each           = { for k, v in local.domains : k => v if v != "" }
  source             = "../custom-domain"
  domain_name        = each.value
  cert_domain        = var.custom_domain.cert_domain
  hosted_zone_domain = var.custom_domain.hosted_zone_domain
}

module "http_api" {
  depends_on      = [module.custom_domain]
  source          = "../http-api"
  naming_template = var.naming_template
  integrations    = var.integrations
  authorizer = var.authorizer == null ? null : {
    authorization_header = var.authorizer.authorization_header
    arn                  = aws_lambda_function.authorizer[0].arn
    invoke_arn           = aws_lambda_function.authorizer[0].invoke_arn
  }
  domain = local.domains.http
}

module "ws_api" {
  depends_on        = [module.custom_domain]
  count             = var.ws_enabled ? 1 : 0
  source            = "../ws-api"
  suffix            = var.suffix
  naming_template   = var.naming_template
  functions_bucket  = var.functions_bucket
  functions_s3_path = var.functions_s3_path
  ws_env            = var.ws_env
  authorizer = var.authorizer == null ? null : {
    authorization_header = var.authorizer.authorization_header
    arn                  = aws_lambda_function.authorizer[0].arn
    invoke_arn           = aws_lambda_function.authorizer[0].invoke_arn
  }
  domain = local.domains.ws
}
