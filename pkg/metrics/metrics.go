// Package metrics 提供 Prometheus 指标注册与 HTTP 采集工具。
//
// 启动时调用 Init 注入全局配置，业务模块通过 NewCounter/NewGauge/NewHistogram
// 等工厂函数注册指标，统一应用 namespace 前缀。
package metrics

// Config Prometheus 指标全局配置。
type Config struct {
	Enabled   bool
	Namespace string
}

var cfg Config

// Init 初始化指标工具，应在应用启动早期调用。
func Init(c Config) {
	cfg = c
}

// Enabled 返回指标是否启用（与配置 metrics.enabled 一致）。
func Enabled() bool {
	return cfg.Enabled
}

func namespace() string {
	return cfg.Namespace
}
