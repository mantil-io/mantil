output "functions" {
  value = [for f in module.functions.functions :
    {
      name : f.name
      method : local.functions[f.name].method
      arn : f.arn
      invoke_arn : f.invoke_arn
    }
  ]
}
