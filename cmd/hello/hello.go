package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xinkaiwang/hermes/internal/common"
	"github.com/xinkaiwang/hermes/internal/dao"
	"github.com/xinkaiwang/shardmanager/libs/xklib/kcommon"
	"github.com/xinkaiwang/shardmanager/libs/xklib/klogging"
)

func main() {
	ctx := context.Background()
	logLevel := kcommon.GetEnvString("LOG_LEVEL", "debug")
	logFormat := kcommon.GetEnvString("LOG_FORMAT", "json")
	klogging.SetDefaultLogger(klogging.NewLogrusLogger(ctx).SetConfig(ctx, logLevel, logFormat))

	fmt.Println("Hello, World!")
	fmt.Println(common.GetVersion())
	// test(ctx)

	klogging.Info(ctx).Log("test", "payload")
	testBatch(ctx)

	<-time.After(1 * time.Second)
}

func testBatch(ctx context.Context) {
	batchUploader := dao.NewBatchUploader(ctx)
	eve := &dao.EventJson{
		Event: map[string]interface{}{
			"event": "test",
			"msg":   "hello",
			"level": "debug",
		},
		Time:       time.Now().UnixMilli(),
		Host:       "127.0.0.1",
		Source:     "batch",
		SourceType: "json",
		Index:      "main",
	}
	batchUploader.ChEvents <- eve
}

func test(ctx context.Context) {
	uploader := dao.GetUploader()
	payload := map[string]interface{}{
		"event": "test",
		"msg":   "hello",
		"level": "info",
	}
	eventJson := &dao.EventJson{
		Event:      payload,
		Time:       time.Now().UnixMilli(),
		Host:       "127.0.0.1",
		Source:     "test",
		SourceType: "json",
		Index:      "main",
	}
	ke := kcommon.TryCatchRun(ctx, func() {
		status := uploader.Upload(eventJson)
		fmt.Println(status)
	})
	if ke != nil {
		fmt.Println(ke.FullString())
	}

}
