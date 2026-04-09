package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MrBuryy/subscriptions-service/internal/model"
	"github.com/MrBuryy/subscriptions-service/internal/repository"
)

type fakeSubscriptionRepository struct {
	createFn  func(ctx context.Context, sub *model.Subscription) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
}

func (f *fakeSubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	if f.createFn != nil {
		return f.createFn(ctx, sub)
	}
	return nil
}

func (f *fakeSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeSubscriptionRepository) List(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error) {
	return nil, nil
}

func (f *fakeSubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	return nil
}

func (f *fakeSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestSubscriptionService_Create(t *testing.T) {
	validUserID := uuid.New()
	validEndDate := "03-2025"

	tests := []struct {
		name        string
		serviceName string
		price       int
		userID      string
		startDate   string
		endDate     *string
		repo        *fakeSubscriptionRepository
		wantErr     error
		check       func(t *testing.T, got *model.Subscription)
	}{
		{
			name:        "success",
			serviceName: "Netflix",
			price:       999,
			userID:      validUserID.String(),
			startDate:   "01-2025",
			endDate:     &validEndDate,
			repo: &fakeSubscriptionRepository{
				createFn: func(ctx context.Context, sub *model.Subscription) error {
					sub.ID = uuid.New()
					sub.CreatedAt = time.Now()
					sub.UpdatedAt = time.Now()
					return nil
				},
			},
			check: func(t *testing.T, got *model.Subscription) {
				if got == nil {
					t.Fatal("expected subscription, got nil")
				}
				if got.ServiceName != "Netflix" {
					t.Fatalf("expected service name Netflix, got %q", got.ServiceName)
				}
				if got.Price != 999 {
					t.Fatalf("expected price 999, got %d", got.Price)
				}
				if got.UserID != validUserID {
					t.Fatalf("expected user id %s, got %s", validUserID, got.UserID)
				}
				if got.StartDate.Format("2006-01-02") != "2025-01-01" {
					t.Fatalf("expected start date 2025-01-01, got %s", got.StartDate.Format("2006-01-02"))
				}
				if got.EndDate == nil {
					t.Fatal("expected end date, got nil")
				}
				if got.EndDate.Format("2006-01-02") != "2025-03-01" {
					t.Fatalf("expected end date 2025-03-01, got %s", got.EndDate.Format("2006-01-02"))
				}
			},
		},
		{
			name:        "empty service name",
			serviceName: "   ",
			price:       999,
			userID:      validUserID.String(),
			startDate:   "01-2025",
			endDate:     nil,
			repo:        &fakeSubscriptionRepository{},
			wantErr:     ErrInvalidService,
		},
		{
			name:        "price less than or equal zero",
			serviceName: "Netflix",
			price:       0,
			userID:      validUserID.String(),
			startDate:   "01-2025",
			endDate:     nil,
			repo:        &fakeSubscriptionRepository{},
			wantErr:     ErrInvalidPrice,
		},
		{
			name:        "invalid user id",
			serviceName: "Netflix",
			price:       999,
			userID:      "bad-uuid",
			startDate:   "01-2025",
			endDate:     nil,
			repo:        &fakeSubscriptionRepository{},
			wantErr:     ErrInvalidUserID,
		},
		{
			name:        "invalid start date",
			serviceName: "Netflix",
			price:       999,
			userID:      validUserID.String(),
			startDate:   "2025-01",
			endDate:     nil,
			repo:        &fakeSubscriptionRepository{},
			wantErr:     ErrInvalidStartDate,
		},
		{
			name:        "end date before start date",
			serviceName: "Netflix",
			price:       999,
			userID:      validUserID.String(),
			startDate:   "03-2025",
			endDate:     strPtr("02-2025"),
			repo:        &fakeSubscriptionRepository{},
			wantErr:     ErrInvalidEndDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSubscriptionService(tt.repo)

			got, err := svc.Create(
				context.Background(),
				tt.serviceName,
				tt.price,
				tt.userID,
				tt.startDate,
				tt.endDate,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				if got != nil {
					t.Fatalf("expected nil subscription, got %+v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestSubscriptionService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      string
		repo    *fakeSubscriptionRepository
		wantErr error
		check   func(t *testing.T, got *model.Subscription)
	}{
		{
			name: "success",
			id:   existingID.String(),
			repo: &fakeSubscriptionRepository{
				getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
					return &model.Subscription{
						ID:          id,
						ServiceName: "Netflix",
						Price:       999,
						UserID:      uuid.New(),
						StartDate:   time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
					}, nil
				},
			},
			check: func(t *testing.T, got *model.Subscription) {
				if got == nil {
					t.Fatal("expected subscription, got nil")
				}
				if got.ID != existingID {
					t.Fatalf("expected id %s, got %s", existingID, got.ID)
				}
				if got.ServiceName != "Netflix" {
					t.Fatalf("expected service name Netflix, got %q", got.ServiceName)
				}
			},
		},
		{
			name:    "invalid id",
			id:      "bad-uuid",
			repo:    &fakeSubscriptionRepository{},
			wantErr: ErrInvalidID,
		},
		{
			name: "not found",
			id:   existingID.String(),
			repo: &fakeSubscriptionRepository{
				getByIDFn: func(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
					return nil, repository.ErrNotFound
				},
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSubscriptionService(tt.repo)

			got, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				if got != nil {
					t.Fatalf("expected nil subscription, got %+v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}