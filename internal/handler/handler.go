package handler

import (
	"encoding/json"
	"net/http"

	"github.com/xinkaiwang/hermes/api"
	"github.com/xinkaiwang/hermes/internal/biz"
	"github.com/xinkaiwang/shardmanager/libs/xklib/kerror"
	"github.com/xinkaiwang/shardmanager/libs/xklib/klogging"
	"github.com/xinkaiwang/shardmanager/libs/xklib/kmetrics"
)

type Handler struct {
	app *biz.App
}

func NewHandler(app *biz.App) *Handler {
	return &Handler{app: app}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// 包装所有处理器以添加错误处理中间件
	mux.Handle("/api/ping", ErrorHandlingMiddleware(http.HandlerFunc(h.PingHandler)))
	mux.Handle("/api/post", ErrorHandlingMiddleware(http.HandlerFunc(h.PostHandler)))
}

// PingHandler 处理 /api/ping 请求
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只允许 GET 方法
	if r.Method != http.MethodGet {
		panic(kerror.Create("MethodNotAllowed", "only GET method is allowed").
			WithErrorCode(kerror.EC_INVALID_PARAMETER))
	}

	// 记录请求信息
	klogging.Verbose(r.Context()).
		Log("PingRequest", "received ping request")

	// 处理请求
	var resp api.PingResponse
	kmetrics.InstrumentSummaryRunVoid(r.Context(), "biz.Ping", func() {
		resp = h.app.Ping(r.Context())
	}, "")

	// 记录响应信息
	klogging.Info(r.Context()).
		With("status", resp.Status).
		With("version", resp.Version).
		Log("PingResponse", "sending ping response")

	// 返回响应
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(kerror.Create("EncodingError", "failed to encode response").
			WithErrorCode(kerror.EC_INTERNAL_ERROR).
			With("error", err.Error()))
	}
}

// curl -k http://localhost:8080/api/post -d '{"events": [{"modle": "loader.routeMiddleware", "event":"AddMiddleware", "type":"UserData", "route":"/delay", "service":"rain"}]}'
// curl -k http://localhost:8080/api/post -d '{"events": [{"modle": "loader.routeMiddleware", "event":"AddMiddleware", "type":"UserData", "route":"/delay", "service":"rain", "host":"127.0.0.1"},{"event":"AddMiddleware", "type":"UserData2", "route":"/delay2", "service":"rain2", "host":"127.0.0.2"}]}'

func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只允许 POST 方法
	if r.Method != http.MethodPost {
		panic(kerror.Create("MethodNotAllowed", "only POST method is allowed").
			WithErrorCode(kerror.EC_INVALID_PARAMETER))
	}

	var req api.PostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		panic(kerror.Create("DecodingError", "failed to decode request").
			WithErrorCode(kerror.EC_INVALID_PARAMETER).
			With("error", err.Error()))
	}

	// 记录请求信息
	klogging.Verbose(r.Context()).
		Log("PostRequest", "received post request")

	// 处理请求
	var resp api.PostResponse
	kmetrics.InstrumentSummaryRunVoid(r.Context(), "biz.Post", func() {
		resp = h.app.Post(r.Context(), req, r.RemoteAddr)
	}, "")

	// 记录响应信息
	klogging.Info(r.Context()).
		With("count", resp.Count).
		Log("PostResponse", "sending post response")

	// 返回响应
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(kerror.Create("EncodingError", "failed to encode response").
			WithErrorCode(kerror.EC_INTERNAL_ERROR).
			With("error", err.Error()))
	}
}
