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

func (a *App) Post(ctx context.Context, req api.PostRequest) api.PostResponse {
	for _, event := range req.Events {
		eve := &dao.EventJson{
			Event: event,
		}
		if req.Host != "" {
			eve.Host = req.Host
		}
		if req.Source != "" {
			eve.Source = req.Source
		} else {
			eve.Source = "hermes"
		}
		if req.SourceType != "" {
			eve.SourceType = req.SourceType
		} else {
			eve.SourceType = "json"
		}
		if req.Index != "" {
			eve.Index = req.Index
		} else {
			eve.Index = "main"
		}
		a.batchUploader.ChEvents <- eve
	}
	return api.PostResponse{
		Count: len(req.Events),
	}
}
