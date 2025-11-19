package main

import (
	"context"
	"fmt"
	"github.com/platonso/avito-pr-service/internal/app"
	"github.com/platonso/avito-pr-service/internal/config"
	"log/slog"
	"os"
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

	a, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("failed to create app", slog.Any("error", err))
		os.Exit(1)
	}

	// TODO in goroutine
	if err = a.Run(); err != nil {
		logger.Error("failed to run server", slog.Any("error", err))
	}
}
