package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/MrBuryy/subscriptions-service/internal/handler/dto"
	"github.com/MrBuryy/subscriptions-service/internal/model"
	"github.com/MrBuryy/subscriptions-service/internal/service"
)

type fakeSubscriptionService struct {
	createFn func(
		ctx context.Context,
		serviceName string,
		price int,
		userID string,
		startDate string,
		endDate *string,
	) (*model.Subscription, error)

	getByIDFn func(ctx context.Context, id string) (*model.Subscription, error)
}

func (f *fakeSubscriptionService) Create(
	ctx context.Context,
	serviceName string,
	price int,
	userID string,
	startDate string,
	endDate *string,
) (*model.Subscription, error) {
	return f.createFn(ctx, serviceName, price, userID, startDate, endDate)
}

func (f *fakeSubscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	return f.getByIDFn(ctx, id)
}

func testSubscription() *model.Subscription {
	now := time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)

	return &model.Subscription{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ServiceName: "Yandex Plus",
		Price:       499,
		UserID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		StartDate:   time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     &endDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestSubscriptionHandler_Create_Success(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		createFn: func(
			ctx context.Context,
			serviceName string,
			price int,
			userID string,
			startDate string,
			endDate *string,
		) (*model.Subscription, error) {
			if serviceName != "Yandex Plus" {
				t.Fatalf("unexpected serviceName: %s", serviceName)
			}
			if price != 499 {
				t.Fatalf("unexpected price: %d", price)
			}
			if userID != "22222222-2222-2222-2222-222222222222" {
				t.Fatalf("unexpected userID: %s", userID)
			}
			if startDate != "01-2026" {
				t.Fatalf("unexpected startDate: %s", startDate)
			}
			if endDate == nil || *endDate != "12-2026" {
				t.Fatalf("unexpected endDate: %v", endDate)
			}

			return testSubscription(), nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	body := `{
		"service_name":"Yandex Plus",
		"price":499,
		"user_id":"22222222-2222-2222-2222-222222222222",
		"start_date":"01-2026",
		"end_date":"12-2026"
	}`

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp dto.APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("expected nil error, got %+v", resp.Error)
	}
}

func TestSubscriptionHandler_Create_InvalidJSON(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		createFn: func(
			ctx context.Context,
			serviceName string,
			price int,
			userID string,
			startDate string,
			endDate *string,
		) (*model.Subscription, error) {
			t.Fatal("service Create must not be called")
			return nil, nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	body := `{"service_name":"Yandex Plus","unknown":"field"}`
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Create_ValidationError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		createFn: func(
			ctx context.Context,
			serviceName string,
			price int,
			userID string,
			startDate string,
			endDate *string,
		) (*model.Subscription, error) {
			return nil, service.ErrInvalidPrice
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	body := `{
		"service_name":"Yandex Plus",
		"price":0,
		"user_id":"22222222-2222-2222-2222-222222222222",
		"start_date":"01-2026"
	}`

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Create_InternalError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		createFn: func(
			ctx context.Context,
			serviceName string,
			price int,
			userID string,
			startDate string,
			endDate *string,
		) (*model.Subscription, error) {
			return nil, errors.New("db is down")
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	body := `{
		"service_name":"Yandex Plus",
		"price":499,
		"user_id":"22222222-2222-2222-2222-222222222222",
		"start_date":"01-2026"
	}`

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_Success(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		getByIDFn: func(ctx context.Context, id string) (*model.Subscription, error) {
			if id != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("unexpected id: %s", id)
			}
			return testSubscription(), nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "11111111-1111-1111-1111-111111111111")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp dto.APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("expected nil error, got %+v", resp.Error)
	}
}

func TestSubscriptionHandler_GetByID_InvalidID(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		getByIDFn: func(ctx context.Context, id string) (*model.Subscription, error) {
			return nil, service.ErrInvalidID
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/bad-id", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "bad-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_NotFound(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		getByIDFn: func(ctx context.Context, id string) (*model.Subscription, error) {
			return nil, service.ErrNotFound
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "11111111-1111-1111-1111-111111111111")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_InternalError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		getByIDFn: func(ctx context.Context, id string) (*model.Subscription, error) {
			return nil, errors.New("db error")
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "11111111-1111-1111-1111-111111111111")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}