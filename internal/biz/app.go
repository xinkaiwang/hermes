package biz

import (
	"context"
	"strconv"
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

func (a *App) Post(ctx context.Context, req api.PostRequest, remoteAddr string) api.PostResponse {
	for _, event := range req.Events {
		eve := &dao.EventJson{
			Event: event,
		}
		eve.Time = parseTime(event["time"])
		if req.Host != "" {
			eve.Host = req.Host
		} else {
			eve.Host = remoteAddr
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

func parseTime(timeVal interface{}) int64 {
	if timeVal == nil {
		return time.Now().UnixMilli()
	}
	if timeStr, ok := timeVal.(string); ok {
		// try to parse to int64
		int64Val, err := strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			return int64Val
		}

		// try to parse to time.Time
		t, err := time.Parse(time.RFC3339, timeStr)
		if err == nil {
			return t.UnixMilli()
		}
	}
	if int64Val, ok := timeVal.(int64); ok {
		return int64Val
	}
	return time.Now().UnixMilli()
}
