package client

import (
	"clienthttp/internal/structs"
	"context"
	"net/http"
)

func (c ClientHttp) Patch(ctx context.Context, input structs.PatchRequest, options ...structs.NewRequestModifier) (*structs.Response, error) {
	if input.QueryParams == nil {
		input.QueryParams = make(map[string]string)
	}

	return c.DoRequest(
		ctx,
		http.MethodPatch,
		input.Endpoint,
		input.QueryParams,
		input.Body,
		input.Headers,
		options...,
	)
}
