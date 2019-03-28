terraform {
  backend "s3" {
    key = "log-streamer-trigger.tfstate"
  }
}
