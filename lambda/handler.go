package lambda

import (
	"fmt"
	"os"
	"time"
)

// Event model for lambda input event
type Event struct {
	AwsLogs struct {
		Data string `json:"data"`
	} `json:"awslogs"`
}

// HandleRequest is main function for handling requests
// @params event: the lambda event received
// @returns: (status of request, error)
func HandleRequest(event Event) (string, error) {
	env := os.Getenv("ENV")
	region := os.Getenv("AWS_REGION")
	esHost := os.Getenv("ES_HOST")
	esIndexPrefix := os.Getenv("ES_INDEX_PREFIX")

	if event.AwsLogs.Data == "integration test" {
		return TestElasticsearchConnectivty(esHost, region)
	}

	logs, err := NewAwsLog(&NewAwsLogInput{
		env:           &env,
		region:        &region,
		esHost:        &esHost,
		esIndexPrefix: &esIndexPrefix,
		zippedData:    &event.AwsLogs.Data,
	})
	if err != nil {
		fmt.Printf("error processing event: %v", err)
		return "", err
	}
	fmt.Printf("handling logs for %s\n", logs.Data.LogGroup)
	status, err := logs.BulkUpdateES()

	// retry bulk update twice if it fails
	for i := 0; i < 2 && err != nil; i++ {
		fmt.Printf("%s - %v\n", status, err)
		fmt.Println("retrying bulk update after 2 seconds")
		time.Sleep(2 * time.Second)
		status, err = logs.BulkUpdateES()
	}

	if err != nil {
		fmt.Println("bulk update failed three times, sending to DLQ")
	}
	fmt.Println(status)
	return status, err
}
