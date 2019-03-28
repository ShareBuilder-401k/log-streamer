provider "aws" {
  region = "${var.region}"
}

resource "aws_lambda_function" "log_streamer" {
  filename         = "./lambda.zip"
  function_name    = "${var.function_name}-${var.env}-${var.region}"
  description      = "lambda to receive logs from cloudwatch and send them to elasticsearch through a signed request"
  role             = "${data.aws_iam_role.role.arn}"
  handler          = "${var.lambda_handler}"
  runtime          = "go1.x"
  timeout          = 300
  source_code_hash = "${base64sha256(file("./lambda.zip"))}"
  publish          = true

  dead_letter_config {
    target_arn = "${aws_sqs_queue.log_streamer_dlq.arn}"
  }

  environment {
    variables = {
      ENV             = "${var.env}"
      ES_HOST         = "${var.es_host}"
      ES_INDEX_PREFIX = "${var.es_index_prefix}"
    }
  }
}

resource "aws_sqs_queue" "log_streamer_dlq" {
  name                      = "${var.function_name}-${var.env}-${var.region}-deadletter-queue"
  kms_master_key_id         = "/aws/sqs"
  message_retention_seconds = 1209600
}

resource "aws_cloudwatch_metric_alarm" "marketing_engagement_dlq_alarm" {
  alarm_name          = "${var.function_name}-${var.env}-${var.region}-dlq-messages"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "1"
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = "900"
  statistic           = "Sum"
  threshold           = "1"
  alarm_description   = "This metric monitors lambda messages enqueued into the dead-letter queue"
  alarm_actions       = ["${data.aws_sns_topic.deadletter_topic.arn}"]

  dimensions {
    QueueName = "${aws_sqs_queue.log_streamer_dlq.name}"
  }
}

data "aws_sns_topic" "deadletter_topic" {
  name = "alerts-${var.env}-${var.region}"
}

data "aws_iam_role" "role" {
  name = "${var.iam_role}"
}

output "lambda_version" {
  value = "${aws_lambda_function.log_streamer.version}"
}
