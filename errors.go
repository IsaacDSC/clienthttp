package clienthttp

import "errors"

// Sentinel errors for the clienthttp package.
// Use errors.Is() to check for these errors.
var (
	// ErrInvalidBaseURL is returned when the provided base URL is not a valid HTTP/HTTPS URL.
	ErrInvalidBaseURL = errors.New("clienthttp: invalid base URL")

	// ErrRequestFailed is returned when an HTTP request fails with a non-2xx status code.
	ErrRequestFailed = errors.New("clienthttp: request failed")

	// ErrReadResponseBody is returned when reading the response body fails.
	ErrReadResponseBody = errors.New("clienthttp: failed to read response body")
)

// RequestError wraps an error with additional context about the failed request.
type RequestError struct {
	StatusCode int
	URL        string
	Body       []byte
	Err        error
}

func (e *RequestError) Error() string {
	return e.Err.Error()
}

func (e *RequestError) Unwrap() error {
	return e.Err
}

