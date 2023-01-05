package main

import (
	"context"
	"log"
	"os"

	"github.com/schemahero/schemahero/pkg/cli/managercli"
	"github.com/schemahero/schemahero/pkg/trace"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	l := log.New(os.Stdout, "", 0)

	// Write telemetry data to a file.
	f, err := os.Create("traces.txt")
	if err != nil {
		l.Fatal(err)
	}
	defer f.Close()

	exp, err := trace.NewExporter(f)
	if err != nil {
		l.Fatal(err)
	}

	tp := oteltrace.NewTracerProvider(oteltrace.WithBatcher(exp), oteltrace.WithResource(trace.NewResource()))
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			l.Fatal(err)
		}
	}()
	otel.SetTracerProvider(tp)

	managercli.InitAndExecute()
}
