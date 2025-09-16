package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/xinkaiwang/hermes/internal/biz"
	"github.com/xinkaiwang/hermes/internal/common"
	"github.com/xinkaiwang/hermes/internal/handler"
	"github.com/xinkaiwang/shardmanager/libs/xklib/kcommon"
	"github.com/xinkaiwang/shardmanager/libs/xklib/klogging"
	"github.com/xinkaiwang/shardmanager/libs/xklib/kmetrics"
	"github.com/xinkaiwang/shardmanager/libs/xklib/ksysmetrics"
	"go.opencensus.io/metric/metricproducer"
)

func main() {
	ctx := context.Background()
	logLevel := kcommon.GetEnvString("LOG_LEVEL", "debug")
	logFormat := kcommon.GetEnvString("LOG_FORMAT", "json")
	klogging.SetDefaultLogger(klogging.NewLogrusLogger(ctx).SetConfig(ctx, logLevel, logFormat))

	// 记录启动信息
	klogging.Info(ctx).With("version", common.GetVersion()).With("commit", common.GetGitCommit()).With("buildTime", common.GetBuildTime()).With("logLevel", logLevel).With("logFormat", logFormat).With("now", time.Now().Format(time.RFC3339)).Log("ServerStarting", "Starting hermes")

	// 创建 Prometheus 导出器
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "hellosvc",
	})
	if err != nil {
		log.Fatalf("Failed to create Prometheus exporter: %v", err)
	}

	// 注册 kmetrics 注册表
	registry := kmetrics.GetKmetricsRegistry()
	metricproducer.GlobalManager().AddProducer(registry)

	// 注册系统指标注册表
	metricproducer.GlobalManager().AddProducer(ksysmetrics.GetRegistry())

	// 启动系统指标收集器
	ksysmetrics.StartSysMetricsCollector(ctx, 15*time.Second, common.GetVersion())

	// 获取端口配置
	apiPort := kcommon.GetEnvInt("API_PORT", 8080)
	metricsPort := kcommon.GetEnvInt("METRICS_PORT", 9090)

	// 创建 metrics 路由
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", pe)

	// 创建主路由
	app := biz.NewApp(ctx)
	handler := handler.NewHandler(app)
	mainMux := http.NewServeMux()
	handler.RegisterRoutes(mainMux)

	// 创建主 HTTP 服务器
	mainServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", apiPort),
		Handler: mainMux,
	}

	// 创建 metrics HTTP 服务器
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", metricsPort),
		Handler: metricsMux,
	}

	// 记录端口配置
	klogging.Info(ctx).
		With("api_port", apiPort).
		With("metrics_port", metricsPort).
		Log("ServerConfig", "Server ports configuration")

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		klogging.Info(ctx).Log("ServerShutdown", "Shutting down servers...")
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := mainServer.Shutdown(ctx); err != nil {
			klogging.Error(ctx).With("error", err).Log("MainServerShutdownError", "Main server shutdown error")
		}
		if err := metricsServer.Shutdown(ctx); err != nil {
			klogging.Error(ctx).With("error", err).Log("MetricsServerShutdownError", "Metrics server shutdown error")
		}
	}()

	// 启动 metrics 服务器
	go func() {
		klogging.Info(ctx).With("addr", metricsServer.Addr).Log("MetricsServerStarting", "Metrics server starting")
		if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			klogging.Error(ctx).With("error", err).Log("MetricsServerError", "Metrics server error")
		}
	}()

	// 启动主服务器
	klogging.Info(ctx).With("addr", mainServer.Addr).Log("MainServerStarting", "Main server starting")
	if err := mainServer.ListenAndServe(); err != http.ErrServerClosed {
		klogging.Error(ctx).With("error", err).Log("MainServerError", "Main server error")
	}
	klogging.Info(ctx).Log("ServerShutdown", "Servers stopped")
}
