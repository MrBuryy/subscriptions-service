package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/MrBuryy/subscriptions-service/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5433/subscriptions?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func truncateSubscriptions(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		TRUNCATE TABLE subscriptions
		RESTART IDENTITY
		CASCADE
	`)
	if err != nil {
		t.Fatalf("failed to truncate subscriptions: %v", err)
	}
}

func cleanupSubscriptionsTable(t *testing.T, db *sql.DB) {
	t.Helper()

	truncateSubscriptions(t, db)
	t.Cleanup(func() {
		truncateSubscriptions(t, db)
	})
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func assertEndDateEqual(t *testing.T, got, want *time.Time) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Fatalf("expected EndDate to be nil, got %v", *got)
		}
		return
	}

	if got == nil {
		t.Fatal("expected EndDate to be set")
	}

	if !got.Equal(*want) {
		t.Fatalf("unexpected EndDate: got %v want %v", *got, *want)
	}
}

func createSubscriptionWithData(
	t *testing.T,
	repo *SubscriptionRepository,
	data model.Subscription,
) *model.Subscription {
	t.Helper()

	sub := &model.Subscription{
		ServiceName: data.ServiceName,
		Price:       data.Price,
		UserID:      data.UserID,
		StartDate:   data.StartDate,
		EndDate:     data.EndDate,
	}

	if err := repo.Create(context.Background(), sub); err != nil {
		t.Fatalf("failed to create test subscription: %v", err)
	}

	return sub
}

func assertSubscriptionEqual(t *testing.T, got, want model.Subscription) {
	t.Helper()

	if got.ID != want.ID {
		t.Fatalf("unexpected ID: got %v want %v", got.ID, want.ID)
	}

	if got.ServiceName != want.ServiceName {
		t.Fatalf("unexpected ServiceName: got %q want %q", got.ServiceName, want.ServiceName)
	}

	if got.Price != want.Price {
		t.Fatalf("unexpected Price: got %v want %v", got.Price, want.Price)
	}

	if got.UserID != want.UserID {
		t.Fatalf("unexpected UserID: got %v want %v", got.UserID, want.UserID)
	}

	if !got.StartDate.Equal(want.StartDate) {
		t.Fatalf("unexpected StartDate: got %v want %v", got.StartDate, want.StartDate)
	}

	assertEndDateEqual(t, got.EndDate, want.EndDate)

	if got.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}

	if got.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}
}