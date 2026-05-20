// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"gl.eda1.ru/go/go-service-template/config"
	"gl.eda1.ru/go/go-service-template/internal/controller/grpcHandlers"
	v1 "gl.eda1.ru/go/go-service-template/internal/controller/http/v1"
	"gl.eda1.ru/go/go-service-template/internal/domain"
	"gl.eda1.ru/go/go-service-template/internal/usecase"
	"gl.eda1.ru/go/go-service-template/pkg/httpserver"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
	"gl.eda1.ru/go/go-service-template/pkg/metricsCollector"
	"gl.eda1.ru/go/go-service-template/pkg/mysql"
	"gl.eda1.ru/go/go-service-template/pkg/tracer"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.NewLogger(slogger(&cfg.Log))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Metrcics collector
	metrics := metricsCollector.NewPrometheusMetricsHttpCollector()

	// Traces provider
	traces := tracer.NewOtelTraceProvider(cfg.Tracer)

	// mongoServer, err := mongo.New(ctx, cfg.Mongo.URI, cfg.Mongo.Timeout, l)
	// if err != nil {
	//	l.Error(fmt.Errorf("app - Run - mongo.New: %w", err))
	//}

	// conn := mysqlGorm.GetConnector(&cfg.MySQL)
	// err := runAutoMigrations(conn)
	// if err != nil {
	//	l.Error(fmt.Errorf("app - Run - runAutoMigrations: %w", err))
	//	return
	// }
	conn := mysql.GetConnector(&cfg.MySQL, traces)

	repositories := domain.NewRepositories(l, conn)

	uc := usecase.New(ctx, l, repositories)

	// RabbitMQ RPC Server
	// rmqRouter := amqprpc.NewRouter(svcs)
	//
	// rmqServer, err := server.New(
	//	ctx,
	//	cfg.RMQ.DsnList,
	//	cfg.RMQ.ServerExchangeName,
	//	cfg.RMQ.ServerExchangeType,
	//	cfg.RMQ.Queues,
	//	rmqRouter,
	//	l)
	// if err != nil {
	//	 l.Fatal(fmt.Errorf("app - Run - rmqServer - server.New: %w", err))
	// }
	//
	// go rmqServer.WatchDog(ctx)

	// GRPC Server
	go grpcHandlers.NewHandler(cfg, uc, l, metrics, traces)

	// HTTP Server
	handler := gin.New()
	handler.Use(metricsCollector.MetricsMiddleware(metrics))
	handler.Use(tracer.TracerMiddleware(traces))
	v1.NewRouter(handler, uc, l)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info(context.Background(), "app - Run - signal: "+s.String())
	case err := <-httpServer.Notify():
		l.Error(context.Background(), fmt.Sprintf("app - Run - httpServer.Notify: %v", err))
	}

	cancel()
	// Shutdown
	err := httpServer.Shutdown()
	if err != nil {
		l.Error(context.Background(), fmt.Sprintf("app - Run - httpServer.Shutdown: %v", err))
	}

	// err = rmqServer.Shutdown()
	// if err != nil {
	//	l.Error(fmt.Errorf("app - Run - rmqServer.Shutdown: %w", err))
	//}

	// Close MySQL connections
	conn.Close()

	traces.Shutdown()

	// err = mongoServer.Shutdown()
	// if err != nil {
	//	l.Error(fmt.Errorf("app - Run - mongoServer.Shutdown: %w", err))
	//}
}

/*func runAutoMigrations(conn *mysqlGorm.Connector) error {
	dbCon, err := conn.GetMaster(context.TODO())
	if err != nil {
		return err
	}

	return dbCon.AutoMigrate(&example.ExampleCounter{})
}*/
