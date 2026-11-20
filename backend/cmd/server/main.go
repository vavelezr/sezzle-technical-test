// Command server runs the calculator REST API.
//
// This file is the composition root: it reads configuration and wires the
// dependencies together. It contains no business logic.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sezzle-calculator/backend/internal/config"
	"sezzle-calculator/backend/internal/operations"
	"sezzle-calculator/backend/internal/registry"
	"sezzle-calculator/backend/internal/service"
	httptransport "sezzle-calculator/backend/internal/transport/http"
)

func main() {
	if err := run(); err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("server exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ops, err := registry.New(operations.All()...)
	if err != nil {
		return err
	}
	calculator := service.NewCalculatorService(ops)

	handler := httptransport.NewHandler(calculator, logger, cfg.MaxBodyBytes)
	api := httptransport.Chain(
		httptransport.NewRouter(handler),
		httptransport.RequestID(),
		httptransport.Logging(logger),
		httptransport.Recover(logger),
		httptransport.CORS(cfg.AllowedOrigins),
		httptransport.Timeout(cfg.RequestTimeout),
	)

	server := httptransport.NewServer(api,
		httptransport.WithAddr(cfg.Addr()),
		httptransport.WithReadTimeout(cfg.ReadTimeout),
		httptransport.WithWriteTimeout(cfg.WriteTimeout),
		httptransport.WithIdleTimeout(cfg.IdleTimeout),
		httptransport.WithShutdownTimeout(cfg.ShutdownTimeout),
		httptransport.WithLogger(logger),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return server.Run(ctx)
}
