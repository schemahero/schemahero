package trace

import (
	"context"
	"io"

	"github.com/schemahero/schemahero/pkg/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc/credentials"
)

const (
	TraceName = "schemahero"
)

func NewExporter(w io.Writer) (oteltrace.SpanExporter, error) {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	secureOption = otlptracegrpc.WithInsecure()

	return otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(secureOption, otlptracegrpc.WithEndpoint("localhost:4317")),
	)

	// return stdouttrace.New(
	// 	stdouttrace.WithWriter(w),
	// 	stdouttrace.WithPrettyPrint(),
	// 	// stdouttrace.WithoutTimestamps(),
	// )
}

func NewResource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(TraceName),
			semconv.ServiceVersionKey.String(version.Version()),
			attribute.String("environment", "dev"),
		),
	)

	return r
}
