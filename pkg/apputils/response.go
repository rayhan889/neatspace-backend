package apputils

import (
	"net/http"

	"github.com/rayhan889/intern-payroll/internal/application"
)

// BaseResponse defines the standard structure for API responses.
type BaseResponse struct {
	StatusCode   int                `json:"status_code"`
	ErrorMessage string             `json:"error_message,omitempty"`
	StackTrace   string             `json:"stack_trace,omitempty"`
	Data         *interface{}       `json:"data"`
	Errors       *[]ErrorValidation `json:"errors,omitempty"`
	Version      string             `json:"version"`
}

// ErrorValidation represents a single validation error.
type ErrorValidation struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}

// SuccessResponse creates a successful response with the given data.
func SuccessResponse(data interface{}) *BaseResponse {
	return &(BaseResponse{
		StatusCode: http.StatusOK,
		Data:       &data,
		Version:    application.Version,
	})
}

// ErrorResponse creates an error response with the given status code and error message.
func ErrorResponse(statusCode int, errorMessage string, stackTrace string) *BaseResponse {
	return &(BaseResponse{
		StatusCode:   statusCode,
		StackTrace:   stackTrace,
		ErrorMessage: errorMessage,
		Version:      application.Version,
	})
}

// ErrorValidationResponse creates a validation error response with the given status code, errors, and error message.
func ErrorValidationResponse(statusCode int, errors []ErrorValidation, errorMessage string) *BaseResponse {
	return &(BaseResponse{
		StatusCode:   statusCode,
		ErrorMessage: errorMessage,
		Errors:       &errors,
		Version:      application.Version,
	})
}
