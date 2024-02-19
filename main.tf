output "api_gateway_invoke_url" {
  value       = "https://${aws_api_gateway_rest_api.api.id}.execute-api.us-east-1.amazonaws.com/${aws_api_gateway_deployment.api_deployment.stage_name}/convoy"
  description = "The URL to invoke the /convoy endpoint of the API Gateway"
}