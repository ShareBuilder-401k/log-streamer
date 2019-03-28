package lambda

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

var client = &http.Client{}

// AwsLog is a class to receive data from lambda and send it to ES
type AwsLog struct {
	CompressedInput string
	Data            AwsLogData
	ESBulkPayload   string
	env             string
	region          string
	esHost          string
	esIndexPrefix   string
}

// AwsLogData is the model cloudwatch sends over to lambda
type AwsLogData struct {
	MessageType string     `json:"messageType"`
	LogGroup    string     `json:"logGroup"`
	LogStream   string     `json:"logStream"`
	LogEvents   []LogEvent `json:"logEvents"`
}

// LogEvent is the model for log events
type LogEvent struct {
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	ID        string `json:"id"`
}

// ESUpdateResults is the model for es bulk update response
type ESUpdateResults struct {
	Items []ESUpdateResultItem `json:"items"`
}

// ESUpdateResultItem is the model for es bulk update response
type ESUpdateResultItem struct {
	Status int `json:"status"`
}

// NewAwsLogInput is the struct for creating a new AwsLog struct
type NewAwsLogInput struct {
	env           *string
	region        *string
	esHost        *string
	esIndexPrefix *string
	zippedData    *string
}

// NewAwsLog is the constructor for AwsLog
// @params: i *NewAwsLogInput - new aws log input struct
// @returns: (new *AwsLog struct, error)
func NewAwsLog(i *NewAwsLogInput) (*AwsLog, error) {
	d, err := expandAwsInput(i.zippedData)
	if err != nil {
		return nil, err
	}
	b := esBulkPayload(&d, *i.env, *i.region, *i.esIndexPrefix)
	return &AwsLog{
		CompressedInput: *i.zippedData,
		Data:            d,
		ESBulkPayload:   b,
		env:             *i.env,
		region:          *i.region,
		esHost:          *i.esHost,
		esIndexPrefix:   *i.esIndexPrefix,
	}, nil
}

// BulkUpdateES will send a bulk update payload to ES through a signed request
// @returns: (bulk update status, error)
func (l *AwsLog) BulkUpdateES() (string, error) {
	if l.ESBulkPayload == "" {
		return "handled control message", nil
	}

	r, err := client.Do(l.SignedBulkUpdateRequest())
	if err != nil {
		return "elasticsearch update error", err
	} else if r.StatusCode != 200 {
		rp, _ := ioutil.ReadAll(r.Body)
		return "elasticsearch update error", fmt.Errorf("elasticsearch update error: %s", string(rp))
	}
	rb, _ := ioutil.ReadAll(r.Body)
	resp := ESUpdateResults{}
	json.Unmarshal(rb, &resp)
	successfulItems, failedItems := 0, 0
	for _, i := range resp.Items {
		if i.Status >= 300 {
			failedItems++
		} else {
			successfulItems++
		}
	}

	return fmt.Sprintf("elasticsearch bulk update successful\nsuccuessful items: %d\nfailed items: %d", successfulItems, failedItems), nil
}

// SignedBulkUpdateRequest will create a signed bulk update request
// @returns: aws signed request based on payload
func (l *AwsLog) SignedBulkUpdateRequest() *http.Request {
	signer := v4.NewSigner(credentials.NewEnvCredentials())
	request, _ := http.NewRequest("POST", l.esHost+"/_bulk", strings.NewReader(l.ESBulkPayload))
	body := bytes.NewReader([]byte(l.ESBulkPayload))
	request.Header.Add("Content-Type", "application/json")
	signer.Sign(request, body, "es", l.region, time.Now())
	return request
}

// expandAwsInput decodes and unzips data and unmarshals into AwsLogData struct
// @params: s *string - a base64 encoded string
// @returns: (AwsLogData struct, error)
func expandAwsInput(s *string) (AwsLogData, error) {
	logs := AwsLogData{}
	data, err := base64.StdEncoding.DecodeString(*s)
	if err != nil {
		fmt.Println("error decoding input")
		return logs, err
	}
	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		fmt.Println("error unzipping input: ", err)
		return logs, err
	}
	b, formatErr := ioutil.ReadAll(r)
	if formatErr != nil {
		return logs, formatErr
	}
	json.Unmarshal(b, &logs)
	return logs, nil
}

func todaysIndex(prefix string) string {
	current := time.Now()
	return fmt.Sprintf("%s-%s", prefix, current.Format("2006.01.02"))
}

func esBulkPayload(awsLogData *AwsLogData, env, region, prefix string) string {
	if awsLogData.MessageType == "CONTROL_MESSAGE" {
		return ""
	}
	bulkPayload := ""
	for _, event := range awsLogData.LogEvents {
		s := map[string]string{
			"@id":        event.ID,
			"@timestamp": time.Unix(0, event.Timestamp*int64(time.Millisecond)).Format(time.RFC3339Nano),
			"message":    event.Message,
			"env":        env,
			"region":     region,
			"log_group":  awsLogData.LogGroup,
			"log_stream": awsLogData.LogStream,
		}
		FlattenMessage(&s, event.Message)
		source, _ := json.Marshal(s)
		a := map[string]interface{}{
			"index": map[string]string{
				"_index": todaysIndex(prefix),
				"_type":  "_doc",
				"_id":    event.ID,
			},
		}
		action, _ := json.Marshal(a)
		bulkPayload = fmt.Sprintf("%s\n%s\n%s\n", bulkPayload, string(action), string(source))
	}
	return bulkPayload
}

// FlattenMessage will JSON parse message and append nested fields as top level keys in source
func FlattenMessage(source *map[string]string, message string) {
	in, out := map[string]interface{}{}, map[string]interface{}{}

	err := json.Unmarshal([]byte(message), &in)
	if err != nil {
		return
	}
	out["JSONFlag"] = false
	flattenJSON(&in, &out, "json_log_data")
	// remove message if it has already been parsed into json_log_data
	if b, ok := out["JSONFlag"].(bool); ok && b {
		delete((*source), "message")
	}
	delete(out, "JSONFlag")
	for key := range out {
		if s, ok := out[key].(string); ok {
			(*source)[key] = s
		}
	}
}

// flattenJSON will go through the in object and put all nested fields on the top level of out
// handles nested objects through recursion
func flattenJSON(in *map[string]interface{}, out *map[string]interface{}, prefix string) {
	for key := range *in {
		prefixedKey := fmt.Sprintf("%s.%s", prefix, key)

		if s, ok := (*in)[key].(string); ok {
			(*out)["JSONFlag"] = true
			(*out)[prefixedKey] = s
		} else if m, ok := (*in)[key].(map[string]interface{}); ok {
			flattenJSON(&m, out, prefixedKey)
		} else if array, ok := (*in)[key].([]interface{}); ok {
			for i, v := range array {
				if s, ok := v.(string); ok {
					(*out)[fmt.Sprintf("%s.%d", prefixedKey, i)] = s
				} else if m, ok := v.(map[string]interface{}); ok {
					flattenJSON(&m, out, fmt.Sprintf("%s.%d", prefixedKey, i))
				}
			}
		}
	}
}
