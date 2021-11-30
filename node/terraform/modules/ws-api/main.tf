locals {
  // ensure the naming template contains "mantil" since these resources are never user-created
  naming_template = length(regexall(".*mantil.*", var.naming_template)) > 0 ? var.naming_template : "mantil-${var.naming_template}"
  ws_handler = {
    name   = format(local.naming_template, "ws-handler")
    s3_key = "${var.functions_s3_path}/ws-handler.zip"
  }
  ws_forwarder = {
    name   = format(local.naming_template, "ws-forwarder")
    s3_key = "${var.functions_s3_path}/ws-forwarder.zip"
  }
  dynamodb_table = format(local.naming_template, "ws-conns")
  ws_env = merge(var.ws_env, {
    "MANTIL_KV_TABLE" = local.dynamodb_table
  })
}
