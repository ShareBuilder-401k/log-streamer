package lambda

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

var region = "eu-west-3"
var env = "preprod"
var esHost = "http://www.com"
var esIndexPrefix = "test"

func TestBulkUpdateES_ControlMessage(t *testing.T) {
	log, _ := NewAwsLog(&NewAwsLogInput{
		env:           &env,
		region:        &region,
		esHost:        &esHost,
		esIndexPrefix: &esIndexPrefix,
		zippedData:    &zippedData,
	})
	log.ESBulkPayload = ""
	s, _ := log.BulkUpdateES()

	if s != "handled control message" {
		t.Errorf("expected \"handled control message\", got %s", s)
	}
}

func TestSignedBulkUpdateRequestPayload(t *testing.T) {
	esIndexPrefix := "non-prod"
	env := ""
	log, _ := NewAwsLog(&NewAwsLogInput{
		env:           &env,
		region:        &region,
		esHost:        &esHost,
		esIndexPrefix: &esIndexPrefix,
		zippedData:    &zippedData,
	})
	r := log.SignedBulkUpdateRequest()
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	body := string(bodyBytes)

	index := fmt.Sprintf("%s-%s", "non-prod", time.Now().Format("2006.01.02"))
	m := `{"@id":"123","@timestamp":"1969-12-31T16:00:10-08:00","env":"","log_group":"/test","log_stream":"test","message":"hello","region":"eu-west-3"}`
	expected := strings.Replace(m, "index_name", index, -1)

	if !strings.Contains(body, expected) {
		t.Errorf("unexpected bulk request payload. expected line %s to appear\n, got %s", expected, body)
	}
}

func TestFlattenMessage_InvalidJSON(t *testing.T) {
	message := "hello"
	source := map[string]string{
		"message": "hello",
	}
	FlattenMessage(&source, message)
	expected := map[string]string{
		"message": "hello",
	}
	for key := range expected {
		if expected[key] != source[key] {
			t.Errorf("unexpected flattened message. expected %v to appear\n, got %v", expected, source)
		}

	}
}

func TestFlattenMessage(t *testing.T) {
	message := `{"Properties":{"Protocol":"HTTP/1.1","EventId":{"Id":1}},"Renderings":{"HostingRequestStartingLog":[{"Format":"l","path":"/ping"},{"Format":"l","path":"/healthcheck"}]}}`
	source := map[string]string{
		"message": "hello",
	}
	FlattenMessage(&source, message)
	expected := map[string]string{
		"json_log_data.Properties.Protocol":                           "HTTP/1.1",
		"json_log_data.Renderings.HostingRequestStartingLog.0.Format": "l",
		"json_log_data.Renderings.HostingRequestStartingLog.0.path":   "/ping",
		"json_log_data.Renderings.HostingRequestStartingLog.1.Format": "l",
		"json_log_data.Renderings.HostingRequestStartingLog.1.path":   "/healthcheck",
	}
	for key := range expected {
		if expected[key] != source[key] {
			t.Errorf("unexpected flattened message. expected %v to appear\n, got %v", expected, source)
		}
	}
	for key := range source {
		if expected[key] != source[key] {
			t.Errorf("unexpected flattened message. expected %v to appear\n, got %v", expected, source)
		}
	}
}
