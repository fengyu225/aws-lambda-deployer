resource "aws_iam_role" "api_gateway" {
  name               = "api-gateway-role"
  assume_role_policy = data.aws_iam_policy_document.apigateway_assume_role.json
}

data "aws_iam_policy_document" "apigateway_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["apigateway.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "apigw_cloudwatch_logs_permissions" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
      "logs:PutLogEvents",
      "logs:GetLogEvents",
      "logs:FilterLogEvents",
    ]

    resources = ["*"]
  }
}

data "aws_iam_policy_document" "invoke_lambda_permissions" {
  statement {
    effect = "Allow"

    actions = [
      "lambda:InvokeFunction"
    ]

    resources = [
      "*"
    ]
  }
}

resource "aws_iam_role_policy" "allow_apigw_cloudwatch" {
  name   = "allow-api-gateway-cloudwatch"
  role   = aws_iam_role.api_gateway.id
  policy = data.aws_iam_policy_document.apigw_cloudwatch_logs_permissions.json
}

resource "aws_iam_role_policy" "allow_apigw_invoke_lambda" {
  name   = "allow-api-gateway-invoke-lambda"
  role   = aws_iam_role.api_gateway.id
  policy = data.aws_iam_policy_document.invoke_lambda_permissions.json
}

resource "aws_iam_role" "lambda" {
  name               = "lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "lambda_authorizer" {
  name               = "lambda_authorizer_role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_iam_role_policy" "authorizer_lambda_policy" {
  name = "lambda_authorizer_policy"
  role = aws_iam_role.lambda_authorizer.id

  policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Resource = "arn:aws:logs:*:*:*"
        Effect   = "Allow"
      },
    ]
  })
}

resource "aws_iam_role" "lambda-deployer" {
  name               = "lambda-deployer"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_iam_role_policy_attachment" "lambda-deployer-sqs" {
  role       = aws_iam_role.lambda-deployer.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole"
}

data "aws_iam_policy_document" "lambda-deployer" {
  statement {
    sid     = "AllowLambdaDeployer"
    effect  = "Allow"
    actions = [
      "lambda:*",
      "cloudwatch:*",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "allow_lambda_deployer" {
  name   = "allow-lambda-deployer-consume-sqs"
  role   = aws_iam_role.lambda-deployer.id
  policy = data.aws_iam_policy_document.lambda-deployer.json
}
