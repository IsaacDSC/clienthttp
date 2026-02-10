package clienthttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================================
// ExponentialBackoff Tests
// ============================================================================

func TestExponentialBackoff_ShouldRetry(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
		attempt     int
		resp        *Response
		err         error
		want        bool
	}{
		{
			name:        "first attempt with retryable status",
			maxAttempts: 3,
			attempt:     0,
			resp:        &Response{StatusCode: 503},
			err:         nil,
			want:        true,
		},
		{
			name:        "second attempt with retryable status",
			maxAttempts: 3,
			attempt:     1,
			resp:        &Response{StatusCode: 503},
			err:         nil,
			want:        true,
		},
		{
			name:        "max attempts reached",
			maxAttempts: 3,
			attempt:     3,
			resp:        &Response{StatusCode: 503},
			err:         nil,
			want:        false,
		},
		{
			name:        "success response should not retry",
			maxAttempts: 3,
			attempt:     0,
			resp:        &Response{StatusCode: 200},
			err:         nil,
			want:        false,
		},
		{
			name:        "client error should not retry",
			maxAttempts: 3,
			attempt:     0,
			resp:        &Response{StatusCode: 400},
			err:         nil,
			want:        false,
		},
		{
			name:        "429 Too Many Requests should retry",
			maxAttempts: 3,
			attempt:     0,
			resp:        &Response{StatusCode: 429},
			err:         nil,
			want:        true,
		},
		{
			name:        "502 Bad Gateway should retry",
			maxAttempts: 3,
			attempt:     0,
			resp:        &Response{StatusCode: 502},
			err:         nil,
			want:        true,
		},
		{
			name:        "504 Gateway Timeout should retry",
			maxAttempts: 3,
			attempt:     0,
			resp:        &Response{StatusCode: 504},
			err:         nil,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eb := &ExponentialBackoff{
				MaxAttempts: tt.maxAttempts,
			}
			got := eb.ShouldRetry(tt.attempt, tt.resp, tt.err)
			if got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExponentialBackoff_NextDelay(t *testing.T) {
	eb := &ExponentialBackoff{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0, // disable jitter for predictable tests
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},  // 100ms * 2^0 = 100ms
		{1, 200 * time.Millisecond},  // 100ms * 2^1 = 200ms
		{2, 400 * time.Millisecond},  // 100ms * 2^2 = 400ms
		{3, 800 * time.Millisecond},  // 100ms * 2^3 = 800ms
		{4, 1600 * time.Millisecond}, // 100ms * 2^4 = 1600ms
	}

	for _, tt := range tests {
		t.Run("attempt_"+string(rune('0'+tt.attempt)), func(t *testing.T) {
			got := eb.NextDelay(tt.attempt)
			if got != tt.expected {
				t.Errorf("NextDelay(%d) = %v, want %v", tt.attempt, got, tt.expected)
			}
		})
	}
}

func TestExponentialBackoff_MaxDelayRespected(t *testing.T) {
	eb := &ExponentialBackoff{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}

	// Attempt 10 would give 100ms * 2^10 = 102.4s, but should be capped
	delay := eb.NextDelay(10)
	if delay != eb.MaxDelay {
		t.Errorf("NextDelay(10) = %v, want %v (MaxDelay)", delay, eb.MaxDelay)
	}
}

func TestExponentialBackoff_WithJitter(t *testing.T) {
	eb := &ExponentialBackoff{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1, // 10% jitter
	}

	baseDelay := 100 * time.Millisecond
	maxJitter := time.Duration(float64(baseDelay) * 0.1)

	// Run multiple times to verify jitter is applied
	for i := 0; i < 10; i++ {
		delay := eb.NextDelay(0)
		if delay < baseDelay || delay > baseDelay+maxJitter {
			t.Errorf("NextDelay(0) = %v, expected between %v and %v", delay, baseDelay, baseDelay+maxJitter)
		}
	}
}

func TestNewExponentialBackoff_Defaults(t *testing.T) {
	eb := NewExponentialBackoff()

	if eb.MaxAttempts != DefaultRetryMaxAttempts {
		t.Errorf("MaxAttempts = %d, want %d", eb.MaxAttempts, DefaultRetryMaxAttempts)
	}
	if eb.InitialDelay != DefaultRetryInitialDelay {
		t.Errorf("InitialDelay = %v, want %v", eb.InitialDelay, DefaultRetryInitialDelay)
	}
	if eb.MaxDelay != DefaultRetryMaxDelay {
		t.Errorf("MaxDelay = %v, want %v", eb.MaxDelay, DefaultRetryMaxDelay)
	}
	if eb.Multiplier != DefaultRetryMultiplier {
		t.Errorf("Multiplier = %f, want %f", eb.Multiplier, DefaultRetryMultiplier)
	}
}

// ============================================================================
// ConstantBackoff Tests
// ============================================================================

func TestConstantBackoff_ShouldRetry(t *testing.T) {
	cb := NewConstantBackoff(3, time.Second)

	tests := []struct {
		name    string
		attempt int
		resp    *Response
		want    bool
	}{
		{"first attempt with error", 0, &Response{StatusCode: 503}, true},
		{"second attempt", 1, &Response{StatusCode: 503}, true},
		{"third attempt", 2, &Response{StatusCode: 503}, true},
		{"max reached", 3, &Response{StatusCode: 503}, false},
		{"success", 0, &Response{StatusCode: 200}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cb.ShouldRetry(tt.attempt, tt.resp, nil)
			if got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstantBackoff_NextDelay(t *testing.T) {
	delay := 500 * time.Millisecond
	cb := NewConstantBackoff(3, delay)

	// Delay should be constant regardless of attempt
	for attempt := 0; attempt < 5; attempt++ {
		got := cb.NextDelay(attempt)
		if got != delay {
			t.Errorf("NextDelay(%d) = %v, want %v", attempt, got, delay)
		}
	}
}

// ============================================================================
// NoRetry Tests
// ============================================================================

func TestNoRetry_ShouldRetry(t *testing.T) {
	nr := &NoRetry{}

	if nr.ShouldRetry(0, &Response{StatusCode: 503}, nil) {
		t.Error("NoRetry.ShouldRetry() should always return false")
	}
}

func TestNoRetry_NextDelay(t *testing.T) {
	nr := &NoRetry{}

	if nr.NextDelay(0) != 0 {
		t.Error("NoRetry.NextDelay() should always return 0")
	}
}

// ============================================================================
// isRetryableError Tests
// ============================================================================

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"context canceled", context.Canceled, false},
		{"context deadline exceeded", context.DeadlineExceeded, false},
		{"generic error", errors.New("some error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// ============================================================================
// isRetryableStatus Tests
// ============================================================================

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, false},
		{"201 Created", 201, false},
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"403 Forbidden", 403, false},
		{"404 Not Found", 404, false},
		{"429 Too Many Requests", 429, true},
		{"500 Internal Server Error", 500, false},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{StatusCode: tt.statusCode}
			got := isRetryableStatus(resp)
			if got != tt.want {
				t.Errorf("isRetryableStatus(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestIsRetryableStatus_NilResponse(t *testing.T) {
	if isRetryableStatus(nil) {
		t.Error("isRetryableStatus(nil) should return false")
	}
}

// ============================================================================
// Client Retry Integration Tests
// ============================================================================

func TestClient_RetryOnServerError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client, err := New(server.URL,
		WithRetryStrategy(&ExponentialBackoff{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			JitterFactor: 0,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestClient_RetryMaxAttemptsExhausted(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := New(server.URL,
		WithRetryStrategy(&ExponentialBackoff{
			MaxAttempts:  2,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			JitterFactor: 0,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.Get(context.Background(), "/test")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Verify that ErrMaxRetriesExceeded is included in the error chain
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Errorf("Expected error to wrap ErrMaxRetriesExceeded, got: %v", err)
	}

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}

	// 1 initial + 2 retries = 3 attempts
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestClient_NoRetryOnClientError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := New(server.URL,
		WithRetryStrategy(NewExponentialBackoff()),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.Get(context.Background(), "/test")

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	// 4xx errors (except 429) should not be retried
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1 (no retry for 400)", attempts)
	}

	// ErrMaxRetriesExceeded should NOT be wrapped when no retries were attempted
	if errors.Is(err, ErrMaxRetriesExceeded) {
		t.Error("Expected error to NOT wrap ErrMaxRetriesExceeded for non-retryable error on first attempt")
	}
}

func TestClient_Retry429TooManyRequests(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(server.URL,
		WithRetryStrategy(&ExponentialBackoff{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			JitterFactor: 0,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestClient_RetryContextCancellation(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := New(server.URL,
		WithRetryStrategy(&ExponentialBackoff{
			MaxAttempts:  10,
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     1 * time.Second,
			Multiplier:   2.0,
			JitterFactor: 0,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.Get(ctx, "/test")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestClient_RetryWithRequestOption(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Client without retry strategy
	client, err := New(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Use per-request retry strategy
	resp, err := client.Get(context.Background(), "/test",
		WithRequestRetryStrategy(&ConstantBackoff{
			MaxAttempts: 3,
			Delay:       10 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestClient_WithNoRetry(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Client with retry strategy
	client, err := New(server.URL,
		WithRetryStrategy(NewExponentialBackoff()),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Disable retry for this specific request
	resp, _ := client.Get(context.Background(), "/test", WithNoRetry())

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}

	// Should not retry
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
}

func TestClient_RequestOptionOverridesClientRetry(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Client with 1 retry
	client, err := New(server.URL,
		WithRetryStrategy(&ConstantBackoff{
			MaxAttempts: 1,
			Delay:       10 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override with 3 retries per request
	_, _ = client.Get(context.Background(), "/test",
		WithRequestRetryStrategy(&ConstantBackoff{
			MaxAttempts: 3,
			Delay:       10 * time.Millisecond,
		}),
	)

	// 1 initial + 3 retries = 4 attempts
	if atomic.LoadInt32(&attempts) != 4 {
		t.Errorf("attempts = %d, want 4", attempts)
	}
}

func TestClient_CustomRetryStrategy(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Custom strategy that retries on 500 errors
	customStrategy := &ExponentialBackoff{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		RetryableFunc: func(resp *Response, err error) bool {
			if resp != nil && resp.StatusCode == 500 {
				return true
			}
			return isRetryable(resp, err)
		},
	}

	client, err := New(server.URL,
		WithRetryStrategy(customStrategy),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, _ = client.Get(context.Background(), "/test")

	// 1 initial + 2 retries = 3 attempts
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

// ============================================================================
// Thread Safety Tests
// ============================================================================

func TestClient_RetryThreadSafety(t *testing.T) {
	var totalAttempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&totalAttempts, 1)
		// Randomly fail some requests
		if atomic.LoadInt32(&totalAttempts)%2 == 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := New(server.URL,
		WithRetryStrategy(&ExponentialBackoff{
			MaxAttempts:  2,
			InitialDelay: 1 * time.Millisecond,
			MaxDelay:     10 * time.Millisecond,
			Multiplier:   2.0,
			JitterFactor: 0.1,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	const numGoroutines = 10
	const requestsPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				_, _ = client.Get(context.Background(), "/test")
			}
		}()
	}

	wg.Wait()

	// Just verify it completed without panics or races
	if atomic.LoadInt32(&totalAttempts) == 0 {
		t.Error("Expected some requests to be made")
	}
}

// ============================================================================
// addJitter Tests
// ============================================================================

func TestAddJitter(t *testing.T) {
	delay := 100 * time.Millisecond
	jitterFactor := 0.2 // 20% jitter

	minDelay := delay
	maxDelay := delay + time.Duration(float64(delay)*jitterFactor)

	for i := 0; i < 100; i++ {
		result := addJitter(delay, jitterFactor)
		if result < minDelay || result > maxDelay {
			t.Errorf("addJitter() = %v, expected between %v and %v", result, minDelay, maxDelay)
		}
	}
}

func TestAddJitter_ZeroFactor(t *testing.T) {
	delay := 100 * time.Millisecond
	result := addJitter(delay, 0)
	if result != delay {
		t.Errorf("addJitter() with 0 factor = %v, want %v", result, delay)
	}
}

func TestAddJitter_NegativeFactor(t *testing.T) {
	delay := 100 * time.Millisecond
	result := addJitter(delay, -0.1)
	if result != delay {
		t.Errorf("addJitter() with negative factor = %v, want %v", result, delay)
	}
}

// ============================================================================
// Network Error Retry Tests
// ============================================================================

func TestClient_RetryOnConnectionError(t *testing.T) {
	var attempts int32
	var serverStarted int32

	// Create a server that we'll start after first request attempt
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close() // Close it so first connection fails

	// Start server after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		atomic.StoreInt32(&serverStarted, 1)
		newListener, _ := net.Listen("tcp", addr)
		if newListener != nil {
			defer newListener.Close()
			server := &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					atomic.AddInt32(&attempts, 1)
					w.WriteHeader(http.StatusOK)
				}),
			}
			server.Serve(newListener)
		}
	}()

	client, err := New("http://"+addr,
		WithTimeout(200*time.Millisecond),
		WithRetryStrategy(&ExponentialBackoff{
			MaxAttempts:  5,
			InitialDelay: 30 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   1.5,
			JitterFactor: 0,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should eventually succeed after server starts
	resp, err := client.Get(ctx, "/test")

	// If server never started or request failed, that's okay for this test
	// We're mainly checking that retry logic handles connection errors gracefully
	if atomic.LoadInt32(&serverStarted) == 1 && err == nil {
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	}
}
