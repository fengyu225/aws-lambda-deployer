resource "aws_cloudwatch_log_group" "api_gateway_log_group" {
  name = "/aws/api-gateway/${aws_api_gateway_rest_api.api.name}-access-logs"
}

resource "aws_cloudwatch_log_resource_policy" "api_gateway_logging_policy" {
  policy_name = "ApiGatewayLoggingPolicy"
  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Service = "apigateway.amazonaws.com" }
        Action    = ["logs:CreateLogStream", "logs:PutLogEvents"]
        Resource  = aws_cloudwatch_log_group.api_gateway_log_group.arn
      },
    ]
  })
}