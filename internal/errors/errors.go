package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Configuration errors
	ErrConfigLoad    ErrorType = "config_load_error"
	ErrConfigSave    ErrorType = "config_save_error"
	ErrConfigInvalid ErrorType = "config_invalid_error"

	// API errors
	ErrAPIRequest   ErrorType = "api_request_error"
	ErrAPIResponse  ErrorType = "api_response_error"
	ErrAPIAuth      ErrorType = "api_auth_error"
	ErrAPIRateLimit ErrorType = "api_rate_limit_error"

	// Bot errors
	ErrBotInit    ErrorType = "bot_init_error"
	ErrBotSend    ErrorType = "bot_send_error"
	ErrBotReceive ErrorType = "bot_receive_error"

	// Storage errors
	ErrStorageRead   ErrorType = "storage_read_error"
	ErrStorageWrite  ErrorType = "storage_write_error"
	ErrStorageDelete ErrorType = "storage_delete_error"
)

// CustomError represents a custom error with type and context
type CustomError struct {
	Type    ErrorType
	Message string
	Context map[string]interface{}
	Err     error
}

// Error implements the error interface
func (e *CustomError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s", e.Type, e.Message))

	if e.Context != nil {
		sb.WriteString(" Context: ")
		for k, v := range e.Context {
			sb.WriteString(fmt.Sprintf("%s=%v ", k, v))
		}
	}

	if e.Err != nil {
		sb.WriteString(fmt.Sprintf(" Original error: %v", e.Err))
	}

	return sb.String()
}

// Unwrap implements the errors.Unwrap interface
func (e *CustomError) Unwrap() error {
	return e.Err
}

// New creates a new custom error
func New(errType ErrorType, message string, context map[string]interface{}, err error) *CustomError {
	return &CustomError{
		Type:    errType,
		Message: message,
		Context: context,
		Err:     err,
	}
}

// IsType checks if an error is of a specific type
func IsType(err error, errType ErrorType) bool {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type == errType
	}
	return false
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, message string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}
	return New(errType, message, context, err)
}
