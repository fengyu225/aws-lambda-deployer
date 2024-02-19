resource "aws_api_gateway_rest_api" "api" {
  name        = "API"
  description = "API Gateway with Lambda Authorizer"
}

resource "aws_api_gateway_resource" "api_resource" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "convoy"
}

resource "aws_api_gateway_method" "api_method" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.api_resource.id
  http_method   = "GET"
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.authorizer.id
}

resource "aws_api_gateway_authorizer" "authorizer" {
  name                   = "api-authorizer"
  rest_api_id            = aws_api_gateway_rest_api.api.id
  type                   = "TOKEN"
  authorizer_credentials = aws_iam_role.api_gateway.arn
  authorizer_uri         = "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/${aws_lambda_function.lambda_authorizer.arn}/invocations"
}

resource "aws_api_gateway_integration" "lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.api_resource.id
  http_method             = aws_api_gateway_method.api_method.http_method
  // When using the Lambda Proxy integration (AWS_PROXY),
  // API Gateway forwards the incoming request to Lambda in a standard format,
  // including the original HTTP method, headers, path, and query parameters.
  // This forwarding is done using a POST request to the Lambda invocation API, which triggers the Lambda function.
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/${aws_lambda_function.api_handler.arn}:production/invocations"
}

resource "aws_api_gateway_deployment" "api_deployment" {
  depends_on  = [aws_api_gateway_integration.lambda_integration]
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "prod"
}

resource "aws_api_gateway_account" "api_gateway_account" {
  cloudwatch_role_arn = aws_iam_role.api_gateway.arn

  depends_on = [
    aws_iam_role_policy.allow_apigw_cloudwatch
  ]
}

resource "aws_api_gateway_method_settings" "settings" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "prod"
  method_path = "*/*"

  settings {
    metrics_enabled    = true
    logging_level      = "INFO"
    data_trace_enabled = true
  }

  depends_on = [
    aws_api_gateway_deployment.api_deployment
  ]
}

resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api_handler.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn = "${aws_api_gateway_rest_api.api.execution_arn}/*/GET/convoy"
}

resource "aws_lambda_permission" "api_gateway_invoke_api_handler_production" {
  statement_id  = "AllowExecutionFromAPIGatewayProductionVersion"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.api_handler.function_name}:production"
  principal     = "apigateway.amazonaws.com"
  source_arn = "${aws_api_gateway_rest_api.api.execution_arn}/*/GET/convoy"
}

resource "aws_lambda_permission" "api_gateway_invoke_authorizer" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda_authorizer.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "arn:aws:execute-api:us-east-1:072422391281:${aws_api_gateway_rest_api.api.id}/*/*/*"
}