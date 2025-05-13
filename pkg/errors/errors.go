package errors

import "fmt"

const (
	// General errors
	ErrServerInternal   = 1000 // Internal server error
	ErrResourceNotFound = 1001 // Not found
	ErrInvalidRequest   = 1002 // Bad request

	// Database errors
	ErrDatabaseConnection = 2000 // Database connection error
	ErrDatabaseQuery      = 2001 // Database query error
	ErrDatabaseInsert     = 2002 // Database insert error
	ErrDatabaseUpdate     = 2003 // Database update error
	ErrDatabaseDelete     = 2004 // Database delete error

	// Authentication errors
	ErrAuthUnauthorized       = 3000 // Unauthorized access
	ErrAuthForbidden          = 3001 // Forbidden access
	ErrAuthTokenExpired       = 3002 // Token has expired
	ErrAuthInvalidPassword    = 3003 // Invalid password
	ErrAuthPasswordHashFailed = 3004 // Failed to hash password
	ErrAuthPasswordMismatch   = 3005 // Password not matched
	ErrAuthPasswordNotChanged = 3006 // Old and new password should be different

	// Common errors
	ErrInvalidParse   = 4000 // Parse error, missing fields, etc.
	ErrInvalidData    = 4001 // Validation error
	ErrCacheSet       = 4002 // Set cache error
	ErrCacheGet       = 4003 // Get cache error
	ErrCacheDelete    = 4004 // Delete cache error
	ErrCacheList      = 4005 // List cache error
	ErrCacheKeyExists = 4006 // Exists cache error)
)

// AppError represents a custom error with a code and message.
type AppError struct {
	Code    int    `json:"code"`    // Error code
	Message string `json:"message"` // Error message
	Err     error  `json:"-"`       // Underlying error (optional)
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code: %d, message: %s, error: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// Wrap creates a new AppError with an underlying error.
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// New creates a new AppError without an underlying error.
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}
