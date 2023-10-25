package main

import (
	"context"
	"errors"
	//"time"

	//"go.opentelemetry.io/otel"
	//"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	//"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	//"go.opentelemetry.io/otel/propagation"
	//"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	//"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func setupOTelSDK(
	ctx context.Context,
	serviceName, serviceVersion string,
) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// setup resource
	_, err = newResource(serviceName, serviceVersion)
	if err != nil {
		handleErr(err)
		return
	}

	return
}

func newResource(
	serviceName, serviceVersion string,
) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
}
