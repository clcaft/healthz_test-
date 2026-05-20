package tracer

import "context"

type Span interface {
	End()
}

type TraceProvider interface {
	Start(ctx context.Context, name, traceId, spanId string) (context.Context, Span)
	Span(ctx context.Context, name string) (context.Context, Span)
	Shutdown()
}
