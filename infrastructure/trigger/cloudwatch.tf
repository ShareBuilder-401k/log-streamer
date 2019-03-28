provider "aws" {
  region = "${var.region}"
}

resource "aws_cloudwatch_log_subscription_filter" "log_filter" {
  count           = "${length(var.log_groups)}"
  name            = "${element(var.log_groups, count.index)}_subscription_filter"
  log_group_name  = "${element(var.log_groups, count.index)}"
  filter_pattern  = ""
  destination_arn = "${data.aws_lambda_function.active.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch_invocation" {
  statement_id  = "allow_cloudwatch_invocation"
  action        = "lambda:InvokeFunction"
  function_name = "${data.aws_lambda_function.active.arn}"
  principal     = "logs.${var.region}.amazonaws.com"
  source_arn    = "arn:aws:logs:${var.region}:${data.aws_caller_identity.current.account_id}:log-group*"
}

data "aws_lambda_function" "active" {
  function_name = "${var.function_name}-${var.env}-${var.region}"
  qualifier     = "ACTIVE"
}

data "aws_caller_identity" "current" {}
