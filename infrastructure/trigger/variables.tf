variable "region" {
  description = "aws deployment region"
}

variable "env" {
  description = "aws deployment env"
}

variable "function_name" {
  default     = "log-streamer"
  description = "lambda function name"
}

variable "log_groups" {
  type        = "list"
  description = "comma seperated list of cloudwatch log groups to send to elasticsearch"
}
