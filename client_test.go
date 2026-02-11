package clienthttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================================
// Client Creation Tests
// ============================================================================

func TestNew_ValidURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https URL", "https://api.example.com", false},
		{"http URL", "http://localhost:8080", false},
		{"URL with trailing slash", "https://api.example.com/", false},
		{"invalid URL", "not-a-url", true},
		{"ftp URL", "ftp://files.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("New(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestNew_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(server.URL,
		WithTimeout(15*time.Second),
		WithDialTimeout(5*time.Second),
		WithMaxIdleConns(50),
		WithMaxIdleConnsPerHost(10),
		WithMaxConnsPerHost(25),
		WithIdleConnTimeout(60*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify options were applied
	if client.config.timeout != 15*time.Second {
		t.Errorf("timeout = %v, want %v", client.config.timeout, 15*time.Second)
	}
	if client.config.dialTimeout != 5*time.Second {
		t.Errorf("dialTimeout = %v, want %v", client.config.dialTimeout, 5*time.Second)
	}
	if client.config.transport.maxIdleConns != 50 {
		t.Errorf("maxIdleConns = %v, want %v", client.config.transport.maxIdleConns, 50)
	}
}

func TestNew_DefaultConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify defaults
	if client.config.timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", client.config.timeout, DefaultTimeout)
	}
	if client.config.dialTimeout != DefaultDialTimeout {
		t.Errorf("dialTimeout = %v, want %v", client.config.dialTimeout, DefaultDialTimeout)
	}
	if client.config.transport.maxIdleConns != DefaultMaxIdleConns {
		t.Errorf("maxIdleConns = %v, want %v", client.config.transport.maxIdleConns, DefaultMaxIdleConns)
	}
}

// ============================================================================
// HTTP Method Tests
// ============================================================================

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "hello"}`))
	}))
	defer server.Close()

	client, _ := New(server.URL)
	resp, err := client.Get(context.Background(), "/test")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !resp.OK() {
		t.Error("Expected OK() to be true")
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	client, _ := New(server.URL)
	resp, err := client.Post(context.Background(), "/users", []byte(`{"name": "John"}`))

	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL)
	resp, err := client.Put(context.Background(), "/users/1", []byte(`{"name": "Jane"}`))

	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL)
	resp, err := client.Patch(context.Background(), "/users/1", []byte(`{"name": "Jane"}`))

	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, _ := New(server.URL)
	resp, err := client.Delete(context.Background(), "/users/1")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

// ============================================================================
// Request Options Tests
// ============================================================================

func TestClient_WithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		limit := r.URL.Query().Get("limit")

		if page != "1" || limit != "10" {
			t.Errorf("Query params: page=%s, limit=%s, want page=1, limit=10", page, limit)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL)
	_, err := client.Get(context.Background(), "/users",
		WithQuery("page", "1"),
		WithQuery("limit", "10"),
	)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestClient_WithHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		custom := r.Header.Get("X-Custom-Header")
		if custom != "custom-value" {
			t.Errorf("X-Custom-Header = %s, want custom-value", custom)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL)
	_, err := client.Get(context.Background(), "/test",
		WithHeader("X-Custom-Header", "custom-value"),
	)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestClient_WithBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-token" {
			t.Errorf("Authorization = %s, want Bearer my-token", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL)
	_, err := client.Get(context.Background(), "/protected",
		WithBearerToken("my-token"),
	)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestClient_WithBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			t.Errorf("BasicAuth: user=%s, pass=%s, ok=%v", user, pass, ok)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL)
	_, err := client.Get(context.Background(), "/admin",
		WithBasicAuth("admin", "secret"),
	)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

// ============================================================================
// Timeout Tests
// ============================================================================

func TestClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL, WithTimeout(100*time.Millisecond))
	_, err := client.Get(context.Background(), "/slow")

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Check error contains timeout indication
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestClient_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client, _ := New(server.URL)
	resp, err := client.Get(context.Background(), "/missing")

	if err == nil {
		t.Error("Expected error for 404 response")
	}

	// Response should still be available
	if resp == nil {
		t.Fatal("Expected response even with error")
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	// Check error type
	var httpErr *Error
	if !errors.As(err, &httpErr) {
		t.Error("Expected *Error type")
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("Error.StatusCode = %d, want %d", httpErr.StatusCode, http.StatusNotFound)
	}
}

// ============================================================================
// Response Methods Tests
// ============================================================================

func TestResponse_OK(t *testing.T) {
	tests := []struct {
		statusCode int
		want       bool
	}{
		{200, true},
		{201, true},
		{204, true},
		{299, true},
		{300, false},
		{400, false},
		{500, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			resp := &Response{StatusCode: tt.statusCode}
			if got := resp.OK(); got != tt.want {
				t.Errorf("OK() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_JSON(t *testing.T) {
	type Data struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	resp := &Response{
		StatusCode: 200,
		Body:       []byte(`{"id": 1, "name": "test"}`),
	}

	var data Data
	err := resp.JSON(&data)

	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}
	if data.ID != 1 || data.Name != "test" {
		t.Errorf("JSON() = %+v, want {ID:1 Name:test}", data)
	}
}

func TestResponse_String(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
		Body:       []byte("hello world"),
	}

	if got := resp.String(); got != "hello world" {
		t.Errorf("String() = %q, want %q", got, "hello world")
	}
}

// ============================================================================
// Connection Pool Tests
// ============================================================================

func TestClient_ConnectionPool(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL,
		WithMaxIdleConns(10),
		WithMaxIdleConnsPerHost(5),
		WithMaxConnsPerHost(10),
	)

	ctx := context.Background()

	// Make multiple sequential requests
	for i := 0; i < 5; i++ {
		_, err := client.Get(ctx, fmt.Sprintf("/test/%d", i))
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}

	if atomic.LoadInt32(&requestCount) != 5 {
		t.Errorf("Expected 5 requests, got %d", requestCount)
	}
}

// ============================================================================
// Auditor Tests
// ============================================================================

type testAuditor struct {
	requests  []*AuditRequest
	responses []*AuditResponse
}

func (a *testAuditor) Log(ctx context.Context, req *AuditRequest, resp *AuditResponse) {
	a.requests = append(a.requests, req)
	a.responses = append(a.responses, resp)
}

func TestClient_WithAuditor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	auditor := &testAuditor{}
	client, _ := New(server.URL, WithAuditor(auditor))

	_, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if len(auditor.requests) != 1 {
		t.Errorf("Expected 1 request logged, got %d", len(auditor.requests))
	}
	if len(auditor.responses) != 1 {
		t.Errorf("Expected 1 response logged, got %d", len(auditor.responses))
	}

	if auditor.requests[0].Method != http.MethodGet {
		t.Errorf("Request method = %s, want GET", auditor.requests[0].Method)
	}
	if auditor.responses[0].StatusCode != http.StatusOK {
		t.Errorf("Response status = %d, want 200", auditor.responses[0].StatusCode)
	}
}

// ============================================================================
// PostForm Tests
// ============================================================================

func TestClient_PostForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
			t.Errorf("Content-Type = %s, want application/x-www-form-urlencoded", ct)
		}

		r.ParseForm()
		if r.FormValue("username") != "john" || r.FormValue("password") != "secret" {
			t.Errorf("Form values: username=%s, password=%s", r.FormValue("username"), r.FormValue("password"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := New(server.URL)
	_, err := client.PostForm(context.Background(), "/login", map[string]string{
		"username": "john",
		"password": "secret",
	})

	if err != nil {
		t.Fatalf("PostForm failed: %v", err)
	}
}
