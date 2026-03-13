package clienthttp

import (
	"errors"
	"fmt"
)

// Sentinel errors for the clienthttp package.
// Use errors.Is() to check for these errors.
var (
	// ErrInvalidURL is returned when the provided base URL is not valid.
	ErrInvalidURL = errors.New("clienthttp: invalid URL")

	// ErrInvalidPath is returned when the provided request path is not valid.
	// This typically happens when a full URL is passed instead of a path
	// relative to the client's base URL.
	ErrInvalidPath = errors.New("clienthttp: invalid path")

	// ErrTimeout is returned when a request times out.
	ErrTimeout = errors.New("clienthttp: request timeout")

	// ErrRequestFailed is returned when an HTTP request fails with a non-2xx status.
	ErrRequestFailed = errors.New("clienthttp: request failed")

	// ErrMaxRetriesExceeded is returned when all retry attempts have been exhausted.
	ErrMaxRetriesExceeded = errors.New("clienthttp: max retries exceeded")
)

// Error represents an HTTP client error with additional context.
// Use errors.As() to extract detailed error information.
type Error struct {
	Op         string // Operation that failed (e.g., "GET", "POST")
	URL        string // URL of the failed request
	StatusCode int    // HTTP status code (0 if not applicable)
	Body       []byte // Response body (may be nil)
	Err        error  // Underlying error
}

// Error returns a human-readable error message.
func (e *Error) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("clienthttp: %s %s: status %d", e.Op, e.URL, e.StatusCode)
	}
	if e.Err != nil {
		return fmt.Sprintf("clienthttp: %s %s: %v", e.Op, e.URL, e.Err)
	}
	return fmt.Sprintf("clienthttp: %s %s: unknown error", e.Op, e.URL)
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// newError creates a new Error with the given parameters.
func newError(op, url string, statusCode int, body []byte, err error) *Error {
	return &Error{
		Op:         op,
		URL:        url,
		StatusCode: statusCode,
		Body:       body,
		Err:        err,
	}
}
