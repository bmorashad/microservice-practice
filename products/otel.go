package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

var tracer trace.Tracer

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context, serviceName, serviceVersion string) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = fn(ctx)
		}
		shutdownFuncs = nil
		return err
	}
	// return otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint("localhost:4317"), otlptracegrpc.WithInsecure())
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
	)

	// Set up resource.
	r, err := newResource(serviceName, serviceVersion)
	if err != nil {
		log.Fatalf(err.Error(), shutdown(ctx))
		return
	}

	// Set up propagator.
	prop := newPropagator()
	exp, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatalf(err.Error(), shutdown(ctx))
		return
	}
	tp := newTracerProvider(exp, r, prop)
	otel.SetTextMapPropagator(prop)
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(serviceName)
	return
}

func newTracerProvider(exp sdktrace.SpanExporter, resource *resource.Resource, prop propagation.TextMapPropagator) *sdktrace.TracerProvider {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource),
	)
	return tp
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		))
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
