package main

import (
	"log/slog"
	"os"

	"github.com/MrBuryy/subscriptions-service/internal/app"
	"github.com/MrBuryy/subscriptions-service/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := app.Run(cfg, logger); err != nil {
		logger.Error("application stopped", slog.Any("error", err))
		os.Exit(1)
	}
}