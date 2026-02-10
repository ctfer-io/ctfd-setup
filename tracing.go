package ctfdsetup

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	serviceName = "ctfd-setup"
)

type OTelSetup struct {
	Shutdown       func(context.Context) error
	TracerProvider trace.TracerProvider
	LogProvider    log.LoggerProvider
}

func SetupOTelSDK(ctx context.Context, version string) (*OTelSetup, error) {
	// Set up propagator
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

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

	// Then create the span exporter
	texp, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, err
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(texp)),
		sdktrace.WithResource(r),
	)

	// And the log exporter
	lexp, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return nil, err
	}
	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(lexp)),
		sdklog.WithResource(r),
	)

	// And we are ready to go!
	return &OTelSetup{
		Shutdown: func(ctx context.Context) error {
			return multierr.Combine(
				tracerProvider.Shutdown(context.WithoutCancel(ctx)),
				logProvider.Shutdown(context.WithoutCancel(ctx)),
			)
		},
		TracerProvider: tracerProvider,
		LogProvider:    logProvider,
	}, nil
}

func StartAPISpan(ctx context.Context, tracer trace.Tracer) (context.Context, trace.Span) {
	method := getCallerFunctionName()

	return tracer.Start(
		ctx,
		fmt.Sprintf("api/%s", method),
	)
}

func LogAPICall(ctx context.Context) {
	Log().Debug(ctx, "api call", zap.String("method", getCallerFunctionName()))
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
