package clienthttp

import (
	"clienthttp/pkg/structs"
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestClientHttp_Get(t *testing.T) {
	type args struct {
		ctx     context.Context
		input   structs.GetRequest
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
			name: "successful_get_request",
			mock: func(ctx context.Context, method string, endpoint string, queryParams map[string]string, body []byte, headers map[string]string, options ...structs.NewRequestModifier) (*structs.Response, error) {
				if method != http.MethodGet {
					t.Errorf("Expected method %s, got %s", http.MethodGet, method)
				}
				if endpoint != "/test-endpoint" {
					t.Errorf("Expected endpoint %s, got %s", "/test-endpoint", endpoint)
				}
				if body != nil {
					t.Errorf("Expected body to be nil for GET request")
				}
				return &structs.Response{
					StatusCode: 200,
					Body:       []byte(`{"data": "test"}`),
					Headers:    http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
			args: args{
				ctx: context.Background(),
				input: structs.GetRequest{
					BaseInput: structs.BaseInput{
						Endpoint:    "/test-endpoint",
						QueryParams: map[string]string{"param1": "value1", "param2": "value2"},
						Headers:     map[string]string{"Accept": "application/json"},
					},
				},
				options: []structs.NewRequestModifier{},
			},
			want: &structs.Response{
				StatusCode: 200,
				Body:       []byte(`{"data": "test"}`),
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
				input: structs.GetRequest{
					BaseInput: structs.BaseInput{
						Endpoint: "/test-endpoint",
						Headers:  map[string]string{},
						// QueryParams intentionally left nil
					},
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

			got, err := mockClient.Get(tt.args.ctx, tt.args.input, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientHttp.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientHttp.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
