package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// CounterOpts Counter 指标选项。
type CounterOpts struct {
	Name string
	Help string
}

// NewCounter 创建并注册 Counter 指标。
func NewCounter(opts CounterOpts) prometheus.Counter {
	return promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace(),
		Name:      opts.Name,
		Help:      opts.Help,
	})
}

// NewCounterVec 创建并注册带标签的 Counter 指标。
func NewCounterVec(opts CounterOpts, labelNames ...string) *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace(),
		Name:      opts.Name,
		Help:      opts.Help,
	}, labelNames)
}
