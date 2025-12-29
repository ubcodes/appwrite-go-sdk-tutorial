package models

// APIError represents a structured error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   APIError  `json:"error"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message, field string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error: APIError{
			Code:    code,
			Message: message,
			Field:   field,
		},
	}
}


