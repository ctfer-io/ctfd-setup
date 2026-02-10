package ctfdsetup

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
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
	exp, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return err
	}

	tracerProvider = sdktrace.NewTracerProvider(
		// We need to have the burden of a simple span processor as the process might be short-lived
		// because a batch processor can not give enough time to export data...
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exp)),
		sdktrace.WithResource(r),
	)
	Tracer = tracerProvider.Tracer(serviceName)
	return nil
}

func SetupOtelSDK(ctx context.Context, version string) (shutdown func(context.Context) error, err error) {
	// Set up propagator
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Get existing provider to avoid overrides, and if none set defines our own
	existingProvider := otel.GetTracerProvider()
	if _, isNoop := existingProvider.(tracenoop.TracerProvider); !isNoop {
		// Configure lib tracer so it does not remain a noop thus drop traces
		Tracer = existingProvider.Tracer(serviceName)

		// Do nothing, it is externally managed
		return func(_ context.Context) error {
			return nil
		}, nil
	}

	// Ensure default SDK resources and the required service name are set
	r, err := resource.Merge(
		resource.Environment(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, err
	}

	// Set up trace provider
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
