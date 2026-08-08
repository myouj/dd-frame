package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HistogramOpts Histogram 指标选项。
type HistogramOpts struct {
	Name    string
	Help    string
	Buckets []float64 // 为空时使用 prometheus.DefBuckets
}

// NewHistogram 创建并注册 Histogram 指标。
func NewHistogram(opts HistogramOpts) prometheus.Histogram {
	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	return promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace(),
		Name:      opts.Name,
		Help:      opts.Help,
		Buckets:   buckets,
	})
}

// NewHistogramVec 创建并注册带标签的 Histogram 指标。
func NewHistogramVec(opts HistogramOpts, labelNames ...string) *prometheus.HistogramVec {
	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	return promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace(),
		Name:      opts.Name,
		Help:      opts.Help,
		Buckets:   buckets,
	}, labelNames)
}
