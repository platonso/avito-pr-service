package main

import (
	"context"
	"fmt"
	"github.com/platonso/avito-pr-service/internal/app"
	"github.com/platonso/avito-pr-service/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := app.New(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	errChan := make(chan error, 1)
	go func() {
		if err := a.Run(); err != nil {
			errChan <- err
		}
	}()

	logger.Info("application started", slog.String("port", cfg.HTTPPort))

	select {
	case sig := <-sigChan:
		logger.Info("received signal", slog.String("signal", sig.String()))
		cancel()
	case err := <-errChan:
		logger.Error("server error", slog.Any("error", err))
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown gracefully: %w", err)
	}

	logger.Info("application stopped gracefully")
	return nil
}
