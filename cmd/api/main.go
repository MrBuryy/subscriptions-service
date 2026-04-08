package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MrBuryy/subscriptions-service/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.PostgresDSN())
	if err != nil {
		logger.Error("failed to open postgres connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("failed to ping postgres", "error", err)
		os.Exit(1)
	}

	logger.Info("postgres connected",
		"app_env", cfg.AppEnv,
		"http_addr", cfg.HTTPAddr,
		"postgres_host", cfg.PostgresHost,
		"postgres_db", cfg.PostgresDB,
	)
}