package handler

import (
	"net/http"

	"github.com/MrBuryy/subscriptions-service/internal/handler/dto"
)

// Health godoc
// @Summary Health check
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthResponseEnvelope
// @Router /health [get]
func Health(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, dto.HealthData{
		Message: "ok",
	})
}