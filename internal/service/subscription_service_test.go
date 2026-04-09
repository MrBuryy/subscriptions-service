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
	createFn func(ctx context.Context, sub *model.Subscription) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	listFn   func(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error)
	updateFn func(ctx context.Context, sub *model.Subscription) error
	deleteFn func(ctx context.Context, id uuid.UUID) error
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
	if f.listFn != nil {
		return f.listFn(ctx, filter)
	}
	return nil, nil
}

func (f *fakeSubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, sub)
	}
	return nil
}

func (f *fakeSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
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

func TestSubscriptionService_List_SuccessWithoutFilters(t *testing.T) {
	repo := &fakeSubscriptionRepository{
		listFn: func(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error) {
			if filter.UserID != nil {
				t.Fatalf("expected nil UserID filter")
			}
			if filter.ServiceName != nil {
				t.Fatalf("expected nil ServiceName filter")
			}
			if filter.Limit != 20 {
				t.Fatalf("expected limit 20, got %d", filter.Limit)
			}
			if filter.Offset != 0 {
				t.Fatalf("expected offset 0, got %d", filter.Offset)
			}

			return []model.Subscription{
				{ServiceName: "Netflix"},
				{ServiceName: "Spotify"},
			}, nil
		},
	}

	svc := NewSubscriptionService(repo)

	got, err := svc.List(context.Background(), nil, nil, 20, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(got))
	}
}

func TestSubscriptionService_List_SuccessWithUserIDFilter(t *testing.T) {
	userID := uuid.New()

	repo := &fakeSubscriptionRepository{
		listFn: func(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error) {
			if filter.UserID == nil {
				t.Fatalf("expected UserID filter")
			}
			if *filter.UserID != userID {
				t.Fatalf("expected userID %v, got %v", userID, *filter.UserID)
			}
			return []model.Subscription{}, nil
		},
	}

	svc := NewSubscriptionService(repo)

	userIDStr := "  " + userID.String() + "  "
	_, err := svc.List(context.Background(), &userIDStr, nil, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSubscriptionService_List_InvalidUserID(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	userID := "bad-uuid"

	_, err := svc.List(context.Background(), &userID, nil, 10, 0)
	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestSubscriptionService_List_InvalidEmptyServiceNameFilter(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	serviceName := "   "

	_, err := svc.List(context.Background(), nil, &serviceName, 10, 0)
	if !errors.Is(err, ErrInvalidService) {
		t.Fatalf("expected ErrInvalidService, got %v", err)
	}
}

func TestSubscriptionService_List_DefaultLimit(t *testing.T) {
	repo := &fakeSubscriptionRepository{
		listFn: func(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error) {
			if filter.Limit != 10 {
				t.Fatalf("expected default limit 10, got %d", filter.Limit)
			}
			return []model.Subscription{}, nil
		},
	}

	svc := NewSubscriptionService(repo)

	_, err := svc.List(context.Background(), nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSubscriptionService_List_NegativeOffset(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	_, err := svc.List(context.Background(), nil, nil, 10, -1)
	if !errors.Is(err, ErrInvalidOffset) {
		t.Fatalf("expected ErrInvalidOffset, got %v", err)
	}
}

func TestSubscriptionService_Update_Success(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()

	repo := &fakeSubscriptionRepository{
		updateFn: func(ctx context.Context, sub *model.Subscription) error {
			if sub.ID != id {
				t.Fatalf("expected id %v, got %v", id, sub.ID)
			}
			if sub.ServiceName != "Netflix" {
				t.Fatalf("expected service name Netflix, got %s", sub.ServiceName)
			}
			if sub.Price != 999 {
				t.Fatalf("expected price 999, got %d", sub.Price)
			}
			if sub.UserID != userID {
				t.Fatalf("expected userID %v, got %v", userID, sub.UserID)
			}
			if sub.StartDate.Year() != 2025 || sub.StartDate.Month() != time.May || sub.StartDate.Day() != 1 {
				t.Fatalf("unexpected start date: %v", sub.StartDate)
			}
			if sub.EndDate == nil {
				t.Fatal("expected end date, got nil")
			}
			if sub.EndDate.Year() != 2025 || sub.EndDate.Month() != time.December || sub.EndDate.Day() != 1 {
				t.Fatalf("unexpected end date: %v", *sub.EndDate)
			}

			sub.UpdatedAt = time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC)
			return nil
		},
	}

	svc := NewSubscriptionService(repo)

	endDate := "12-2025"

	got, err := svc.Update(
		context.Background(),
		id.String(),
		"Netflix",
		999,
		userID.String(),
		"05-2025",
		&endDate,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ID != id {
		t.Fatalf("expected id %v, got %v", id, got.ID)
	}
	if got.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}
}

func TestSubscriptionService_Update_InvalidID(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	_, err := svc.Update(context.Background(), "bad-id", "Netflix", 999, uuid.New().String(), "05-2025", nil)
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestSubscriptionService_Update_InvalidUserID(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	_, err := svc.Update(context.Background(), uuid.New().String(), "Netflix", 999, "bad-user-id", "05-2025", nil)
	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestSubscriptionService_Update_InvalidStartDate(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	_, err := svc.Update(context.Background(), uuid.New().String(), "Netflix", 999, uuid.New().String(), "bad-date", nil)
	if !errors.Is(err, ErrInvalidStartDate) {
		t.Fatalf("expected ErrInvalidStartDate, got %v", err)
	}
}

func TestSubscriptionService_Update_InvalidEndDate(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	endDate := "bad-date"

	_, err := svc.Update(context.Background(), uuid.New().String(), "Netflix", 999, uuid.New().String(), "05-2025", &endDate)
	if !errors.Is(err, ErrInvalidEndDate) {
		t.Fatalf("expected ErrInvalidEndDate, got %v", err)
	}
}

func TestSubscriptionService_Update_EndBeforeStart(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	endDate := "04-2025"

	_, err := svc.Update(context.Background(), uuid.New().String(), "Netflix", 999, uuid.New().String(), "05-2025", &endDate)
	if !errors.Is(err, ErrInvalidEndDate) {
		t.Fatalf("expected ErrInvalidEndDate, got %v", err)
	}
}

func TestSubscriptionService_Update_NotFound(t *testing.T) {
	repo := &fakeSubscriptionRepository{
		updateFn: func(ctx context.Context, sub *model.Subscription) error {
			return repository.ErrNotFound
		},
	}

	svc := NewSubscriptionService(repo)

	_, err := svc.Update(context.Background(), uuid.New().String(), "Netflix", 999, uuid.New().String(), "05-2025", nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSubscriptionService_Delete_Success(t *testing.T) {
	id := uuid.New()

	repo := &fakeSubscriptionRepository{
		deleteFn: func(ctx context.Context, gotID uuid.UUID) error {
			if gotID != id {
				t.Fatalf("expected id %v, got %v", id, gotID)
			}
			return nil
		},
	}

	svc := NewSubscriptionService(repo)

	err := svc.Delete(context.Background(), id.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSubscriptionService_Delete_InvalidID(t *testing.T) {
	repo := &fakeSubscriptionRepository{}
	svc := NewSubscriptionService(repo)

	err := svc.Delete(context.Background(), "bad-id")
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestSubscriptionService_Delete_NotFound(t *testing.T) {
	repo := &fakeSubscriptionRepository{
		deleteFn: func(ctx context.Context, gotID uuid.UUID) error {
			return repository.ErrNotFound
		},
	}

	svc := NewSubscriptionService(repo)

	err := svc.Delete(context.Background(), uuid.New().String())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}