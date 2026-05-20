package metricsCollector

import (
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware(collector MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		collector.CountHttpRequest(c.Request.URL.Path, time.Since(start))
	}
}
