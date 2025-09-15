package main

import (
	"context"
	"fmt"

	"github.com/xinkaiwang/hermes/internal/common"
	"github.com/xinkaiwang/hermes/internal/dao"
	"github.com/xinkaiwang/shardmanager/libs/xklib/kcommon"
)

func main() {
	ctx := context.Background()
	fmt.Println("Hello, World!")
	fmt.Println(common.GetVersion())

	uploader := dao.GetUploader()
	payload := map[string]interface{}{
		"event": "test",
		"msg":   "hello",
		"level": "info",
	}
	ke := kcommon.TryCatchRun(ctx, func() {
		status := uploader.Upload(payload)
		fmt.Println(status)
	})
	if ke != nil {
		fmt.Println(ke.FullString())
	}
}
