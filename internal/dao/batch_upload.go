package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xinkaiwang/shardmanager/libs/xklib/kcommon"
	"github.com/xinkaiwang/shardmanager/libs/xklib/kerror"
	"github.com/xinkaiwang/shardmanager/libs/xklib/klogging"
)

type BatchUploader struct {
	ctx      context.Context
	ChEvents chan *EventJson
	client   *http.Client
}

func NewBatchUploader(ctx context.Context) *BatchUploader {
	bu := &BatchUploader{
		ctx:      ctx,
		ChEvents: make(chan *EventJson, 1000),
		client:   &http.Client{},
	}
	go bu.Start()
	return bu
}

func (b *BatchUploader) Start() {
	maxCount := kcommon.GetEnvInt("MAX_BATCH_COUNT", 100)
	maxSize := kcommon.GetEnvInt("MAX_BATCH_SIZE", 1024*1024)
	maxDelayMs := kcommon.GetEnvInt("MAX_BATCH_DELAY_MS", 100)
	if maxDelayMs <= 0 {
		maxDelayMs = 1
	}
	klogging.Info(b.ctx).With("maxCount", maxCount).With("maxSize", maxSize).With("maxDelayMs", maxDelayMs).Log("BatchUploader", "Start")
	stop := false
	var sb strings.Builder
	batchSize := 0
	for !stop {
		klogging.Debug(b.ctx).Log("BatchUploader", "Step1")
		// forever loop
		// 1. if chan not empty, append to payload
		// 2. upload if a) maxEvents reached or b) maxSize reached or c) chan empty

		select {
		case eve := <-b.ChEvents:
			klogging.Info(b.ctx).With("eve", eve).With("batchSize", batchSize).With("sb", sb.String()).Log("BatchUploader", "Step2")
			if eve == nil {
				stop = true
				break
			}
			if batchSize > 0 {
				sb.WriteString("\n")
			}
			batchSize++
			jsonData, err := json.Marshal(eve)
			if err != nil {
				ke := kerror.Wrap(err, "MarshallingFailed", "", false)
				panic(ke)
			}
			sb.WriteString(string(jsonData))
			klogging.Info(b.ctx).With("eve", eve).With("batchSize", batchSize).With("sb", sb.String()).Log("BatchUploader", "Step2.1")
			if sb.Len() >= maxSize {
				b.Upload(sb.String(), batchSize)
				sb.Reset()
				batchSize = 0
			}
			if batchSize >= maxCount {
				b.Upload(sb.String(), batchSize)
				sb.Reset()
				batchSize = 0
			}
		case <-time.After(time.Duration(maxDelayMs) * time.Millisecond):
			klogging.Debug(b.ctx).With("batchSize", batchSize).Log("BatchUploader", "Step3")
			if batchSize > 0 {
				b.Upload(sb.String(), batchSize)
				sb.Reset()
				batchSize = 0
			}
		}
	}
}

func (b *BatchUploader) Upload(payload string, count int) string { // count is the number of events in the payload, for metrics/logging purpose only
	size := len(payload)
	startTimeMs := kcommon.GetMonoTimeMs()
	klogging.Debug(b.ctx).WithDebug("payload", payload).With("count", count).Log("Upload", "started")
	// prepare data
	url := fmt.Sprintf("%s/services/collector/event", GetSplunkEndpoint())
	token := GetSplunkToken()

	// prepare request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		ke := kerror.Wrap(err, "NewRequestFailed", "", false)
		panic(ke)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Splunk %s", token))
	request.Header.Set("Content-Type", "application/json")

	// send request
	response, err := b.client.Do(request)
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

	elapsedMs := kcommon.GetMonoTimeMs() - startTimeMs
	klogging.Info(b.ctx).With("response", response.Status).With("size", size).With("elapsedMs", elapsedMs).With("count", count).Log("Upload", "Completed")
	return response.Status
}
