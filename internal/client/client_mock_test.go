package client

import (
	"clienthttp/internal/structs"
	"context"
)

// mockClientHttp is a test helper to mock the ClientHttp structure
type mockClientHttp struct {
	doRequestFunc func(ctx context.Context, method string, endpoint string, queryParams map[string]string, body []byte, headers map[string]string, options ...structs.NewRequestModifier) (*structs.Response, error)
}

// DoRequest implements the method to be called by the HTTP method functions
func (m *mockClientHttp) DoRequest(
	ctx context.Context,
	method string,
	endpoint string,
	queryParams map[string]string,
	body []byte,
	headers map[string]string,
	options ...structs.NewRequestModifier,
) (*structs.Response, error) {
	return m.doRequestFunc(ctx, method, endpoint, queryParams, body, headers, options...)
}

// Del implements the Del method for testing
func (m *mockClientHttp) Del(ctx context.Context, input structs.DelRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return m.DoRequest(
		ctx,
		"DELETE",
		input.Endpoint,
		input.QueryParams,
		nil,
		input.Headers,
		options...,
	)
}

// Get implements the Get method for testing
func (m *mockClientHttp) Get(ctx context.Context, input structs.GetRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return m.DoRequest(
		ctx,
		"GET",
		input.Endpoint,
		input.QueryParams,
		nil,
		input.Headers,
		options...,
	)
}

// Patch implements the Patch method for testing
func (m *mockClientHttp) Patch(ctx context.Context, input structs.PatchRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return m.DoRequest(
		ctx,
		"PATCH",
		input.Endpoint,
		input.QueryParams,
		input.Body,
		input.Headers,
		options...,
	)
}

// Post implements the Post method for testing
func (m *mockClientHttp) Post(ctx context.Context, input structs.PostRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return m.DoRequest(
		ctx,
		"POST",
		input.Endpoint,
		input.QueryParams,
		input.Body,
		input.Headers,
		options...,
	)
}

// Put implements the Put method for testing
func (m *mockClientHttp) Put(ctx context.Context, input structs.PutRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return m.DoRequest(
		ctx,
		"PUT",
		input.Endpoint,
		input.QueryParams,
		input.Body,
		input.Headers,
		options...,
	)
}
