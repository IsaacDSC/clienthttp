package client

import (
	"clienthttp/internal/structs"
	"context"
	"net/http"
)

func (c ClientHttp) Get(ctx context.Context, input structs.GetRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return c.DoRequest(
		ctx,
		http.MethodGet,
		input.Endpoint,
		input.QueryParams,
		nil,
		input.Headers,
		options...,
	)
}
