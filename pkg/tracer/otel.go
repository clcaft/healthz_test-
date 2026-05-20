package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"gl.eda1.ru/go/go-service-template/config"
)

type otelTracerProvider struct {
	enabled  bool
	tracer   trace.Tracer
	shutdown func()
}

type otelSpan struct {
	span *trace.Span
}

type traceStartedKeyType int

const traceStartedKey traceStartedKeyType = iota

func NewOtelTraceProvider(cfg config.Tracer) TraceProvider {
	if cfg.TracerCollector == "" {
		return &otelTracerProvider{}
	}

	grpcConnection, err := grpc.NewClient(cfg.TracerCollector,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("Failed to init tracer gPRC connection: %v\n", err)
		return &otelTracerProvider{}
	}

	ctx := context.Background()
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(grpcConnection))
	if err != nil {
		fmt.Printf("Failed to create trace exporter: %v\n", err)
		return &otelTracerProvider{}
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(getResource(cfg.TracerService)),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	shutdownFn := func() {
		_ = tracerProvider.Shutdown(ctx)
	}

	tracer := otel.Tracer(cfg.TracerService)

	return &otelTracerProvider{
		enabled:  true,
		shutdown: shutdownFn,
		tracer:   tracer,
	}
}

func getResource(serviceName string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)
}

func (o *otelTracerProvider) Shutdown() {
	if !o.enabled {
		return
	}
	o.shutdown()
}

func (o *otelTracerProvider) Span(ctx context.Context, name string) (context.Context, Span) {
	if !o.enabled || !isTraceStarted(ctx) {
		return ctx, &otelSpan{}
	}

	newCtx, span := o.tracer.Start(ctx, name)

	return newCtx, &otelSpan{
		span: &span,
	}
}

func (o *otelTracerProvider) Start(ctx context.Context, name, traceId, spanId string) (context.Context, Span) {
	if !o.enabled {
		return ctx, &otelSpan{}
	}
	var err error
	if traceId != "" {
		ctx, err = createContextWithTraceId(ctx, traceId, spanId)
		if err != nil {
			return ctx, &otelSpan{}
		}
	}

	ctx, span := o.tracer.Start(ctx, name)
	ctx = context.WithValue(ctx, traceStartedKey, true)

	return ctx, &otelSpan{
		span: &span,
	}
}

func createContextWithTraceId(ctx context.Context, traceId, spanId string) (context.Context, error) {
	parsedTraceId, err := trace.TraceIDFromHex(traceId)
	if err != nil {
		fmt.Printf("Unable to parse traceId traceId %s: %v\n", traceId, err)
		return ctx, err
	}
	parsedSpanId, err := trace.SpanIDFromHex(spanId)
	if err != nil {
		fmt.Printf("Unable to parse spanId traceId %s: %v\n", spanId, err)
		parsedSpanId = trace.SpanID{}
	}
	ctx = trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: parsedTraceId,
		SpanID:  parsedSpanId,
		Remote:  true,
	}))
	return ctx, nil
}

func isTraceStarted(ctx context.Context) bool {
	if started, ok := ctx.Value(traceStartedKey).(bool); ok {
		return started
	}
	return false
}

func (s *otelSpan) End() {
	if s.span != nil {
		(*s.span).End()
	}
}
