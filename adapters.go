package clienthttp

import "context"

// AuditoryAdapter is an interface for auditing HTTP requests and responses.
// Implement this interface to save audit logs of all HTTP traffic.
type AuditoryAdapter interface {
	Save(ctx context.Context, request *Request, response *Response)
}

// CorrelationIDAdapter is a function that extracts or generates a correlation ID from the context.
// This is used to track requests across services.
type CorrelationIDAdapter func(ctx context.Context) string

