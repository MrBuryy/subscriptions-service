package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/MrBuryy/subscriptions-service/internal/handler/dto"
	"github.com/MrBuryy/subscriptions-service/internal/model"
	"github.com/MrBuryy/subscriptions-service/internal/service"
)

type SubscriptionService interface {
	Create(
		ctx context.Context,
		serviceName string,
		price int,
		userID string,
		startDate string,
		endDate *string,
	) (*model.Subscription, error)

	GetByID(ctx context.Context, id string) (*model.Subscription, error)
}

type SubscriptionHandler struct {
	service SubscriptionService
}

func NewSubscriptionHandler(service SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
	}
}

func toSubscriptionResponse(sub *model.Subscription) dto.SubscriptionResponse {
	var endDate *string
	if sub.EndDate != nil {
		formatted := service.FormatMonthYear(*sub.EndDate)
		endDate = &formatted
	}

	return dto.SubscriptionResponse{
		ID:          sub.ID.String(),
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   service.FormatMonthYear(sub.StartDate),
		EndDate:     endDate,
		CreatedAt:   sub.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   sub.UpdatedAt.Format(time.RFC3339),
	}
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := dto.APIResponse{
		Data:  data,
		Error: nil,
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(response)
}

func WriteError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := dto.APIResponse{
		Data: nil,
		Error: &dto.APIError{
			Code:    code,
			Message: message,
		},
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(response)
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			dto.ErrorCodeInvalidRequest,
			"invalid request body",
		)
		return
	}

	sub, err := h.service.Create(
		r.Context(),
		req.ServiceName,
		req.Price,
		req.UserID,
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidService),
			errors.Is(err, service.ErrInvalidPrice),
			errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidStartDate),
			errors.Is(err, service.ErrInvalidEndDate):
			WriteError(
				w,
				http.StatusBadRequest,
				dto.ErrorCodeValidationError,
				err.Error(),
			)
			return
		default:
			WriteError(
				w,
				http.StatusInternalServerError,
				dto.ErrorCodeInternalError,
				"internal server error",
			)
			return
		}
	}

	WriteJSON(w, http.StatusCreated, toSubscriptionResponse(sub))
}

func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidID):
			WriteError(
				w,
				http.StatusBadRequest,
				dto.ErrorCodeValidationError,
				err.Error(),
			)
			return
		case errors.Is(err, service.ErrNotFound):
			WriteError(
				w,
				http.StatusNotFound,
				dto.ErrorCodeNotFound,
				"subscription not found",
			)
			return
		default:
			WriteError(
				w,
				http.StatusInternalServerError,
				dto.ErrorCodeInternalError,
				"internal server error",
			)
			return
		}
	}

	WriteJSON(w, http.StatusOK, toSubscriptionResponse(sub))
}