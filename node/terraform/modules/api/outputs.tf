output "http_url" {
  value = aws_apigatewayv2_api.http.api_endpoint
}

output "ws_url" {
  value = aws_apigatewayv2_api.ws.api_endpoint
}
