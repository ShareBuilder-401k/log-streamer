package lambda

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var zippedData = "H4sICPFQnl0AA3RlbXAuanNvbgCr5lJQUMpNLS5OTE8NqSxIVbJSUMrJT1fSAYkDGe5F+aUFIEH9ktTiErhwcElRamIuSBxF2LUsNa+kGCgcDRRQUKgGk0CpkkygFSWJuSCTDA2AQAcmA7UaZFJGak5OvhJcJjMFJGhoZKwEFqkFkrFctQAe8j0HsAAAAA=="

func TestIntegration(t *testing.T) {
	e := Event{}
	e.AwsLogs.Data = "integration test"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if r.Header.Get("Authorization") == "" {
				t.Errorf("expected Authorization header to be set")
			}

			if r.Header.Get("Content-Type") == "" {
				t.Errorf("expected Content-Type header to be set")
			}

			if r.Header.Get("X-Amz-Date") == "" {
				t.Errorf("expected X-Amz-Date header to be set")
			}

			if r.Header.Get("X-Amz-Security-Token") == "" {
				t.Errorf("expected X-Amz-Security-Token header to be set")
			}
			w.WriteHeader(http.StatusOK)
		}
	}))

	os.Setenv("AWS_ACCESS_KEY_ID", "key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	os.Setenv("AWS_SESSION_TOKEN", "token")
	os.Setenv("ES_HOST", ts.URL)

	status, err := HandleRequest(e)
	if err != nil || status != "success" {
		t.Errorf("unexpected integration test error")
	}
}

func TestIntegration_Error(t *testing.T) {
	e := Event{}
	e.AwsLogs.Data = "integration test"

	os.Setenv("AWS_ACCESS_KEY_ID", "key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	os.Setenv("AWS_SESSION_TOKEN", "token")
	os.Setenv("ES_HOST", "http://phoneybaloney.co.uk")

	status, err := HandleRequest(e)
	if err == nil || status != "failed" {
		t.Errorf("expected integration test failure, instead got status: %s\n, error: %v", status, err)
	}
}

func TestHandleRequest_BadZippedData(t *testing.T) {
	e := Event{}
	e.AwsLogs.Data = "bad_data"
	_, err := HandleRequest(e)

	if err.Error() != "illegal base64 data at input byte 3" {
		t.Errorf("expected error expanding aws log data, got %v", err)
	}
}

func TestRequest_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if r.RequestURI != "/_bulk" {
				t.Errorf("incorrect update path created: expected /_bulk, got %s", r.RequestURI)
			}
			response := []interface{}{}
			r, _ := json.Marshal(response)
			w.WriteHeader(http.StatusBadGateway)
			w.Write(r)
		}

	}))
	os.Setenv("ES_HOST", ts.URL)
	e := Event{}
	e.AwsLogs.Data = zippedData
	status, _ := HandleRequest(e)

	if status != "elasticsearch update error" {
		t.Errorf("unexpected es error update status: %s", status)
	}
}

func TestRequest_EsError(t *testing.T) {
	os.Setenv("ES_HOST", "http://phoneybaloney.co.uk")
	e := Event{}
	e.AwsLogs.Data = zippedData
	status, _ := HandleRequest(e)

	if status != "elasticsearch update error" {
		t.Errorf("unexpected es error update status: %s", status)
	}
}

func TestRequest_BulkUpdate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if r.RequestURI != "/_bulk" {
				t.Errorf("incorrect update path created: expected /_bulk, got %s", r.RequestURI)
			}
			if r.Header.Get("Authorization") == "" {
				t.Errorf("expected Authorization header to be set")
			}

			if r.Header.Get("Content-Type") == "" {
				t.Errorf("expected Content-Type header to be set")
			}

			if r.Header.Get("X-Amz-Date") == "" {
				t.Errorf("expected X-Amz-Date header to be set")
			}

			if r.Header.Get("X-Amz-Security-Token") == "" {
				t.Errorf("expected X-Amz-Security-Token header to be set")
			}
			response := ESUpdateResults{
				Items: []ESUpdateResultItem{
					ESUpdateResultItem{Status: 200},
					ESUpdateResultItem{Status: 400},
					ESUpdateResultItem{Status: 200},
					ESUpdateResultItem{Status: 500},
				},
			}
			r, _ := json.Marshal(response)
			w.Write(r)
		}

	}))
	os.Setenv("ES_HOST", ts.URL)
	e := Event{}
	e.AwsLogs.Data = zippedData
	os.Setenv("AWS_ACCESS_KEY_ID", "key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	os.Setenv("AWS_SESSION_TOKEN", "token")
	status, _ := HandleRequest(e)

	if status != expectedESUpdateStatus {
		t.Errorf("unexpected es update status:%s", status)
	}
}

var expectedESUpdateStatus = fmt.Sprintf(`elasticsearch bulk update successful
succuessful items: 2
failed items: 2`)
