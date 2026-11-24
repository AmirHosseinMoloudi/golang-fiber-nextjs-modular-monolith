package observability

import (
	"context"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Logger zerolog.Logger
	Tracer trace.Tracer

	// Prometheus RED Metrics
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total count of HTTP requests processed by path, method and status code.",
		},
		[]string{"path", "method", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"path", "method"},
	)

	DomainEventsPublishedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "domain_events_published_total",
			Help: "Total count of published domain events.",
		},
		[]string{"event_type"},
	)

	ActiveTodosGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_todos_total",
			Help: "Current count of active todos by status and priority.",
		},
		[]string{"status", "priority"},
	)
)

func InitLogger(env string) zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	if env == "production" {
		Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	} else {
		Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	}
	return Logger
}

func InitTracer(ctx context.Context, serviceName, otelEndpoint string) (*sdktrace.TracerProvider, error) {
	if otelEndpoint == "" {
		// Default to noop tracer provider if endpoint not provided
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		Tracer = tp.Tracer(serviceName)
		return tp, nil
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		otlptracegrpc.WithTimeout(3*time.Second),
	)
	if err != nil {
		Logger.Warn().Err(err).Msg("Could not connect to OTel collector, falling back to local tracer")
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		Tracer = tp.Tracer(serviceName)
		return tp, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	Tracer = tp.Tracer(serviceName)
	return tp, nil
}
