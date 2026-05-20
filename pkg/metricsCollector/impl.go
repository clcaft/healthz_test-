package metricsCollector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type prometheusMetricsHttpCollector struct {
	httpRequestSummary      *prometheus.SummaryVec
	expectedErrorsCounter   *prometheus.CounterVec
	unexpectedErrorsCounter *prometheus.CounterVec
}

func NewPrometheusMetricsHttpCollector() MetricsCollector {
	httpRequestSummary := promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_summary",
			Help: "HTTP request duration",
		},
		[]string{"uri"},
	)
	expectedErrorsCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "expected_error_counter",
			Help: "Count of expected errors",
		},
		[]string{"label"},
	)
	unexpectedErrorsCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "unexpected_error_counter",
			Help: "Count of unexpected errors",
		},
		[]string{"label"},
	)

	return &prometheusMetricsHttpCollector{
		httpRequestSummary:      httpRequestSummary,
		expectedErrorsCounter:   expectedErrorsCounter,
		unexpectedErrorsCounter: unexpectedErrorsCounter,
	}
}

func (p *prometheusMetricsHttpCollector) CountHttpRequest(uri string, duration time.Duration) {
	p.httpRequestSummary.WithLabelValues(uri).Observe(duration.Seconds())
}

func (p *prometheusMetricsHttpCollector) CountUnexpectedError(label string) {
	p.unexpectedErrorsCounter.WithLabelValues(label).Inc()
}

func (p *prometheusMetricsHttpCollector) CountExpectedError(label string) {
	p.expectedErrorsCounter.WithLabelValues(label).Inc()
}
