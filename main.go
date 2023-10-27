package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	// start HTTP server
	srv := &http.Server{
		Addr: ":8080",
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		ReadTimeout: time.Second,
		WriteTimeout: 10 * time.Second,
		Handler: newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// main loop
	select {
	case err = <-srvErr:
		return
	case <-ctx.Done():
		// waiting for first Ctrl+C.
		// stop receiving signal notifications as soon as possible
		stop()
	}

	// When Shutdown is called, ListenAndServ immediately returns
	// ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern
	// as the HTTP route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// register the handlers
	handleFunc("/rolldice", rolldice)

	// add HTTP instrumentation for the whole server.
	return otelhttp.NewHandler(mux, "/")
}
