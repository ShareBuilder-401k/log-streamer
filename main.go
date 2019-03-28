package main

import (
	"github.com/Sharebuilder-401k/log-streamer/lambda"
	aws "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	aws.Start(lambda.HandleRequest)
}
