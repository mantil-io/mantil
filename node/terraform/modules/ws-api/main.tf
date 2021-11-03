locals {
  ws_handler = {
    name   = "${var.prefix}-ws-handler-${var.suffix}"
    s3_key = "${var.functions_s3_path}/ws-handler.zip"
  }
  ws_forwarder = {
    name   = "${var.prefix}-ws-forwarder-${var.suffix}"
    s3_key = "${var.functions_s3_path}/ws-forwarder.zip"
  }
  dynamodb_table = "${var.prefix}-ws-connections-${var.suffix}"
  ws_env = merge(var.ws_env, {
    "MANTIL_KV_TABLE" = local.dynamodb_table
  })
}
