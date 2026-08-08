package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// GaugeOpts Gauge 指标选项。
type GaugeOpts struct {
	Name string
	Help string
}

// NewGauge 创建并注册 Gauge 指标。
func NewGauge(opts GaugeOpts) prometheus.Gauge {
	return promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace(),
		Name:      opts.Name,
		Help:      opts.Help,
	})
}

// NewGaugeVec 创建并注册带标签的 Gauge 指标。
func NewGaugeVec(opts GaugeOpts, labelNames ...string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace(),
		Name:      opts.Name,
		Help:      opts.Help,
	}, labelNames)
}
