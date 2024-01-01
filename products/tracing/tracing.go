package tracing

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func newResource(ctx context.Context, serviceName, hostName, hostType string) (*resource.Resource, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			// semconv.ServiceNamespaceKey.String("US-West-1"),
			semconv.HostNameKey.String(hostName),
			semconv.HostTypeKey.String(hostType),
		),
	)
	return res, err
}

func newExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	return otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithDialOption(),
	)
}

func nexExporterOld(ctx context.Context) (*otlptrace.Exporter, error) {
	client := otlptracegrpc.NewClient(otlptracegrpc.WithInsecure())
	return otlptrace.New(ctx, client)
}

func InitOtelTrace(ctx context.Context, serviceName, hostName, hostType string) func(ctx context.Context) error {
	traceExporter, err := newExporter(ctx)
	if err != nil {
		log.Fatalf("failed to initialize stdouttrace export pipeline: %v", err)
	}
	res, err := newResource(ctx, serviceName, hostName, hostType)
	if err != nil {
		log.Fatal("Error while creating otel resource")
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		// sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	// defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)
	tracerName := fmt.Sprintf("%s-tracer", serviceName)
	tp.Tracer(tracerName)
	// priority := attribute.Key("business.priority")
	// appEnv := attribute.Key("prod.env")
	return tp.Shutdown
}
