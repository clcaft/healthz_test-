package tracer

import "github.com/gin-gonic/gin"

func TracerMiddleware(traces TraceProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := traces.Span(c.Request.Context(), c.Request.URL.Path)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		span.End()
	}
}
