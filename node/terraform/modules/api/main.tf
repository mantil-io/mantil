terraform {
  experiments = [module_variable_optional_attrs]
}

module "http_api" {
  source       = "../http-api"
  prefix       = var.prefix
  suffix       = var.suffix
  integrations = var.integrations
  authorizer = var.authorizer == null ? null : {
    authorization_header = var.authorizer.authorization_header
    arn                  = aws_lambda_function.authorizer[0].arn
    invoke_arn           = aws_lambda_function.authorizer[0].invoke_arn
  }
}

module "ws_api" {
  count             = var.ws_enabled ? 1 : 0
  source            = "../ws-api"
  prefix            = var.prefix
  suffix            = var.suffix
  functions_bucket  = var.functions_bucket
  functions_s3_path = var.functions_s3_path
  ws_env            = var.ws_env
  authorizer = var.authorizer == null ? null : {
    authorization_header = var.authorizer.authorization_header
    arn                  = aws_lambda_function.authorizer[0].arn
    invoke_arn           = aws_lambda_function.authorizer[0].invoke_arn
  }
}
