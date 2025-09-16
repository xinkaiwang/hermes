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

type EventJson struct {
	Event      map[string]interface{} `json:"event"`                // exp: {"msg":"hello","level":"info"}
	Time       int64                  `json:"time,omitempty"`       // epoch ms exp: 1726339200000
	Host       string                 `json:"host,omitempty"`       // exp: 127.0.0.1
	Source     string                 `json:"source,omitempty"`     // exp: /var/log/syslog
	SourceType string                 `json:"sourcetype,omitempty"` // exp: syslog
	Index      string                 `json:"index,omitempty"`      // exp: main
}

func GetUploader() Uploader {
	if currentUploader == nil {
		currentUploader = NewSplunkUploader()
	}
	return currentUploader
}

type Uploader interface {
	Upload(eventJson *EventJson) string
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
	str := os.Getenv("SPLUNK_TOKEN") // exp: 58DE661B-AA5A-44C2-A658-XXXXXXXXXXXX
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
func (uploader *SplunkUploader) Upload(eve *EventJson) string {
	// prepare data
	url := fmt.Sprintf("%s/services/collector/event", GetSplunkEndpoint())
	token := GetSplunkToken()

	// // Wrap the event data in the proper Splunk HEC format
	// splunkEvent := map[string]interface{}{
	// 	"event":      object,
	// 	"sourcetype": "json",
	// 	"index":      "main",
	// }

	jsonData, err := json.Marshal(eve)
	if err != nil {
		ke := kerror.Wrap(err, "MarshallingFailed", "", false)
		panic(ke)
	}
	fmt.Println(string(jsonData))

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
