package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Details    string `json:"details,omitempty"`
	Timestamp  string `json:"timestamp"`
	StackTrace string `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *APIError) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":    e.Code,
			"message": e.Message,
			"details": e.Details,
			"timestamp": e.Timestamp,
		},
	}
}

func (e *APIError) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	json.NewEncoder(w).Encode(e.ToJSON())
}

func NewAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  currentTimestamp(),
		StackTrace: string(debug.Stack()),
	}
}

func NewAPIErrorf(statusCode int, code, format string, args ...interface{}) *APIError {
	return &APIError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		StatusCode: statusCode,
		Timestamp:  currentTimestamp(),
	}
}

func Unauthorized(message string) *APIError {
	return NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(message string) *APIError {
	return NewAPIError(http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(message string) *APIError {
	return NewAPIError(http.StatusNotFound, "NOT_FOUND", message)
}

func BadRequest(message string) *APIError {
	return NewAPIError(http.StatusBadRequest, "BAD_REQUEST", message)
}

func Conflict(message string) *APIError {
	return NewAPIError(http.StatusConflict, "CONFLICT", message)
}

func Internal(message string) *APIError {
	return NewAPIError(http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func RateLimit(retryAfter int) *APIError {
	return NewAPIError(http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
		fmt.Sprintf("Rate limit exceeded. Retry after %d seconds.", retryAfter))
}

func ValidationError(message string) *APIError {
	return NewAPIError(http.StatusBadRequest, "VALIDATION_ERROR", message)
}

func currentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
