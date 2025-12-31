package clienthttp

import "net/http"

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// IsStatusSuccessfully returns true if the status code is in the 2xx range.
func (r Response) IsStatusSuccessfully() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// Request represents an HTTP request for auditing purposes.
type Request struct {
	Url     string
	Method  string
	Headers map[string][]string
	Params  string
	Cookies []*http.Cookie
	Body    []byte
}

// RequestModifier is a function that can modify an HTTP request before it is sent.
type RequestModifier func(r *http.Request)

// BaseInput contains common fields for all request types.
type BaseInput struct {
	Endpoint    string
	QueryParams map[string]string
	Headers     map[string]string
}

// GetRequest represents input for a GET request.
type GetRequest struct {
	BaseInput
}

// DelRequest represents input for a DELETE request.
type DelRequest struct {
	BaseInput
}

// PatchRequest represents input for a PATCH request.
type PatchRequest struct {
	BaseInput
	Body []byte
}

// PostRequest represents input for a POST request.
type PostRequest struct {
	BaseInput
	Body []byte
}

// PutRequest represents input for a PUT request.
type PutRequest struct {
	BaseInput
	Body []byte
}

