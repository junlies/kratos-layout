package trace

import (
	"context"
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func NewMeter(name string, provider metric.MeterProvider) (metric.Meter, error) {
	return provider.Meter(name), nil
}

func NewMeterProvider(name string, bc *conf.Bootstrap) (metric.MeterProvider, error) {
	metricConf := bc.GetOtel().Metric
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	if metricConf.EnableExemplar {
		err = metrics.EnableOTELExemplar()
		if err != nil {
			return nil, err
		}
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(name),
				attribute.String("env", bc.GetEnv().String()),
			),
		),
		sdkmetric.WithReader(exporter),
		sdkmetric.WithView(
			metrics.DefaultSecondsHistogramView(metrics.DefaultServerSecondsHistogramName),
		),
	)
	otel.SetMeterProvider(provider)
	return provider, nil
}

func NewTracerProvider(ctx context.Context, name string, bc *conf.Bootstrap, textMapPropagator propagation.TextMapPropagator) (trace.TracerProvider, error) {
	traceConf := bc.Otel.Trace
	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(traceConf.Endpoint)}
	if traceConf.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	client := otlptracegrpc.NewClient(opts...)
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(name),
				attribute.String("env", bc.GetEnv().String()),
			),
		),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(textMapPropagator)
	return tp, nil
}

func NewTracer(name string, tp trace.TracerProvider) (trace.Tracer, error) {
	return tp.Tracer(name), nil
}

func NewTextMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		tracing.Metadata{},
		propagation.Baggage{},
		propagation.TraceContext{},
	)
}
