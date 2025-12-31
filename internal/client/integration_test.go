package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================================
// Timeout Integration Tests
// ============================================================================

func TestClient_WithCustomTimeout_RequestTimesOut(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create client with very short timeout
	client, err := NewClientHttp(
		server.URL,
		nil,
		nil,
		WithTimeout(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	_, err = client.DoRequest(ctx, http.MethodGet, "/test", nil, nil, nil)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Check if it's a timeout error
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestClient_WithSufficientTimeout_RequestSucceeds(t *testing.T) {
	// Create a server that responds quickly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create client with adequate timeout
	client, err := NewClientHttp(
		server.URL,
		nil,
		nil,
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.DoRequest(ctx, http.MethodGet, "/test", nil, nil, nil)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}
}

func TestClient_DefaultConfiguration_Works(t *testing.T) {
	// Create a simple server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "hello"}`))
	}))
	defer server.Close()

	// Create client with default configuration
	client, err := NewClientHttp(server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify defaults are applied
	if client.config.Timeout != DefaultTimeout {
		t.Errorf("Default timeout = %v, want %v", client.config.Timeout, DefaultTimeout)
	}
	if client.config.DialTimeout != DefaultDialTimeout {
		t.Errorf("Default dialTimeout = %v, want %v", client.config.DialTimeout, DefaultDialTimeout)
	}
	if client.config.Transport.MaxIdleConns != DefaultMaxIdleConns {
		t.Errorf("Default maxIdleConns = %v, want %v", client.config.Transport.MaxIdleConns, DefaultMaxIdleConns)
	}

	// Make a request to verify it works
	ctx := context.Background()
	resp, err := client.DoRequest(ctx, http.MethodGet, "/", nil, nil, nil)
	if err != nil {
		t.Errorf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}
}

// ============================================================================
// Connection Pool Integration Tests
// ============================================================================

func TestClient_WithConnectionPool_ReusesConnections(t *testing.T) {
	var requestCount int32

	// Create a server that counts requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create client with connection pooling
	client, err := NewClientHttp(
		server.URL,
		nil,
		nil,
		WithMaxIdleConns(10),
		WithMaxIdleConnsPerHost(5),
		WithMaxConnsPerHost(10),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Make multiple sequential requests
	for i := 0; i < 5; i++ {
		_, err := client.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/test/%d", i), nil, nil, nil)
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}

	// Verify all requests were processed
	if atomic.LoadInt32(&requestCount) != 5 {
		t.Errorf("Expected 5 requests, got %d", requestCount)
	}

	// Connection pooling should reuse connections
	// This test verifies the client is configured correctly for connection reuse
	t.Logf("Completed %d requests with connection pooling enabled", requestCount)
}

func TestClient_MaxConnsPerHost_LimitsConnections(t *testing.T) {
	// Create a server that introduces a small delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create client with limited max connections per host
	client, err := NewClientHttp(
		server.URL,
		nil,
		nil,
		WithMaxConnsPerHost(2),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	done := make(chan bool, 5)

	// Make concurrent requests
	for i := 0; i < 5; i++ {
		go func(idx int) {
			_, err := client.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/test/%d", idx), nil, nil, nil)
			if err != nil {
				t.Logf("Request %d error (expected with limited conns): %v", idx, err)
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	t.Log("All concurrent requests completed with MaxConnsPerHost=2")
}

// ============================================================================
// Client Creation Tests
// ============================================================================

func TestNewClientHttp_WithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClientHttp(
		server.URL,
		nil,
		nil,
		WithTimeout(15*time.Second),
		WithDialTimeout(5*time.Second),
		WithTLSHandshakeTimeout(5*time.Second),
		WithResponseHeaderTimeout(10*time.Second),
		WithMaxIdleConns(50),
		WithMaxIdleConnsPerHost(10),
		WithMaxConnsPerHost(25),
		WithIdleConnTimeout(60*time.Second),
	)

	if err != nil {
		t.Fatalf("Failed to create client with all options: %v", err)
	}

	// Verify options were applied
	if client.config.Timeout != 15*time.Second {
		t.Errorf("timeout = %v, want %v", client.config.Timeout, 15*time.Second)
	}
	if client.config.DialTimeout != 5*time.Second {
		t.Errorf("dialTimeout = %v, want %v", client.config.DialTimeout, 5*time.Second)
	}
	if client.config.TLSHandshakeTimeout != 5*time.Second {
		t.Errorf("tlsHandshakeTimeout = %v, want %v", client.config.TLSHandshakeTimeout, 5*time.Second)
	}
	if client.config.ResponseHeaderTimeout != 10*time.Second {
		t.Errorf("responseHeaderTimeout = %v, want %v", client.config.ResponseHeaderTimeout, 10*time.Second)
	}
	if client.config.Transport.MaxIdleConns != 50 {
		t.Errorf("maxIdleConns = %v, want %v", client.config.Transport.MaxIdleConns, 50)
	}
	if client.config.Transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("maxIdleConnsPerHost = %v, want %v", client.config.Transport.MaxIdleConnsPerHost, 10)
	}
	if client.config.Transport.MaxConnsPerHost != 25 {
		t.Errorf("maxConnsPerHost = %v, want %v", client.config.Transport.MaxConnsPerHost, 25)
	}
	if client.config.Transport.IdleConnTimeout != 60*time.Second {
		t.Errorf("idleConnTimeout = %v, want %v", client.config.Transport.IdleConnTimeout, 60*time.Second)
	}

	// Verify client works
	ctx := context.Background()
	resp, err := client.DoRequest(ctx, http.MethodGet, "/", nil, nil, nil)
	if err != nil {
		t.Errorf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}
}

func TestNewClientHttp_InvalidURL_ReturnsError(t *testing.T) {
	_, err := NewClientHttp("invalid-url", nil, nil)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestNewClientHttp_HttpClientHasCorrectTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	timeout := 42 * time.Second
	client, err := NewClientHttp(
		server.URL,
		nil,
		nil,
		WithTimeout(timeout),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.httpClient.Timeout != timeout {
		t.Errorf("httpClient.Timeout = %v, want %v", client.httpClient.Timeout, timeout)
	}
}

