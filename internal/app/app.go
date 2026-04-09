package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/MrBuryy/subscriptions-service/docs"
	"github.com/MrBuryy/subscriptions-service/internal/config"
	"github.com/MrBuryy/subscriptions-service/internal/handler"
	"github.com/MrBuryy/subscriptions-service/internal/repository/postgres"
	"github.com/MrBuryy/subscriptions-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func Run(cfg *config.Config, logger *slog.Logger) error {
	logger.Info("config loaded",
		slog.String("app_env", cfg.AppEnv),
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("postgres_host", cfg.PostgresHost),
		slog.String("postgres_port", cfg.PostgresPort),
		slog.String("postgres_db", cfg.PostgresDB),
	)

	db, err := newDB(cfg)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	logger.Info("database connected")

	subRepo := postgres.NewSubscriptionRepository(db)
	subService := service.NewSubscriptionService(subRepo)
	subHandler := handler.NewSubscriptionHandler(subService)

	router := newRouter(subHandler)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrCh := make(chan error, 1)

	go func() {
		logger.Info("http server started", slog.String("addr", cfg.HTTPAddr))

		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}

		serverErrCh <- nil
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stopCh:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-serverErrCh:
		if err != nil {
			_ = db.Close()
			return fmt.Errorf("http server failed: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = db.Close()
		return fmt.Errorf("shutdown server: %w", err)
	}

	logger.Info("http server stopped")

	if err := db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}

	logger.Info("database connection closed")

	return nil
}

func newDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.PostgresDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func newRouter(subHandler *handler.SubscriptionHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))


	r.Get("/health", handler.Health)

	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/summary", subHandler.Summary)
		r.Post("/", subHandler.Create)
		r.Get("/", subHandler.List)
		r.Get("/{id}", subHandler.GetByID)
		r.Put("/{id}", subHandler.Update)
		r.Delete("/{id}", subHandler.Delete)
	})

	return r
}