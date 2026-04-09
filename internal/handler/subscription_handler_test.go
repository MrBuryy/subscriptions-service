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

	listFn func(
		ctx context.Context,
		userID *string,
		serviceName *string,
		limit int,
		offset int,
	) ([]model.Subscription, error)

	updateFn func(
		ctx context.Context,
		id string,
		serviceName string,
		price int,
		userID string,
		startDate string,
		endDate *string,
	) (*model.Subscription, error)

	deleteFn func(ctx context.Context, id string) error

	summaryFn func(
		ctx context.Context,
		from string,
		to string,
		userID *string,
		serviceName *string,
	) (int, error)
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

func (f *fakeSubscriptionService) List(
	ctx context.Context,
	userID *string,
	serviceName *string,
	limit int,
	offset int,
) ([]model.Subscription, error) {
	return f.listFn(ctx, userID, serviceName, limit, offset)
}

func (f *fakeSubscriptionService) Update(
	ctx context.Context,
	id string,
	serviceName string,
	price int,
	userID string,
	startDate string,
	endDate *string,
) (*model.Subscription, error) {
	return f.updateFn(ctx, id, serviceName, price, userID, startDate, endDate)
}

func (f *fakeSubscriptionService) Delete(ctx context.Context, id string) error {
	return f.deleteFn(ctx, id)
}

func (f *fakeSubscriptionService) Summary(
	ctx context.Context,
	from string,
	to string,
	userID *string,
	serviceName *string,
) (int, error) {
	return f.summaryFn(ctx, from, to, userID, serviceName)
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

func withURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
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

	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")

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

	req = withURLParam(req, "id", "bad-id")

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

	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")

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

	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")

	h.GetByID(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSubscriptionHandler_List_Success(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		listFn: func(
			ctx context.Context,
			userID *string,
			serviceName *string,
			limit int,
			offset int,
		) ([]model.Subscription, error) {
			if userID == nil || *userID != "22222222-2222-2222-2222-222222222222" {
				t.Fatalf("unexpected userID: %v", userID)
			}
			if serviceName == nil || *serviceName != "Yandex Plus" {
				t.Fatalf("unexpected serviceName: %v", serviceName)
			}
			if limit != 20 {
				t.Fatalf("unexpected limit: %d", limit)
			}
			if offset != 5 {
				t.Fatalf("unexpected offset: %d", offset)
			}

			return []model.Subscription{*testSubscription()}, nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions?user_id=22222222-2222-2222-2222-222222222222&service_name=Yandex%20Plus&limit=20&offset=5",
		nil,
	)
	rec := httptest.NewRecorder()

	h.List(rec, req)

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

func TestSubscriptionHandler_List_InvalidLimit(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		listFn: func(
			ctx context.Context,
			userID *string,
			serviceName *string,
			limit int,
			offset int,
		) ([]model.Subscription, error) {
			t.Fatal("service List must not be called")
			return nil, nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions?limit=abc", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_List_InvalidOffset(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		listFn: func(
			ctx context.Context,
			userID *string,
			serviceName *string,
			limit int,
			offset int,
		) ([]model.Subscription, error) {
			t.Fatal("service List must not be called")
			return nil, nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions?offset=abc", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_List_ValidationError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		listFn: func(
			ctx context.Context,
			userID *string,
			serviceName *string,
			limit int,
			offset int,
		) ([]model.Subscription, error) {
			return nil, service.ErrInvalidOffset
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions?offset=-1", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_List_InternalError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		listFn: func(
			ctx context.Context,
			userID *string,
			serviceName *string,
			limit int,
			offset int,
		) ([]model.Subscription, error) {
			return nil, errors.New("db error")
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSubscriptionHandler_Update_Success(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		updateFn: func(
			ctx context.Context,
			id string,
			serviceName string,
			price int,
			userID string,
			startDate string,
			endDate *string,
		) (*model.Subscription, error) {
			if id != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("unexpected id: %s", id)
			}
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

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(body))
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

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

func TestSubscriptionHandler_Update_InvalidJSON(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		updateFn: func(
			ctx context.Context,
			id string,
			serviceName string,
			price int,
			userID string,
			startDate string,
			endDate *string,
		) (*model.Subscription, error) {
			t.Fatal("service Update must not be called")
			return nil, nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	body := `{"service_name":"Yandex Plus","unknown":"field"}`
	req := httptest.NewRequest(http.MethodPut, "/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(body))
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Update_ValidationError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		updateFn: func(
			ctx context.Context,
			id string,
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

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(body))
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Update_NotFound(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		updateFn: func(
			ctx context.Context,
			id string,
			serviceName string,
			price int,
			userID string,
			startDate string,
			endDate *string,
		) (*model.Subscription, error) {
			return nil, service.ErrNotFound
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	body := `{
		"service_name":"Yandex Plus",
		"price":499,
		"user_id":"22222222-2222-2222-2222-222222222222",
		"start_date":"01-2026"
	}`

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(body))
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_Update_InternalError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		updateFn: func(
			ctx context.Context,
			id string,
			serviceName string,
			price int,
			userID string,
			startDate string,
			endDate *string,
		) (*model.Subscription, error) {
			return nil, errors.New("db error")
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	body := `{
		"service_name":"Yandex Plus",
		"price":499,
		"user_id":"22222222-2222-2222-2222-222222222222",
		"start_date":"01-2026"
	}`

	req := httptest.NewRequest(http.MethodPut, "/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(body))
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSubscriptionHandler_Delete_Success(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		deleteFn: func(ctx context.Context, id string) error {
			if id != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("unexpected id: %s", id)
			}
			return nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}
}

func TestSubscriptionHandler_Delete_InvalidID(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		deleteFn: func(ctx context.Context, id string) error {
			return service.ErrInvalidID
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/bad-id", nil)
	req = withURLParam(req, "id", "bad-id")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Delete_NotFound(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		deleteFn: func(ctx context.Context, id string) error {
			return service.ErrNotFound
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_Delete_InternalError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("db error")
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodDelete, "/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	req = withURLParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSubscriptionHandler_Summary_Success(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		summaryFn: func(
			ctx context.Context,
			from string,
			to string,
			userID *string,
			serviceName *string,
		) (int, error) {
			if from != "01-2026" {
				t.Fatalf("unexpected from: %s", from)
			}
			if to != "12-2026" {
				t.Fatalf("unexpected to: %s", to)
			}
			if userID == nil || *userID != "22222222-2222-2222-2222-222222222222" {
				t.Fatalf("unexpected userID: %v", userID)
			}
			if serviceName == nil || *serviceName != "Yandex Plus" {
				t.Fatalf("unexpected serviceName: %v", serviceName)
			}

			return 2200, nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/subscriptions/summary?from=01-2026&to=12-2026&user_id=22222222-2222-2222-2222-222222222222&service_name=Yandex%20Plus",
		nil,
	)
	rec := httptest.NewRecorder()

	h.Summary(rec, req)

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

	if !strings.Contains(rec.Body.String(), `"total_price":2200`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSubscriptionHandler_Summary_MissingQueryParams(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		summaryFn: func(
			ctx context.Context,
			from string,
			to string,
			userID *string,
			serviceName *string,
		) (int, error) {
			t.Fatal("service Summary must not be called")
			return 0, nil
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/summary?from=01-2026", nil)
	rec := httptest.NewRecorder()

	h.Summary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Summary_ValidationError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		summaryFn: func(
			ctx context.Context,
			from string,
			to string,
			userID *string,
			serviceName *string,
		) (int, error) {
			return 0, service.ErrInvalidPeriod
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/summary?from=12-2026&to=01-2026", nil)
	rec := httptest.NewRecorder()

	h.Summary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandler_Summary_InternalError(t *testing.T) {
	fakeSvc := &fakeSubscriptionService{
		summaryFn: func(
			ctx context.Context,
			from string,
			to string,
			userID *string,
			serviceName *string,
		) (int, error) {
			return 0, errors.New("db error")
		},
	}

	h := NewSubscriptionHandler(fakeSvc)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/summary?from=01-2026&to=12-2026", nil)
	rec := httptest.NewRecorder()

	h.Summary(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}