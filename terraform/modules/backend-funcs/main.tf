locals {
  name = "mantil"

  # set defaults and prepare function attributes
  functions = { for k, f in var.functions : k =>
    {
      s3_key : try(f.s3_key, "")
      function_name : "${local.name}-${k}"                                                                    // prefix functions name with project name
      runtime : try(f.runtime, "provided.al2")                                                                // default runtime is go
      handler : try(f.handler, "bootstrap")                                                                   // default handler for go is 'bootstrap'
      memory_size : try(f.memory_size, 128)                                                                   // default memory size
      timeout : try(f.timeout, 60)                                                                            // default timeout
      path : try(f.path, k)                                                                                   // default path is function's name
      env : length(merge(var.global_env, try(f.env, {}))) == 0 ? null : merge(var.global_env, try(f.env, {})) // merge global and function local env varialbes
      layers : try(f.layers, [])
    }
  }
}

resource "aws_lambda_function" "functions" {
  for_each = local.functions

  role = aws_iam_role.lambda.arn

  s3_bucket = var.s3_bucket
  s3_key    = each.value.s3_key

  function_name = each.value.function_name
  memory_size   = each.value.memory_size
  timeout       = each.value.timeout
  handler       = each.value.handler
  runtime       = each.value.runtime
  layers        = each.value.layers

  dynamic "environment" {
    for_each = each.value.env[*]
    content {
      variables = environment.value
    }
  }
}