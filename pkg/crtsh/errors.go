// errors.go — Typed error types for structured error handling
package crtsh

import "fmt"

// ErrorType categorizes the kind of error returned by the SDK.
type ErrorType string

const (
	// ErrorTypeSearch indicates a general search failure.
	ErrorTypeSearch ErrorType = "search"
	// ErrorTypeNotFound indicates the requested resource was not found.
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeRateLimit indicates the request was rate-limited by crt.sh.
	ErrorTypeRateLimit ErrorType = "rate_limit"
	// ErrorTypeServer indicates a 5xx server error from crt.sh.
	ErrorTypeServer ErrorType = "server"
	// ErrorTypeParse indicates a failure to parse the response body.
	ErrorTypeParse ErrorType = "parse"
	// ErrorTypeRequest indicates a failure to construct or send the HTTP request.
	ErrorTypeRequest ErrorType = "request"
	// ErrorTypeInvalid indicates an invalid parameter was provided.
	ErrorTypeInvalid ErrorType = "invalid"
)

// Error represents a structured error returned by the crt.sh SDK.
// It implements the error interface and can be inspected using
// IsNotFoundError, IsRateLimitError, IsServerError, etc.
type Error struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error returns a human-readable error message.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause error, enabling errors.Is/As chaining.
func (e *Error) Unwrap() error {
	return e.Cause
}

// IsNotFoundError returns true if the error indicates the requested resource was not found.
func IsNotFoundError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeNotFound
	}
	return false
}

// IsRateLimitError returns true if the error indicates a rate limit was hit.
func IsRateLimitError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeRateLimit
	}
	return false
}

// IsServerError returns true if the error indicates a 5xx server error.
func IsServerError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeServer
	}
	return false
}

// IsParseError returns true if the error indicates a response parsing failure.
func IsParseError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeParse
	}
	return false
}

// IsInvalidError returns true if the error indicates an invalid parameter.
func IsInvalidError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeInvalid
	}
	return false
}
