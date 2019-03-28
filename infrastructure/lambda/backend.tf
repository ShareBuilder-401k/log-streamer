terraform {
  backend "s3" {
    key = "log-streamer-lambda.tfstate"
  }
}
