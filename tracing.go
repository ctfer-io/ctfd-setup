package ctfdsetup

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

const (
	serviceName = "ctfd-setup"
)

var (
	tracerProvider *sdktrace.TracerProvider

	Tracer trace.Tracer = tracenoop.NewTracerProvider().Tracer(serviceName)
)

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func setupTraceProvider(ctx context.Context, r *resource.Resource) error {
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return err
	}

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(r),
	)
	Tracer = tracerProvider.Tracer(serviceName)
	return nil
}

func SetupOtelSDK(ctx context.Context, version string) (shutdown func(context.Context) error, err error) {
	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, err
	}

	// Set up trace provider.
	if nerr := setupTraceProvider(ctx, r); nerr != nil {
		return nil, err
	}
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider.Shutdown, nil
}

func StartAPISpan(ctx context.Context) (context.Context, trace.Span) {
	method := getCallerFunctionName()

	return Tracer.Start(
		ctx,
		fmt.Sprintf("api/%s", method),
	)
}

func getCallerFunctionName() string {
	pc, _, _, _ := runtime.Caller(2)
	fn := runtime.FuncForPC(pc)
	method := "unknown"
	if fn != nil {
		if idx := strings.LastIndex(fn.Name(), "."); idx != -1 {
			method = fn.Name()[idx+1:]
		}
	}
	return method
}
