package tracing

import (
	"context"
	"flag"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func newZipkinResource(ctx context.Context, serviceName, hostName, hostType string) (*resource.Resource, error) {
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

func newZipkinExporter() (*zipkin.Exporter, error) {
	url := flag.String("zipkin", "http://localhost:9411/api/v2/spans", "zipkin url")
	exporter, err := zipkin.New(*url)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func InitOtelZipkinTrace(ctx context.Context, serviceName, hostName, hostType string) func(ctx context.Context) error {
	traceExporter, err := newZipkinExporter()
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
