package client

import (
	"clienthttp/internal/structs"
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestClientHttp_Patch(t *testing.T) {
	type args struct {
		ctx     context.Context
		input   structs.PatchRequest
		options []structs.NewRequestModifier
	}
	tests := []struct {
		name    string
		mock    func(ctx context.Context, method string, endpoint string, queryParams map[string]string, body []byte, headers map[string]string, options ...structs.NewRequestModifier) (*structs.Response, error)
		args    args
		want    *structs.Response
		wantErr bool
	}{
		{
			name: "successful_patch_request",
			mock: func(ctx context.Context, method string, endpoint string, queryParams map[string]string, body []byte, headers map[string]string, options ...structs.NewRequestModifier) (*structs.Response, error) {
				if method != http.MethodPatch {
					t.Errorf("Expected method %s, got %s", http.MethodPatch, method)
				}
				if endpoint != "/test-endpoint" {
					t.Errorf("Expected endpoint %s, got %s", "/test-endpoint", endpoint)
				}
				expectedBody := []byte(`{"key":"value"}`)
				if !reflect.DeepEqual(body, expectedBody) {
					t.Errorf("Expected body %s, got %s", expectedBody, body)
				}
				return &structs.Response{
					StatusCode: 200,
					Body:       []byte(`{"updated": true}`),
					Headers:    http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
			args: args{
				ctx: context.Background(),
				input: structs.PatchRequest{
					BaseInput: structs.BaseInput{
						Endpoint:    "/test-endpoint",
						QueryParams: map[string]string{"param1": "value1"},
						Headers:     map[string]string{"Content-Type": "application/json"},
					},
					Body: []byte(`{"key":"value"}`),
				},
				options: []structs.NewRequestModifier{},
			},
			want: &structs.Response{
				StatusCode: 200,
				Body:       []byte(`{"updated": true}`),
				Headers:    http.Header{"Content-Type": []string{"application/json"}},
			},
			wantErr: false,
		},
		{
			name: "nil_query_params_handled",
			mock: func(ctx context.Context, method string, endpoint string, queryParams map[string]string, body []byte, headers map[string]string, options ...structs.NewRequestModifier) (*structs.Response, error) {
				if queryParams == nil {
					t.Error("Expected queryParams to be initialized, got nil")
				}
				return &structs.Response{
					StatusCode: 200,
					Body:       []byte(`{}`),
					Headers:    http.Header{},
				}, nil
			},
			args: args{
				ctx: context.Background(),
				input: structs.PatchRequest{
					BaseInput: structs.BaseInput{
						Endpoint: "/test-endpoint",
						Headers:  map[string]string{},
						// QueryParams intentionally left nil
					},
					Body: []byte(`{}`),
				},
			},
			want: &structs.Response{
				StatusCode: 200,
				Body:       []byte(`{}`),
				Headers:    http.Header{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock client
			mockClient := &mockClientHttp{
				doRequestFunc: tt.mock,
			}

			got, err := mockClient.Patch(tt.args.ctx, tt.args.input, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientHttp.Patch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientHttp.Patch() = %v, want %v", got, tt.want)
			}
		})
	}
}
