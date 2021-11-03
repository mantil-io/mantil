output "http_url" {
  value = module.http_api.url
}

output "ws_url" {
  value = var.ws_enabled ? module.ws_api[0].url : ""
}
