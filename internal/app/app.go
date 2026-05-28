package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"gl.eda1.ru/go/go-service-template/config"
	v1 "gl.eda1.ru/go/go-service-template/internal/controller/http/v1"
	"gl.eda1.ru/go/go-service-template/internal/domain"
	"gl.eda1.ru/go/go-service-template/internal/usecase"
	"gl.eda1.ru/go/go-service-template/pkg/httpserver"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
	"gl.eda1.ru/go/go-service-template/pkg/metricsCollector"
	"gl.eda1.ru/go/go-service-template/pkg/mysql"
	"gl.eda1.ru/go/go-service-template/pkg/tracer"
)

func Run(cfg *config.Config) {
	l := logger.NewLogger(slogger(&cfg.Log))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Metrcics collector
	metrics := metricsCollector.NewPrometheusMetricsHttpCollector()

	// Traces provider
	traces := tracer.NewOtelTraceProvider(cfg.Tracer)

	// MySQL - опционально, без фатальной ошибки
	var conn *mysql.Connector
	if cfg.MySQL.Host != "" {
		conn = mysql.GetConnector(&cfg.MySQL, traces)
		defer conn.Close()
		l.Info(ctx, "MySQL connected")
	} else {
		l.Info(ctx, "MySQL disabled (no config)")
	}

	repositories := domain.NewRepositories(l, conn)
	uc := usecase.New(ctx, l, repositories)

	// HTTP Server
	handler := gin.New()
	handler.Use(metricsCollector.MetricsMiddleware(metrics))
	handler.Use(tracer.TracerMiddleware(traces))
	v1.NewRouter(handler, uc, l, cfg)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info(ctx, "app - Run - signal: "+s.String())
	case err := <-httpServer.Notify():
		l.Error(ctx, fmt.Sprintf("app - Run - httpServer.Notify: %v", err))
	}

	cancel()

	// Shutdown
	if err := httpServer.Shutdown(); err != nil {
		l.Error(ctx, fmt.Sprintf("app - Run - httpServer.Shutdown: %v", err))
	}

	traces.Shutdown()
}
