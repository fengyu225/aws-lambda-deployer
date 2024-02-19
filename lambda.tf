resource "aws_lambda_function" "api_handler" {
  function_name = "api_handler_function"
  package_type  = "Image"
  image_uri     = "072422391281.dkr.ecr.us-east-1.amazonaws.com/lambda:1"
  role          = aws_iam_role.lambda.arn
  publish       = true
  lifecycle {
    ignore_changes = [
      image_uri,
    ]
  }
}

resource "aws_lambda_function" "lambda_deployer" {
  function_name = "lambda_deployer_function"
  package_type  = "Image"
  image_uri     = "072422391281.dkr.ecr.us-east-1.amazonaws.com/lambda-deployer:latest"
  role          = aws_iam_role.lambda-deployer.arn
  publish       = true
  timeout       = 900
  lifecycle {
    ignore_changes = [
      image_uri,
    ]
  }
}

resource "aws_lambda_function" "lambda_authorizer" {
  function_name = "lambda-authorizer"
  role          = aws_iam_role.lambda_authorizer.arn
  image_uri     = "072422391281.dkr.ecr.us-east-1.amazonaws.com/lambda-authorizer:4"
  package_type  = "Image"
}

resource "aws_lambda_alias" "production_alias" {
  name             = "production"
  description      = "Production alias"
  function_name    = aws_lambda_function.api_handler.function_name
  function_version = aws_lambda_function.api_handler.version
  lifecycle {
    ignore_changes = [
      function_version,
    ]
  }
}