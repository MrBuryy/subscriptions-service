package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"strings"
	"strconv"

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

	List(
		ctx context.Context,
		userID *string,
		serviceName *string,
		limit int,
		offset int,
	) ([]model.Subscription, error)

	Update(
		ctx context.Context,
		id string,
		serviceName string,
		price int,
		userID string,
		startDate string,
		endDate *string,
	) (*model.Subscription, error)

	Delete(ctx context.Context, id string) error

	Summary(
		ctx context.Context,
		from string,
		to string,
		userID *string,
		serviceName *string,
	) (int, error)
}

type SubscriptionHandler struct {
	service SubscriptionService
}

func NewSubscriptionHandler(service SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
	}
}

// Create godoc
// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body dto.CreateSubscriptionRequest true "Create subscription request"
// @Success 201 {object} dto.SubscriptionResponseEnvelope
// @Failure 400 {object} dto.ErrorResponseEnvelope
// @Failure 500 {object} dto.ErrorResponseEnvelope
// @Router /subscriptions [post]
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

// GetByID godoc
// @Summary Get subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} dto.SubscriptionResponseEnvelope
// @Failure 400 {object} dto.ErrorResponseEnvelope
// @Failure 404 {object} dto.ErrorResponseEnvelope
// @Failure 500 {object} dto.ErrorResponseEnvelope
// @Router /subscriptions/{id} [get]
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

// List godoc
// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param service_name query string false "Filter by service name"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} dto.SubscriptionListResponseEnvelope
// @Failure 400 {object} dto.ErrorResponseEnvelope
// @Failure 500 {object} dto.ErrorResponseEnvelope
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var userID *string
	if value := strings.TrimSpace(query.Get("user_id")); value != "" {
		userID = &value
	}

	var serviceName *string
	if value := strings.TrimSpace(query.Get("service_name")); value != "" {
		serviceName = &value
	}

	limit := 10
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid limit")
			return
		}
		limit = parsedLimit
	}

	offset := 0
	if rawOffset := strings.TrimSpace(query.Get("offset")); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil {
			WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid offset")
			return
		}
		offset = parsedOffset
	}

	subs, err := h.service.List(r.Context(), userID, serviceName, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidService),
			errors.Is(err, service.ErrInvalidOffset):
			WriteError(w, http.StatusBadRequest, dto.ErrorCodeValidationError, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, dto.ErrorCodeInternalError, "internal server error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, toSubscriptionResponses(subs))
}

// Update godoc
// @Summary Update subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param request body dto.UpdateSubscriptionRequest true "Update subscription request"
// @Success 200 {object} dto.SubscriptionResponseEnvelope
// @Failure 400 {object} dto.ErrorResponseEnvelope
// @Failure 404 {object} dto.ErrorResponseEnvelope
// @Failure 500 {object} dto.ErrorResponseEnvelope
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "id is required")
		return
	}

	var req dto.UpdateSubscriptionRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "invalid request body")
		return
	}

	sub, err := h.service.Update(
		r.Context(),
		id,
		req.ServiceName,
		req.Price,
		req.UserID,
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidID),
			errors.Is(err, service.ErrInvalidService),
			errors.Is(err, service.ErrInvalidPrice),
			errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidStartDate),
			errors.Is(err, service.ErrInvalidEndDate):
			WriteError(w, http.StatusBadRequest, dto.ErrorCodeValidationError, err.Error())
		case errors.Is(err, service.ErrNotFound):
			WriteError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "subscription not found")
		default:
			WriteError(w, http.StatusInternalServerError, dto.ErrorCodeInternalError, "internal server error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, toSubscriptionResponse(sub))
}

// Delete godoc
// @Summary Delete subscription
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponseEnvelope
// @Failure 404 {object} dto.ErrorResponseEnvelope
// @Failure 500 {object} dto.ErrorResponseEnvelope
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "id is required")
		return
	}

	err := h.service.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidID):
			WriteError(w, http.StatusBadRequest, dto.ErrorCodeValidationError, err.Error())
		case errors.Is(err, service.ErrNotFound):
			WriteError(w, http.StatusNotFound, dto.ErrorCodeNotFound, "subscription not found")
		default:
			WriteError(w, http.StatusInternalServerError, dto.ErrorCodeInternalError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Summary godoc
// @Summary Get subscriptions summary
// @Tags subscriptions
// @Produce json
// @Param from query string true "Start period in MM-YYYY format"
// @Param to query string true "End period in MM-YYYY format"
// @Param user_id query string false "Filter by user ID"
// @Param service_name query string false "Filter by service name"
// @Success 200 {object} dto.SummaryResponseEnvelope
// @Failure 400 {object} dto.ErrorResponseEnvelope
// @Failure 500 {object} dto.ErrorResponseEnvelope
// @Router /subscriptions/summary [get]
func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	from := strings.TrimSpace(query.Get("from"))
	to := strings.TrimSpace(query.Get("to"))

	if from == "" || to == "" {
		WriteError(w, http.StatusBadRequest, dto.ErrorCodeInvalidRequest, "from and to are required")
		return
	}

	var userID *string
	if value := strings.TrimSpace(query.Get("user_id")); value != "" {
		userID = &value
	}

	var serviceName *string
	if value := strings.TrimSpace(query.Get("service_name")); value != "" {
		serviceName = &value
	}

	total, err := h.service.Summary(r.Context(), from, to, userID, serviceName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFrom),
			errors.Is(err, service.ErrInvalidTo),
			errors.Is(err, service.ErrInvalidPeriod),
			errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidService):
			WriteError(w, http.StatusBadRequest, dto.ErrorCodeValidationError, err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, dto.ErrorCodeInternalError, "internal server error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, dto.SummaryResponse{
		TotalPrice: total,
	})
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

func toSubscriptionResponses(subs []model.Subscription) []dto.SubscriptionResponse {
	resp := make([]dto.SubscriptionResponse, 0, len(subs))
	for i := range subs {
		resp = append(resp, toSubscriptionResponse(&subs[i]))
	}
	return resp
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