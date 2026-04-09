package dto

type HealthData struct {
	Message string `json:"message"`
}

type HealthResponseEnvelope struct {
	Data  HealthData `json:"data"`
	Error *APIError  `json:"error"`
}

type SubscriptionResponseEnvelope struct {
	Data  SubscriptionResponse `json:"data"`
	Error *APIError            `json:"error"`
}

type SubscriptionListResponseEnvelope struct {
	Data  []SubscriptionResponse `json:"data"`
	Error *APIError              `json:"error"`
}

type SummaryResponseEnvelope struct {
	Data  SummaryResponse `json:"data"`
	Error *APIError       `json:"error"`
}

type ErrorResponseEnvelope struct {
	Data  any       `json:"data"`
	Error *APIError `json:"error"`
}