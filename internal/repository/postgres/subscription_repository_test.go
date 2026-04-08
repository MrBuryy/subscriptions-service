package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MrBuryy/subscriptions-service/internal/model"
	"github.com/MrBuryy/subscriptions-service/internal/repository"
)

func TestSubscriptionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepository(db)

	tests := []struct {
		name    string
		sub     model.Subscription
		wantErr bool
	}{
		{
			name: "with end date",
			sub: model.Subscription{
				ServiceName: "Spotify",
				Price:       200,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.May, 1),
				EndDate:     timePtr(date(2025, time.June, 1)),
			},
		},
		{
			name: "without end date",
			sub: model.Subscription{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.January, 1),
				EndDate:     nil,
			},
		},
		{
			name: "price is zero",
			sub: model.Subscription{
				ServiceName: "Apple Music",
				Price:       0,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.July, 15),
				EndDate:     nil,
			},
			wantErr: true,
		},
		{
			name: "price is negative",
			sub: model.Subscription{
				ServiceName: "YouTube Premium",
				Price:       -100,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.March, 10),
				EndDate:     timePtr(date(2030, time.March, 10)),
			},
			wantErr: true,
		},
		{
			name: "end date equal start date",
			sub: model.Subscription{
				ServiceName: "Kinopoisk",
				Price:       300,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.June, 1),
			EndDate:     timePtr(date(2025, time.June, 1)),
			},
			wantErr: false,
		},
		{
			name: "end date before start date",
			sub: model.Subscription{
				ServiceName: "Kinopoisk",
				Price:       300,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.June, 10),
				EndDate:     timePtr(date(2025, time.May, 1)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupSubscriptionsTable(t, db)

			sub := tt.sub

			err := repo.Create(context.Background(), &sub)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Create returned error: %v", err)
			}

			if sub.ID == uuid.Nil {
				t.Fatal("expected ID to be set")
			}

			if sub.CreatedAt.IsZero() {
				t.Fatal("expected CreatedAt to be set")
			}

			if sub.UpdatedAt.IsZero() {
				t.Fatal("expected UpdatedAt to be set")
			}

			got, err := repo.GetByID(context.Background(), sub.ID)
			if err != nil {
				t.Fatalf("failed to fetch created subscription: %v", err)
			}

			if got.ServiceName != sub.ServiceName {
				t.Fatalf("unexpected ServiceName: got %q want %q", got.ServiceName, sub.ServiceName)
			}

			if got.Price != sub.Price {
				t.Fatalf("unexpected Price: got %d want %d", got.Price, sub.Price)
			}

			if got.UserID != sub.UserID {
				t.Fatalf("unexpected UserID: got %s want %s", got.UserID, sub.UserID)
			}

			if !got.StartDate.Equal(sub.StartDate) {
				t.Fatalf("unexpected StartDate: got %v want %v", got.StartDate, sub.StartDate)
			}

			assertEndDateEqual(t, got.EndDate, sub.EndDate)
		})
	}
}

func TestSubscriptionRepository_GetByID_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepository(db)

	tests := []struct {
		name string
		sub  model.Subscription
	}{
		{
			name: "with end date",
			sub: model.Subscription{
				ServiceName: "Spotify",
				Price:       200,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.May, 1),
				EndDate:     timePtr(date(2025, time.June, 1)),
			},
		},
		{
			name: "without end date",
			sub: model.Subscription{
				ServiceName: "Netflix",
				Price:       999,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.January, 1),
				EndDate:     nil,
			},
		},
		{
			name: "end date equal start date",
			sub: model.Subscription{
				ServiceName: "Kinopoisk",
				Price:       300,
				UserID:      uuid.New(),
				StartDate:   date(2025, time.June, 1),
				EndDate:     timePtr(date(2025, time.June, 1)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupSubscriptionsTable(t, db)

			sub := tt.sub
			err := repo.Create(context.Background(), &sub)
			if err != nil {
				t.Fatalf("failed to create subscription: %v", err)
			}

			got, err := repo.GetByID(context.Background(), sub.ID)
			if err != nil {
				t.Fatalf("GetByID returned error: %v", err)
			}

			if got.ID != sub.ID {
				t.Fatalf("unexpected ID: got %s want %s", got.ID, sub.ID)
			}

			if got.ServiceName != sub.ServiceName {
				t.Fatalf("unexpected ServiceName: got %q want %q", got.ServiceName, sub.ServiceName)
			}

			if got.Price != sub.Price {
				t.Fatalf("unexpected Price: got %d want %d", got.Price, sub.Price)
			}

			if got.UserID != sub.UserID {
				t.Fatalf("unexpected UserID: got %s want %s", got.UserID, sub.UserID)
			}

			if !got.StartDate.Equal(sub.StartDate) {
				t.Fatalf("unexpected StartDate: got %v want %v", got.StartDate, sub.StartDate)
			}

			assertEndDateEqual(t, got.EndDate, sub.EndDate)

			if got.CreatedAt.IsZero() {
				t.Fatal("expected CreatedAt to be set")
			}

			if got.UpdatedAt.IsZero() {
				t.Fatal("expected UpdatedAt to be set")
			}
		})
	}
}

func TestSubscriptionRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	cleanupSubscriptionsTable(t, db)

	repo := NewSubscriptionRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSubscriptionRepository_List(t *testing.T) {
	db := setupTestDB(t)
	cleanupSubscriptionsTable(t, db)

	repo := NewSubscriptionRepository(db)

	serviceSpotify := "Spotify"

	userID1 := uuid.New()
	userID2 := uuid.New()

	tests := []struct {
		name         string
		filter       model.SubscriptionFilter
		prepare      func(t *testing.T) []model.Subscription
		wantLen      int
		assertResult func(t *testing.T, got []model.Subscription, prepared []model.Subscription)
	}{
		{
			name:   "returns empty slice when no subscriptions",
			filter: model.SubscriptionFilter{},
			prepare: func(t *testing.T) []model.Subscription {
				return nil
			},
			wantLen: 0,
		},
		{
			name: "filters by user id",
			filter: model.SubscriptionFilter{
				UserID: &userID1,
			},
			prepare: func(t *testing.T) []model.Subscription {
				sub1 := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       200,
					UserID:      userID1,
					StartDate:   date(2025, time.January, 1),
					EndDate:     timePtr(date(2025, time.February, 1)),
				})

				_ = createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Netflix",
					Price:       500,
					UserID:      userID2,
					StartDate:   date(2025, time.March, 1),
					EndDate:     nil,
				})

				return []model.Subscription{*sub1}
			},
			wantLen: 1,
			assertResult: func(t *testing.T, got []model.Subscription, prepared []model.Subscription) {
				assertSubscriptionEqual(t, got[0], prepared[0])
			},
		},
		{
			name: "filters by service name",
			filter: model.SubscriptionFilter{
				ServiceName: &serviceSpotify,
				Limit:       10,
				Offset:      0,
			},
			prepare: func(t *testing.T) []model.Subscription {
				sub1 := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       200,
					UserID:      userID1,
					StartDate:   date(2025, time.January, 1),
					EndDate:     nil,
				})
				time.Sleep(5 * time.Millisecond)

				sub2 := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       300,
					UserID:      userID2,
					StartDate:   date(2025, time.February, 1),
					EndDate:     timePtr(date(2025, time.March, 1)),
				})
				time.Sleep(5 * time.Millisecond)

				_ = createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Netflix",
					Price:       999,
					UserID:      uuid.New(),
					StartDate:   date(2025, time.April, 1),
					EndDate:     nil,
				})

				return []model.Subscription{*sub2, *sub1}
			},
			wantLen: 2,
			assertResult: func(t *testing.T, got []model.Subscription, prepared []model.Subscription) {
				assertSubscriptionEqual(t, got[0], prepared[0])
				assertSubscriptionEqual(t, got[1], prepared[1])
			},
		},
		{
			name: "filters by user id and service name together",
			filter: model.SubscriptionFilter{
				UserID:      &userID1,
				ServiceName: &serviceSpotify,
			},
			prepare: func(t *testing.T) []model.Subscription {
				target := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       200,
					UserID:      userID1,
					StartDate:   date(2025, time.January, 1),
					EndDate:     nil,
				})

				_ = createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Netflix",
					Price:       400,
					UserID:      userID1,
					StartDate:   date(2025, time.February, 1),
					EndDate:     nil,
				})

				_ = createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       600,
					UserID:      userID2,
					StartDate:   date(2025, time.March, 1),
					EndDate:     nil,
				})

				return []model.Subscription{*target}
			},
			wantLen: 1,
			assertResult: func(t *testing.T, got []model.Subscription, prepared []model.Subscription) {
				assertSubscriptionEqual(t, got[0], prepared[0])
			},
		},
		{
			name: "uses default limit 10 when limit is zero",
			filter: model.SubscriptionFilter{
				Limit:  0,
				Offset: 0,
			},
			prepare: func(t *testing.T) []model.Subscription {
				prepared := make([]model.Subscription, 0, 12)

				for i := 0; i < 12; i++ {
					sub := createSubscriptionWithData(t, repo, model.Subscription{
						ServiceName: "Service",
						Price:       100 + i,
						UserID:      userID1,
						StartDate:   date(2025, time.January, i+1),
						EndDate:     nil,
					})
					prepared = append(prepared, *sub)
					time.Sleep(5 * time.Millisecond)
				}

				return prepared
			},
			wantLen: 10,
		},
		{
			name: "applies offset",
			filter: model.SubscriptionFilter{
				Limit:  1,
				Offset: 1,
			},
			prepare: func(t *testing.T) []model.Subscription {
				sub1 := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "First",
					Price:       100,
					UserID:      userID1,
					StartDate:   date(2025, time.January, 1),
					EndDate:     nil,
				})
				time.Sleep(5 * time.Millisecond)

				sub2 := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Second",
					Price:       200,
					UserID:      userID1,
					StartDate:   date(2025, time.January, 2),
					EndDate:     nil,
				})
				time.Sleep(5 * time.Millisecond)

				sub3 := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Third",
					Price:       300,
					UserID:      userID1,
					StartDate:   date(2025, time.January, 3),
					EndDate:     nil,
				})

				return []model.Subscription{*sub3, *sub2, *sub1}
			},
			wantLen: 1,
			assertResult: func(t *testing.T, got []model.Subscription, prepared []model.Subscription) {
				assertSubscriptionEqual(t, got[0], prepared[1])
			},
		},
		{
			name:   "maps null end date to nil",
			filter: model.SubscriptionFilter{},
			prepare: func(t *testing.T) []model.Subscription {
				sub := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Netflix",
					Price:       500,
					UserID:      userID1,
					StartDate:   date(2025, time.January, 1),
					EndDate:     nil,
				})

				return []model.Subscription{*sub}
			},
			wantLen: 1,
			assertResult: func(t *testing.T, got []model.Subscription, prepared []model.Subscription) {
				assertSubscriptionEqual(t, got[0], prepared[0])

				if got[0].EndDate != nil {
					t.Fatalf("expected EndDate to be nil, got %v", *got[0].EndDate)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateSubscriptions(t, db)

			prepared := tt.prepare(t)

			got, err := repo.List(context.Background(), tt.filter)
			if err != nil {
				t.Fatalf("List returned error: %v", err)
			}

			if len(got) != tt.wantLen {
				t.Fatalf("expected %d subscriptions, got %d", tt.wantLen, len(got))
			}

			if tt.assertResult != nil {
				tt.assertResult(t, got, prepared)
			}
		})
	}
}
func TestSubscriptionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	cleanupSubscriptionsTable(t, db)

	repo := NewSubscriptionRepository(db)

	tests := []struct {
		name        string
		prepare     func(t *testing.T) *model.Subscription
		update      func(sub *model.Subscription)
		wantErr     error
		assertAfter func(t *testing.T, updated *model.Subscription)
	}{
		{
			name: "updates subscription successfully",
			prepare: func(t *testing.T) *model.Subscription {
				return createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       200,
					UserID:      uuid.New(),
					StartDate:   date(2025, time.January, 1),
					EndDate:     timePtr(date(2025, time.February, 1)),
				})
			},
			update: func(sub *model.Subscription) {
				sub.ServiceName = "Netflix"
				sub.Price = 999
				sub.UserID = uuid.New()
				sub.StartDate = date(2025, time.March, 1)
				sub.EndDate = timePtr(date(2025, time.April, 1))
			},
			assertAfter: func(t *testing.T, updated *model.Subscription) {
				if updated.ServiceName != "Netflix" {
					t.Fatalf("unexpected ServiceName: got %q want %q", updated.ServiceName, "Netflix")
				}
				if updated.Price != 999 {
					t.Fatalf("unexpected Price: got %d want %d", updated.Price, 999)
				}
				if !updated.StartDate.Equal(date(2025, time.March, 1)) {
					t.Fatalf("unexpected StartDate: got %v want %v", updated.StartDate, date(2025, time.March, 1))
				}
				assertEndDateEqual(t, updated.EndDate, timePtr(date(2025, time.April, 1)))
			},
		},
		{
			name: "updates end date to nil",
			prepare: func(t *testing.T) *model.Subscription {
				return createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       200,
					UserID:      uuid.New(),
					StartDate:   date(2025, time.January, 1),
					EndDate:     timePtr(date(2025, time.February, 1)),
				})
			},
			update: func(sub *model.Subscription) {
				sub.EndDate = nil
			},
			assertAfter: func(t *testing.T, updated *model.Subscription) {
				if updated.EndDate != nil {
					t.Fatalf("expected EndDate to be nil, got %v", *updated.EndDate)
				}
			},
		},
		{
			name: "returns not found when subscription does not exist",
			prepare: func(t *testing.T) *model.Subscription {
				return &model.Subscription{
					ID:          uuid.New(),
					ServiceName: "Spotify",
					Price:       200,
					UserID:      uuid.New(),
					StartDate:   date(2025, time.January, 1),
					EndDate:     nil,
				}
			},
			update: func(sub *model.Subscription) {},
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateSubscriptions(t, db)

			sub := tt.prepare(t)
			oldUpdatedAt := sub.UpdatedAt

			if tt.update != nil {
				time.Sleep(5 * time.Millisecond)
				tt.update(sub)
			}

			err := repo.Update(context.Background(), sub)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				return
			}

			if sub.UpdatedAt.IsZero() {
				t.Fatal("expected UpdatedAt to be set")
			}

			if !sub.UpdatedAt.After(oldUpdatedAt) {
				t.Fatalf("expected UpdatedAt to be after previous value, old=%v new=%v", oldUpdatedAt, sub.UpdatedAt)
			}

			got, err := repo.GetByID(context.Background(), sub.ID)
			if err != nil {
				t.Fatalf("failed to get updated subscription: %v", err)
			}

			if got.ID != sub.ID {
				t.Fatalf("unexpected ID: got %v want %v", got.ID, sub.ID)
			}
			if got.ServiceName != sub.ServiceName {
				t.Fatalf("unexpected ServiceName: got %q want %q", got.ServiceName, sub.ServiceName)
			}
			if got.Price != sub.Price {
				t.Fatalf("unexpected Price: got %d want %d", got.Price, sub.Price)
			}
			if got.UserID != sub.UserID {
				t.Fatalf("unexpected UserID: got %v want %v", got.UserID, sub.UserID)
			}
			if !got.StartDate.Equal(sub.StartDate) {
				t.Fatalf("unexpected StartDate: got %v want %v", got.StartDate, sub.StartDate)
			}
			assertEndDateEqual(t, got.EndDate, sub.EndDate)

			if got.UpdatedAt.IsZero() {
				t.Fatal("expected persisted UpdatedAt to be set")
			}

			if tt.assertAfter != nil {
				tt.assertAfter(t, got)
			}
		})
	}
}

func TestSubscriptionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	cleanupSubscriptionsTable(t, db)

	repo := NewSubscriptionRepository(db)

	tests := []struct {
		name        string
		prepare     func(t *testing.T) uuid.UUID
		wantErr     error
		assertAfter func(t *testing.T, id uuid.UUID)
	}{
		{
			name: "deletes subscription successfully",
			prepare: func(t *testing.T) uuid.UUID {
				sub := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       200,
					UserID:      uuid.New(),
					StartDate:   date(2025, time.January, 1),
					EndDate:     timePtr(date(2025, time.February, 1)),
				})

				return sub.ID
			},
			assertAfter: func(t *testing.T, id uuid.UUID) {
				_, err := repo.GetByID(context.Background(), id)
				if !errors.Is(err, repository.ErrNotFound) {
					t.Fatalf("expected ErrNotFound after delete, got %v", err)
				}
			},
		},
		{
			name: "returns not found when subscription does not exist",
			prepare: func(t *testing.T) uuid.UUID {
				return uuid.New()
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "deletes only target subscription",
			prepare: func(t *testing.T) uuid.UUID {
				target := createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Spotify",
					Price:       200,
					UserID:      uuid.New(),
					StartDate:   date(2025, time.January, 1),
					EndDate:     nil,
				})

				_ = createSubscriptionWithData(t, repo, model.Subscription{
					ServiceName: "Netflix",
					Price:       500,
					UserID:      uuid.New(),
					StartDate:   date(2025, time.February, 1),
					EndDate:     nil,
				})

				return target.ID
			},
			assertAfter: func(t *testing.T, id uuid.UUID) {
				_, err := repo.GetByID(context.Background(), id)
				if !errors.Is(err, repository.ErrNotFound) {
					t.Fatalf("expected deleted subscription to be absent, got %v", err)
				}

				got, err := repo.List(context.Background(), model.SubscriptionFilter{
					Limit:  10,
					Offset: 0,
				})
				if err != nil {
					t.Fatalf("failed to list subscriptions after delete: %v", err)
				}

				if len(got) != 1 {
					t.Fatalf("expected 1 remaining subscription, got %d", len(got))
				}

				if got[0].ServiceName != "Netflix" {
					t.Fatalf("unexpected remaining subscription: got %q want %q", got[0].ServiceName, "Netflix")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateSubscriptions(t, db)

			id := tt.prepare(t)

			err := repo.Delete(context.Background(), id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr != nil {
				return
			}

			if tt.assertAfter != nil {
				tt.assertAfter(t, id)
			}
		})
	}
}