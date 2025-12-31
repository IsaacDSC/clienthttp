package client

import (
	"clienthttp/internal/structs"
	"context"
	"net/http"
)

func (c ClientHttp) Del(ctx context.Context, input structs.DelRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return c.DoRequest(
		ctx,
		http.MethodDelete,
		input.Endpoint,
		input.QueryParams,
		nil,
		input.Headers,
		options...,
	)
}
