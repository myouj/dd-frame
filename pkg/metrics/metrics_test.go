package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestInitNamespaceApplied(t *testing.T) {
	Init(Config{Enabled: true, Namespace: "testapp"})

	counter := NewCounter(CounterOpts{
		Name: "events_total",
		Help: "test counter",
	})
	counter.Inc()

	metric := findMetric(t, "testapp_events_total")
	if metric.GetCounter().GetValue() != 1 {
		t.Errorf("counter value = %v, want 1", metric.GetCounter().GetValue())
	}
}

func TestHTTPMiddlewareRecordsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Init(Config{Enabled: true})

	r := gin.New()
	r.Use(HTTPMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	metric := findMetric(t, "http_requests_total")
	if metric.GetCounter().GetValue() != 1 {
		t.Errorf("http_requests_total = %v, want 1", metric.GetCounter().GetValue())
	}
}

func findMetric(t *testing.T, name string) *dto.Metric {
	t.Helper()

	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	for _, mf := range mfs {
		if mf.GetName() == name {
			metrics := mf.GetMetric()
			if len(metrics) == 0 {
				t.Fatalf("metric %q has no samples", name)
			}
			return metrics[0]
		}
	}
	t.Fatalf("metric %q not found", name)
	return nil
}
