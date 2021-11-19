locals {
  // ensure the prefix contains "mantil" since these resources are never user-created
  prefix = length(regexall(".*mantil.*", var.prefix)) > 0 ? var.prefix : "mantil-${var.prefix}"
  ws_handler = {
    name   = "${local.prefix}-ws-handler-${var.suffix}"
    s3_key = "${var.functions_s3_path}/ws-handler.zip"
  }
  ws_forwarder = {
    name   = "${local.prefix}-ws-forwarder-${var.suffix}"
    s3_key = "${var.functions_s3_path}/ws-forwarder.zip"
  }
  dynamodb_table = "${local.prefix}-ws-conns-${var.suffix}"
  ws_env = merge(var.ws_env, {
    "MANTIL_KV_TABLE" = local.dynamodb_table
  })
}
