package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MrBuryy/subscriptions-service/internal/handler/dto"
)

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