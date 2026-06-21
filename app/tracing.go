package app

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	applog "github.com/example/dd-frame/pkg/log"
)

// InitTracing 初始化 OpenTelemetry TracerProvider
//
// 返回 shutdown 函数，应在 main 中 defer 调用。
// 未启用或配置不完整时返回 Noop provider（零开销）。
func InitTracing(cfg *TracingConfig) func(context.Context) error {
	if !cfg.Enabled {
		applog.Info("tracing disabled")
		return func(_ context.Context) error { return nil }
	}

	ctx := context.Background()

	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "dd-frame"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		applog.Error("tracing: failed to create resource", "err", err)
		return func(_ context.Context) error { return nil }
	}

	// OTLP HTTP exporter
	opts := []otlptracehttp.Option{}
	if cfg.Endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		applog.Error("tracing: failed to create exporter", "err", err)
		return func(_ context.Context) error { return nil }
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	applog.Info("tracing initialized",
		"service", serviceName,
		"endpoint", cfg.Endpoint,
		"insecure", cfg.Insecure,
	)

	return tp.Shutdown
}
