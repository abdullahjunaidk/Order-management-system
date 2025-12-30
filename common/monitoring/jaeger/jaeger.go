package jaeger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// JaegerTracer struct.
// This struct is used to represent a Jaeger tracer.
//
// Attributes:
// JaegerTracer struct.
// This struct is used to represent a Jaeger tracer.
//
// Attributes:
//   - shutdown func(): The shutdown function to flush traces.
type JaegerTracer struct {
	shutdown func()
}

// NewJaegerTracer function.
// This function is used to create a new Jaeger tracer.
//
// Parameters:
//   - serviceName (string): The service name.
//   - otlpEndpoint (string): The OTLP endpoint.
//   - protocol (string): The OTLP protocol (http or grpc).
//   - insecure (string): The OTLP insecure.
//
// Returns:
//   - *JaegerTracer: The Jaeger tracer.
//   - error: The error.
func NewJaegerTracer(serviceName, otlpEndpoint, protocol, insecure string) (*JaegerTracer, error) {
	var shutdown func()
	var err error

	switch protocol {
	case "http":
		shutdown, err = setupOTLPHTTPTracing(serviceName, otlpEndpoint, insecure)
	case "grpc":
		shutdown, err = setupOTLPGRPCTracing(serviceName, otlpEndpoint, insecure)
	default:
		return nil, fmt.Errorf("invalid OTLP protocol: %s, use 'http' or 'grpc'", protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to setup OTLP tracing: %w", err)
	}

	return &JaegerTracer{shutdown: shutdown}, nil
}

// Close function.
// This function is used to shutdown the Jaeger tracer and flush traces.
func (jt *JaegerTracer) Close() {
	jt.shutdown()
}

// setupOTLPHTTPTracing function.
// This function is used to set up OTLP HTTP tracing.
//
// Parameters:
//   - serviceName (string): The service name.
//   - otlpEndpoint (string): The OTLP endpoint.
//   - insecure (string): The OTLP insecure.
//
// Returns:
//   - func(): The shutdown function to flush traces.
//   - error: The error.
func setupOTLPHTTPTracing(serviceName, otlpEndpoint, insecure string) (func(), error) {
	// Create the OTLP HTTP exporter
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(otlpEndpoint),
	}

	if insecure == "true" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exp, err := otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP HTTP exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(newResource(serviceName)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}, nil
}

// setupOTLPGRPCTracing function.
// This function is used to set up OTLP gRPC tracing.
//
// Parameters:
//   - serviceName (string): The service name.
//   - otlpEndpoint (string): The OTLP endpoint.
//   - insecure (string): The OTLP insecure.
//
// Returns:
//   - func(): The shutdown function to flush traces.
//   - error: The error.
func setupOTLPGRPCTracing(serviceName, otlpEndpoint, insecure string) (func(), error) {
	// Create the OTLP gRPC exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(otlpEndpoint),
	}

	if insecure == "true" {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exp, err := otlptracegrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP gRPC exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(newResource(serviceName)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}, nil
}

// newResource function.
// This function is used to create a new resource.
//
// Parameters:
//   - serviceName (string): The service name.
//
// Returns:
//   - *resource.Resource: The resource.
func newResource(serviceName string) *resource.Resource {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	extraAttributes := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion("0.1.0"),
		semconv.HostName(hostname),
	}

	r, err := resource.New(
		context.Background(),
		resource.WithAttributes(extraAttributes...),
	)
	if err != nil {
		log.Printf("Could not set resources: %v", err)
	}
	return r
}

// InitGlobalTracer function.
// This function is used to initialize the global tracer.
//
// Parameters:
//   - serviceName (string): The service name.
//   - otlpEndpoint (string): The OTLP endpoint.
//   - protocol (string): The OTLP protocol (http or grpc).
//
// Returns:
//   - io.Closer: The closer to shutdown the tracer.
//   - error: The error.
func InitGlobalTracer(serviceName string, otlpEndpoint string, protocol string) (io.Closer, error) {
	var shutdown func()
	var err error

	switch protocol {
	case "http":
		shutdown, err = setupOTLPHTTPTracing(serviceName, otlpEndpoint, "true")
	case "grpc":
		shutdown, err = setupOTLPGRPCTracing(serviceName, otlpEndpoint, "true")
	default:
		return nil, fmt.Errorf("invalid OTLP protocol: %s, use 'http' or 'grpc'", protocol)
	}

	if err != nil {
		return nil, err
	}

	return closerFunc(shutdown), nil
}

// closerFunc type.
// This type is used to define a closer function.
type closerFunc func()

// Close method.
// This method is used to close the closer function.
func (c closerFunc) Close() error {
	c()
	return nil
}
