package biz

import (
	"context"
	"time"

	"github.com/xinkaiwang/hermes/api"
	"github.com/xinkaiwang/hermes/internal/common"
	"github.com/xinkaiwang/hermes/internal/dao"
)

type App struct {
	ctx           context.Context
	batchUploader *dao.BatchUploader
}

func NewApp(ctx context.Context) *App {
	return &App{
		ctx:           ctx,
		batchUploader: dao.NewBatchUploader(ctx),
	}
}

func (a *App) Ping(ctx context.Context) api.PingResponse {
	return api.PingResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   common.GetVersion(),
	}
}
