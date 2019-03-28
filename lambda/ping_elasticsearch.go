package lambda

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

// TestElasticsearchConnectivty tests connectivity to ES
// @params: (esHost, region string)
// @returns: (status string, error)
func TestElasticsearchConnectivty(esHost, region string) (string, error) {
	signer := v4.NewSigner(credentials.NewEnvCredentials())
	request, _ := http.NewRequest("GET", esHost, nil)
	request.Header.Add("Content-Type", "application/json")
	signer.Sign(request, nil, "es", region, time.Now())

	err := pingElasticsearch(request)
	// retry ping twice if it failed
	for i := 0; i < 2 && err != nil; i++ {
		fmt.Printf("%v, retrying after 2 seconds\n", err)
		time.Sleep(2 * time.Second)
		err = pingElasticsearch(request)
	}
	if err == nil {
		return "success", nil
	}
	return "failed", err
}

func pingElasticsearch(request *http.Request) error {
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("ES connectivity error %v", err)
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d returned from ES", resp.StatusCode)
	}
	return err
}
