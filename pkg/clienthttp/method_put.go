package clienthttp

import (
	"clienthttp/pkg/structs"
	"context"
	"net/http"
)

func (c ClientHttp) Put(ctx context.Context, input structs.PutRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return c.DoRequest(
		ctx,
		http.MethodPut,
		input.Endpoint,
		input.QueryParams,
		input.Body,
		input.Headers,
		options...,
	)
}
