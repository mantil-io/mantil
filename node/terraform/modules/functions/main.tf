locals {
  # set defaults and prepare function attributes
  functions = { for k, f in var.functions : k =>
    {
      s3_key : try(f.s3_key, "")

      function_name : format(var.naming_template, k) // prefix functions name with project name
      runtime : try(f.runtime, "provided.al2")       // default runtime is go
      handler : try(f.handler, "bootstrap")          // default handler for go is 'bootstrap'
      memory_size : try(f.memory_size, 128)          // default memory size
      timeout : try(f.timeout, 900)                  // default timeout
      path : try(f.path, k)                          // default path is function's name
      architecture : try(f.architecture, "arm64")    // default architecture is arm64
      env : length(try(f.env, {})) == 0 ? null : try(f.env, {})
      layers : try(f.layers, [])
      policy : try(f.policy, jsonencode({
        Version = "2012-10-17"
        Statement = [
          {
            Effect   = "Allow"
            Action   = "*"
            Resource = "*"
          }
        ]
        })
      )
    }
  }
}

resource "aws_lambda_function" "functions" {
  for_each = local.functions

  role = aws_iam_role.lambda[each.key].arn

  s3_bucket = var.s3_bucket
  s3_key    = each.value.s3_key

  function_name = each.value.function_name
  memory_size   = each.value.memory_size
  timeout       = each.value.timeout
  handler       = each.value.handler
  runtime       = each.value.runtime
  architectures = [each.value.architecture]
  layers        = each.value.layers

  dynamic "environment" {
    for_each = each.value.env[*]
    content {
      variables = environment.value
    }
  }
}

resource "aws_cloudwatch_log_group" "functions_log_groups" {
  for_each          = local.functions
  name              = "/aws/lambda/${each.value.function_name}"
  retention_in_days = 14
}
