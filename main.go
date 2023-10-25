package main

import (
	"context"
	"errors"
	"log"
	//"net/http"
	"os"
	"os/signal"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// handle SIGINT (ctrl+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// setup OpenTelemetry
	serviceName := "dice"
	serviceVersion := "0.1.0"
	otelShutdown, err := setupOTelSDK(ctx, serviceName, serviceVersion)
	if err != nil {
		return
	}
	// handle shutdown properly so nothing leaks
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// main loop
	select {
	case <-ctx.Done():
		// waiting for first Ctrl+C.
		// stop receiving signal notifications as soon as possible
		stop()
		return
	}
}
