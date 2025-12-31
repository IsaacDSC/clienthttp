package clienthttp

import (
	"clienthttp/internal/adapter"
	"clienthttp/internal/client"
	"clienthttp/internal/structs"
	"context"
)

// Client is the interface for making HTTP requests.
type Client interface {
	// Get performs an HTTP GET request.
	Get(ctx context.Context, input GetRequest, opts ...RequestModifier) (*Response, error)

	// Post performs an HTTP POST request.
	Post(ctx context.Context, input PostRequest, opts ...RequestModifier) (*Response, error)

	// Put performs an HTTP PUT request.
	Put(ctx context.Context, input PutRequest, opts ...RequestModifier) (*Response, error)

	// Patch performs an HTTP PATCH request.
	Patch(ctx context.Context, input PatchRequest, opts ...RequestModifier) (*Response, error)

	// Del performs an HTTP DELETE request.
	Del(ctx context.Context, input DelRequest, opts ...RequestModifier) (*Response, error)

	// DoRequest performs a custom HTTP request with the given method.
	DoRequest(ctx context.Context, method, endpoint string, queryParams map[string]string, body []byte, headers map[string]string, opts ...RequestModifier) (*Response, error)

	// DoFormRequest performs an HTTP POST request with form data.
	DoFormRequest(ctx context.Context, endpoint string, data map[string]string) (*Response, error)
}

// clientWrapper wraps the internal client to implement the public Client interface.
type clientWrapper struct {
	internal *client.ClientHttp
}

// New creates a new HTTP client with the given base URL and options.
// Returns an error if the base URL is invalid.
func New(baseUrl string, auditory AuditoryAdapter, correlation CorrelationIDAdapter, opts ...Option) (Client, error) {
	// Convert public options to internal options
	internalOpts := make([]client.Option, len(opts))
	for i, opt := range opts {
		internalOpts[i] = client.Option(opt)
	}

	// Convert adapters to internal types
	var internalAuditory adapter.AuditoryAdapter
	if auditory != nil {
		internalAuditory = &auditoryWrapper{adapter: auditory}
	}

	var internalCorrelation adapter.CorrelationIDAdapter
	if correlation != nil {
		internalCorrelation = adapter.CorrelationIDAdapter(correlation)
	}

	c, err := client.NewClientHttp(baseUrl, internalAuditory, internalCorrelation, internalOpts...)
	if err != nil {
		return nil, ErrInvalidBaseURL
	}

	return &clientWrapper{internal: c}, nil
}

// auditoryWrapper wraps the public AuditoryAdapter to implement the internal interface.
type auditoryWrapper struct {
	adapter AuditoryAdapter
}

func (w *auditoryWrapper) Save(ctx context.Context, request *structs.Request, response *structs.Response) {
	// Convert internal types to public types
	var pubRequest *Request
	if request != nil {
		pubRequest = &Request{
			Url:     request.Url,
			Method:  request.Method,
			Headers: request.Headers,
			Params:  request.Params,
			Cookies: request.Cookies,
			Body:    request.Body,
		}
	}

	var pubResponse *Response
	if response != nil {
		pubResponse = &Response{
			StatusCode: response.StatusCode,
			Body:       response.Body,
			Headers:    response.Headers,
		}
	}

	w.adapter.Save(ctx, pubRequest, pubResponse)
}

// Get implements Client.Get
func (c *clientWrapper) Get(ctx context.Context, input GetRequest, opts ...RequestModifier) (*Response, error) {
	internalInput := structs.GetRequest{
		BaseInput: structs.BaseInput{
			Endpoint:    input.Endpoint,
			QueryParams: input.QueryParams,
			Headers:     input.Headers,
		},
	}

	internalOpts := convertModifiers(opts)
	resp, err := c.internal.Get(ctx, internalInput, internalOpts...)
	if err != nil {
		return nil, err
	}

	return convertResponse(resp), nil
}

// Post implements Client.Post
func (c *clientWrapper) Post(ctx context.Context, input PostRequest, opts ...RequestModifier) (*Response, error) {
	internalInput := structs.PostRequest{
		BaseInput: structs.BaseInput{
			Endpoint:    input.Endpoint,
			QueryParams: input.QueryParams,
			Headers:     input.Headers,
		},
		Body: input.Body,
	}

	internalOpts := convertModifiers(opts)
	resp, err := c.internal.Post(ctx, internalInput, internalOpts...)
	if err != nil {
		return nil, err
	}

	return convertResponse(resp), nil
}

// Put implements Client.Put
func (c *clientWrapper) Put(ctx context.Context, input PutRequest, opts ...RequestModifier) (*Response, error) {
	internalInput := structs.PutRequest{
		BaseInput: structs.BaseInput{
			Endpoint:    input.Endpoint,
			QueryParams: input.QueryParams,
			Headers:     input.Headers,
		},
		Body: input.Body,
	}

	internalOpts := convertModifiers(opts)
	resp, err := c.internal.Put(ctx, internalInput, internalOpts...)
	if err != nil {
		return nil, err
	}

	return convertResponse(resp), nil
}

// Patch implements Client.Patch
func (c *clientWrapper) Patch(ctx context.Context, input PatchRequest, opts ...RequestModifier) (*Response, error) {
	internalInput := structs.PatchRequest{
		BaseInput: structs.BaseInput{
			Endpoint:    input.Endpoint,
			QueryParams: input.QueryParams,
			Headers:     input.Headers,
		},
		Body: input.Body,
	}

	internalOpts := convertModifiers(opts)
	resp, err := c.internal.Patch(ctx, internalInput, internalOpts...)
	if err != nil {
		return nil, err
	}

	return convertResponse(resp), nil
}

// Del implements Client.Del
func (c *clientWrapper) Del(ctx context.Context, input DelRequest, opts ...RequestModifier) (*Response, error) {
	internalInput := structs.DelRequest{
		BaseInput: structs.BaseInput{
			Endpoint:    input.Endpoint,
			QueryParams: input.QueryParams,
			Headers:     input.Headers,
		},
	}

	internalOpts := convertModifiers(opts)
	resp, err := c.internal.Del(ctx, internalInput, internalOpts...)
	if err != nil {
		return nil, err
	}

	return convertResponse(resp), nil
}

// DoRequest implements Client.DoRequest
func (c *clientWrapper) DoRequest(ctx context.Context, method, endpoint string, queryParams map[string]string, body []byte, headers map[string]string, opts ...RequestModifier) (*Response, error) {
	internalOpts := convertModifiers(opts)
	resp, err := c.internal.DoRequest(ctx, method, endpoint, queryParams, body, headers, internalOpts...)
	if err != nil {
		return nil, err
	}

	return convertResponse(resp), nil
}

// DoFormRequest implements Client.DoFormRequest
func (c *clientWrapper) DoFormRequest(ctx context.Context, endpoint string, data map[string]string) (*Response, error) {
	resp, err := c.internal.DoFormRequest(ctx, endpoint, data)
	if err != nil {
		return nil, err
	}

	return convertResponse(resp), nil
}

// Helper functions for type conversions

func convertResponse(resp *structs.Response) *Response {
	if resp == nil {
		return nil
	}
	return &Response{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Headers:    resp.Headers,
	}
}

func convertModifiers(opts []RequestModifier) []structs.NewRequestModifier {
	result := make([]structs.NewRequestModifier, len(opts))
	for i, opt := range opts {
		result[i] = structs.NewRequestModifier(opt)
	}
	return result
}
