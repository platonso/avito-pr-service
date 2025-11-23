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
	fmt.Println("start")

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)

	cfg, err := config.New()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := app.New(ctx, cfg, logger)
	if err != nil {
		logger.Error("failed to create app", slog.Any("error", err))
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Run server
	errChan := make(chan error, 1)
	go func() {
		if err := a.Run(); err != nil {
			errChan <- err
		}
	}()

	// Catch the signal of completion or starting error
	select {
	case sig := <-sigChan:
		logger.Info("received signal", slog.String("signal", sig.String()))
		cancel()
	case err := <-errChan:
		logger.Error("server error", slog.Any("error", err))
		cancel()
	}

	// Execute graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown gracefully", slog.Any("error", err))
		os.Exit(1)
	}
}
