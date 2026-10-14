package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"git-sync/internal/config"
	"git-sync/internal/httpserver"
	"git-sync/internal/snapshot"
	"git-sync/internal/syncer"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Getenv, os.Stderr)
	stop()
	os.Exit(code)
}
func run(parent context.Context, args []string, env func(string) string, out io.Writer) int {
	cfg, err := config.Parse(args, env, out)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(out, err)
		return 2
	}
	logger := slog.New(slog.NewJSONHandler(out, nil))
	repo, err := snapshot.New(cfg.Repository)
	if err != nil {
		logger.Error("initialize snapshot provider", "error", err)
		return 1
	}
	defer repo.Close()
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	service, err := syncer.New(repo, cfg.Interval, cfg.Timeout, logger, registry)
	if err != nil {
		logger.Error("initialize worker", "error", err)
		return 1
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	var listener net.Listener
	if cfg.HTTPAddr != "" {
		listener, err = net.Listen("tcp", cfg.HTTPAddr)
		if err != nil {
			logger.Error("bind HTTP server", "error", err)
			return 1
		}
	}
	workerDone := make(chan error, 1)
	go func() { workerDone <- service.Run(ctx) }()
	if listener == nil {
		if err = <-workerDone; err != nil {
			logger.Error("worker stopped", "error", err)
			return 1
		}
		return 0
	}
	httpDone := make(chan error, 1)
	go func() { httpDone <- httpserver.Serve(ctx, listener, httpserver.Handler(cfg, service, registry)) }()
	select {
	case err = <-httpDone:
		cancel()
		workerErr := <-workerDone
		if err == nil {
			err = workerErr
		}
	case err = <-workerDone:
		cancel()
		httpErr := <-httpDone
		if err == nil {
			err = httpErr
		}
	}
	if err != nil {
		logger.Error("service stopped", "error", err)
		return 1
	}
	return 0
}
