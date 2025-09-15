package dao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/xinkaiwang/shardmanager/libs/xklib/kerror"
)

var (
	currentUploader Uploader
)

func GetUploader() Uploader {
	if currentUploader == nil {
		currentUploader = NewSplunkUploader()
	}
	return currentUploader
}

type Uploader interface {
	Upload(object map[string]interface{}) string
}

type SplunkUploader struct {
	client *http.Client
}

func NewSplunkUploader() *SplunkUploader {
	return &SplunkUploader{
		client: &http.Client{},
	}
}

func GetSplunkEndpoint() string {
	str := os.Getenv("SPLUNK_ENDPOINT") // exp: https://<host>:8088
	if str == "" {
		ke := kerror.Create("SPLUNK_ENDPOINTNotSet", "")
		panic(ke)
	}
	return str
}

func GetSplunkToken() string {
	str := os.Getenv("SPLUNK_TOKEN") // exp: 58DE661B-AA5A-44C2-A658-16E8D240E78A
	if str == "" {
		ke := kerror.Create("SPLUNK_TOKENNotSet", "")
		panic(ke)
	}
	return str
}

// Splunk HTTP Event Collector (HEC)
/*
	curl -k https://<host>:8088/services/collector/event \
	  -H "Authorization: Splunk <TOKEN>" \
	  -H "Content-Type: application/json" \
	  -d '{"event":{"msg":"hello","level":"info"},"sourcetype":"json","index":"main"}'
*/
func (uploader *SplunkUploader) Upload(object map[string]interface{}) string {
	// prepare data
	url := fmt.Sprintf("%s/services/collector/event", GetSplunkEndpoint())
	token := GetSplunkToken()

	// Wrap the event data in the proper Splunk HEC format
	splunkEvent := map[string]interface{}{
		"event":      object,
		"sourcetype": "json",
		"index":      "main",
	}

	jsonData, err := json.Marshal(splunkEvent)
	if err != nil {
		ke := kerror.Wrap(err, "MarshallingFailed", "", false)
		panic(ke)
	}

	// prepare request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		ke := kerror.Wrap(err, "NewRequestFailed", "", false)
		panic(ke)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Splunk %s", token))
	request.Header.Set("Content-Type", "application/json")

	// send request
	response, err := uploader.client.Do(request)
	if err != nil {
		ke := kerror.Wrap(err, "SendRequestFailed", "", false)
		panic(ke)
	}
	defer response.Body.Close()

	// Read response body for debugging
	if response.StatusCode != 200 {
		body, readErr := io.ReadAll(response.Body)
		if readErr == nil {
			fmt.Printf("Splunk error response: %s\n", string(body))
		}
	}

	return response.Status
}
