package grpcHandlers

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"gl.eda1.ru/go/go-service-template/config"
	proto "gl.eda1.ru/go/go-service-template/internal/proto/go"
	"gl.eda1.ru/go/go-service-template/internal/usecase"
	"gl.eda1.ru/go/go-service-template/pkg/logger"
	"gl.eda1.ru/go/go-service-template/pkg/metricsCollector"
	"gl.eda1.ru/go/go-service-template/pkg/tracer"
)

func NewHandler(cfg *config.Config, uc *usecase.UseCases, l logger.Interface, metrics metricsCollector.MetricsCollector, traces tracer.TraceProvider) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	ctx := context.Background()
	if err != nil {
		l.Fatal(ctx, "Couldn't create connection tcp %v", err)
	}
	// server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	server := grpc.NewServer(grpc.UnaryInterceptor(observabilityUnaryInterceptor(l, metrics, traces)))

	proto.RegisterExampleServer(server, NewExampleHandler(l, uc))

	l.Info(ctx, "Start GRPC server at %d", cfg.GRPC.Port)
	// Register reflection service on gRPC server.
	if cfg.GRPC.EnableServerReflection {
		l.Warn(ctx, "GRPC server reflection is enabled. Disable it in production")
		reflection.Register(server)
	}

	if err = server.Serve(lis); err != nil {
		l.Fatal(ctx, "Couldn't start GRPC server %v", err)
	}
}

// UnaryInterceptor is a simple logger interceptor
func observabilityUnaryInterceptor(l logger.Interface, metrics metricsCollector.MetricsCollector, traces tracer.TraceProvider) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// распределённая трассировка
		md, ok := metadata.FromIncomingContext(ctx)
		if ok && md["metadata"] != nil {
			if len(md["metadata"]) == 2 {
				traceId := md["metadata"][0]
				spanId := md["metadata"][1]
				newCtx, span := traces.Start(ctx, info.FullMethod, traceId, spanId)
				defer span.End()

				ctx = newCtx
			} else {
				fmt.Printf("Invalid metadata. Expected array of traceId and spanId, got %v\n", md["metadata"])
			}
		}

		start := time.Now()
		res, err := handler(ctx, req)
		metrics.CountHttpRequest(info.FullMethod, time.Since(start))

		if err != nil {
			l.Error(ctx, "GRPC call %s returned error: %v", info.FullMethod, err)
			metrics.CountExpectedError(info.FullMethod)
		}

		return res, err
	}
}
