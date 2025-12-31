package clienthttp

import (
	"context"
	"encoding/json"
	"net/http"
)

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// OK returns true if the status code is in the 2xx range.
func (r *Response) OK() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// JSON unmarshals the response body into the provided value.
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.Body, v)
}

// String returns the response body as a string.
func (r *Response) String() string {
	return string(r.Body)
}

// Auditor is an interface for auditing HTTP requests and responses.
// Implement this interface to log or save audit trails of HTTP traffic.
type Auditor interface {
	Log(ctx context.Context, req *AuditRequest, resp *AuditResponse)
}

// AuditRequest contains request information for auditing purposes.
type AuditRequest struct {
	URL     string
	Method  string
	Headers http.Header
	Body    []byte
}

// AuditResponse contains response information for auditing purposes.
type AuditResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// CorrelationIDFunc extracts or generates a correlation ID from the context.
type CorrelationIDFunc func(ctx context.Context) string
