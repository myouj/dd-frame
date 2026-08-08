package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpOnce             sync.Once
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge
)

func initHTTPMetrics() {
	httpOnce.Do(func() {
		httpRequestsTotal = NewCounterVec(CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		}, "method", "path", "status")

		httpRequestDuration = NewHistogramVec(HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, "method", "path")

		httpRequestsInFlight = NewGauge(GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently in flight",
		})
	})
}

// HTTPMiddleware 返回 Prometheus HTTP 指标采集中间件。
//
// 自动采集每个请求的计数、延迟和并发数。须先调用 Init。
func HTTPMiddleware() gin.HandlerFunc {
	initHTTPMetrics()

	return func(c *gin.Context) {
		start := time.Now()
		httpRequestsInFlight.Inc()

		c.Next()

		httpRequestsInFlight.Dec()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
