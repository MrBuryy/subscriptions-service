package dto

type APIResponse struct {
	Data  any       `json:"data"`
	Error *APIError `json:"error"`
}

