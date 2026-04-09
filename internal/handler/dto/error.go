package dto

const (
	ErrorCodeInvalidRequest  = "invalid_request"
	ErrorCodeValidationError = "validation_error"
	ErrorCodeNotFound        = "not_found"
	ErrorCodeInternalError   = "internal_error"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}