package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type nopTracerProvider struct{}

type nopSpan struct {
	span *trace.Span
}

func NewNopTraceProvider() TraceProvider {
	return &nopTracerProvider{}
}

func (n *nopTracerProvider) Start(ctx context.Context, name, traceId, spanId string) (context.Context, Span) {
	return ctx, &nopSpan{}
}

func (n *nopTracerProvider) Span(ctx context.Context, name string) (context.Context, Span) {
	return ctx, &nopSpan{}
}

func (n *nopTracerProvider) Shutdown() {
}

func (n *nopSpan) End() {
}
