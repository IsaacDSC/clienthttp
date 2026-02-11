package clienthttp

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"time"
)

// ============================================================================
// Retry Strategy Interface
// ============================================================================

// RetryStrategy defines how retries should be performed.
// Implement this interface to create custom retry strategies.
type RetryStrategy interface {
	// ShouldRetry determines if the request should be retried.
	// attempt is 0-indexed (0 = first attempt, 1 = first retry, etc.)
	ShouldRetry(attempt int, resp *Response, err error) bool

	// NextDelay returns the delay before the next attempt.
	// attempt is 0-indexed.
	NextDelay(attempt int) time.Duration
}

// ============================================================================
// Exponential Backoff Strategy
// ============================================================================

// Default retry values
const (
	// DefaultRetryMaxAttempts is the default maximum number of retry attempts.
	DefaultRetryMaxAttempts = 3

	// DefaultRetryInitialDelay is the default initial delay between retries.
	DefaultRetryInitialDelay = 100 * time.Millisecond

	// DefaultRetryMaxDelay is the default maximum delay between retries.
	DefaultRetryMaxDelay = 30 * time.Second

	// DefaultRetryMultiplier is the default multiplier for exponential backoff.
	DefaultRetryMultiplier = 2.0

	// DefaultJitterFactor is the default jitter factor (0.0 to 1.0).
	DefaultJitterFactor = 0.1
)

// ExponentialBackoff implements RetryStrategy with exponential backoff and jitter.
// The delay between retries grows exponentially: initialDelay * multiplier^attempt
type ExponentialBackoff struct {
	// MaxAttempts is the maximum number of retry attempts (not including the initial request).
	// A value of 3 means: 1 initial request + up to 3 retries = 4 total attempts.
	MaxAttempts int

	// InitialDelay is the delay before the first retry.
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry.
	Multiplier float64

	// JitterFactor adds randomness to the delay (0.0 to 1.0).
	// A value of 0.1 adds up to 10% random jitter.
	JitterFactor float64

	// RetryableFunc is an optional custom function to determine if a request should be retried.
	// If nil, the default isRetryable function is used.
	RetryableFunc func(resp *Response, err error) bool
}

// NewExponentialBackoff creates an ExponentialBackoff strategy with default values.
func NewExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		MaxAttempts:  DefaultRetryMaxAttempts,
		InitialDelay: DefaultRetryInitialDelay,
		MaxDelay:     DefaultRetryMaxDelay,
		Multiplier:   DefaultRetryMultiplier,
		JitterFactor: DefaultJitterFactor,
	}
}

// ShouldRetry determines if the request should be retried based on the attempt count,
// response, and error.
func (e *ExponentialBackoff) ShouldRetry(attempt int, resp *Response, err error) bool {
	if attempt >= e.MaxAttempts {
		return false
	}

	if e.RetryableFunc != nil {
		return e.RetryableFunc(resp, err)
	}

	return isRetryable(resp, err)
}

// NextDelay calculates the delay before the next retry attempt using exponential backoff
// with optional jitter.
func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	delay := time.Duration(float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt)))

	if delay > e.MaxDelay {
		delay = e.MaxDelay
	}

	if e.JitterFactor > 0 {
		delay = addJitter(delay, e.JitterFactor)
	}

	return delay
}

// ============================================================================
// Constant Backoff Strategy
// ============================================================================

// ConstantBackoff implements RetryStrategy with a constant delay between retries.
type ConstantBackoff struct {
	// MaxAttempts is the maximum number of retry attempts (not including the initial request).
	MaxAttempts int

	// Delay is the constant delay between retries.
	Delay time.Duration

	// RetryableFunc is an optional custom function to determine if a request should be retried.
	// If nil, the default isRetryable function is used.
	RetryableFunc func(resp *Response, err error) bool
}

// NewConstantBackoff creates a ConstantBackoff strategy with the given max attempts and delay.
func NewConstantBackoff(maxAttempts int, delay time.Duration) *ConstantBackoff {
	return &ConstantBackoff{
		MaxAttempts: maxAttempts,
		Delay:       delay,
	}
}

// ShouldRetry determines if the request should be retried.
func (c *ConstantBackoff) ShouldRetry(attempt int, resp *Response, err error) bool {
	if attempt >= c.MaxAttempts {
		return false
	}

	if c.RetryableFunc != nil {
		return c.RetryableFunc(resp, err)
	}

	return isRetryable(resp, err)
}

// NextDelay returns the constant delay.
func (c *ConstantBackoff) NextDelay(attempt int) time.Duration {
	return c.Delay
}

// ============================================================================
// No Retry Strategy
// ============================================================================

// NoRetry is a RetryStrategy that never retries.
type NoRetry struct{}

// ShouldRetry always returns false.
func (n *NoRetry) ShouldRetry(attempt int, resp *Response, err error) bool {
	return false
}

// NextDelay returns 0.
func (n *NoRetry) NextDelay(attempt int) time.Duration {
	return 0
}

// ============================================================================
// Helper Functions
// ============================================================================

// isRetryable determines if a request should be retried based on the response and error.
// This is the default retry logic used when no custom RetryableFunc is provided.
func isRetryable(resp *Response, err error) bool {
	if isRetryableError(err) {
		return true
	}

	if isRetryableStatus(resp) {
		return true
	}

	return false
}

// isRetryableError checks if an error is retryable.
// Network errors and timeouts are considered retryable.
// Context cancellation and deadline exceeded are NOT retryable.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context errors should not be retried
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for net.Error (includes timeout and temporary errors)
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Timeout errors are retryable
		if netErr.Timeout() {
			return true
		}
	}

	// Check for connection refused, reset, etc.
	var opErr *net.OpError
	return errors.As(err, &opErr)
}

// isRetryableStatus checks if an HTTP status code indicates a retryable condition.
// The following status codes are considered retryable:
//   - 429 Too Many Requests
//   - 502 Bad Gateway
//   - 503 Service Unavailable
//   - 504 Gateway Timeout
func isRetryableStatus(resp *Response) bool {
	if resp == nil {
		return false
	}

	switch resp.StatusCode {
	case 429, // Too Many Requests
		502, // Bad Gateway
		503, // Service Unavailable
		504: // Gateway Timeout
		return true
	}

	return false
}

// addJitter adds random jitter to a delay duration.
// jitterFactor should be between 0.0 and 1.0.
// For example, a jitterFactor of 0.1 adds up to 10% random jitter.
func addJitter(delay time.Duration, jitterFactor float64) time.Duration {
	if jitterFactor <= 0 {
		return delay
	}

	jitter := time.Duration(rand.Float64() * jitterFactor * float64(delay))
	return delay + jitter
}
