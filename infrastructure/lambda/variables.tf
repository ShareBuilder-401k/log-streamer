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

variable "lambda_handler" {
  default     = "main"
  description = "lambda handler name"
}

variable "iam_role" {
  description = "lambda execution role"
}

variable "es_host" {
  description = "elasticsearch endpoint url"
}

variable "es_index_prefix" {
  description = "elasticsearch index prefix"
}
