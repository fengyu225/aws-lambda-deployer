resource "aws_sqs_queue" "lambda_deployer_queue" {
  name                       = "lambda-deployer"
  visibility_timeout_seconds = 3600
  message_retention_seconds  = 345600 # 4 days in seconds
  max_message_size           = 262144  # 256 KB in bytes
  policy                     = data.aws_iam_policy_document.sqs_policy.json
}

data "aws_iam_policy_document" "sqs_policy" {
  statement {
    sid    = "owner"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = [aws_iam_role.lambda-deployer.arn]
    }
    actions   = ["SQS:SendMessage", "SQS:ReceiveMessage", "SQS:DeleteMessage", "SQS:GetQueueAttributes"]
  }
}

resource "aws_lambda_event_source_mapping" "sqs_lambda_trigger" {
  event_source_arn = aws_sqs_queue.lambda_deployer_queue.arn
  function_name    = aws_lambda_function.lambda_deployer.arn
}