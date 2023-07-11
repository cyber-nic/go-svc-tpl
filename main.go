// go svc tpl
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	exitCodeSuccess   = 0
	exitCodeErr       = 1
	exitCodeInterrupt = 2
)

var (
	svcOpts *SvcOptions
	debug   *bool
)

func main() {
	parseCLIArgs()

	// context and signal handling
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	// logger
	ctx = contextWithLogger(ctx, newLogger())
	l := loggerFromContext(ctx)
	level.Info(l).Log("msg", "svc start", "debug", debug)

	// interrupt handling
	done := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	}()

	// main service
	svc := NewFooBarSvc(ctx, l, svcOpts)
	go func() {
		if err := svc.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(exitCodeErr)
		}
		os.Exit(exitCodeSuccess)
	}()

	// allow context cancelling
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
			svc.Stop()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		os.Exit(exitCodeInterrupt)
	}()

	// metrics and health
	startWebServer(svc, l, done)
	level.Info(l).Log("exit", <-done)
}

func parseCLIArgs() {
	svcOpts = &SvcOptions{
		FooOption: flag.String("option", "one", "options: one|two|three"),
	}
	debug = flag.Bool("debug", false, "Debug logging level")

	flag.Parse()
}

func startWebServer(svc *FooBar, l log.Logger, exit chan error) {
	go func() {
		port := ":8080"
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			if svc.IsReady() {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ready"))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("not ready"))
		})
		level.Info(l).Log("msg", fmt.Sprintf("Serving '/metrics' on port %s", port))
		level.Info(l).Log("msg", fmt.Sprintf("Serving '/health' on port %s", port))

		server := &http.Server{
			Addr:              port,
			ReadHeaderTimeout: 30 * time.Second,
		}
		exit <- server.ListenAndServe()
	}()
}
