package metricsCollector

import (
	"time"
)

type MetricsCollector interface {
	CountUnexpectedError(label string)
	CountExpectedError(label string)
	CountHttpRequest(uri string, duration time.Duration)
}
