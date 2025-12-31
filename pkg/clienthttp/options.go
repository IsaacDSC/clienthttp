package clienthttp

import "time"

// ============================================================================
// Timeout Options
// ============================================================================

// WithTimeout sets the total request timeout.
// This is the maximum time allowed for the entire request including
// connection, TLS handshake, sending request, and reading response.
// Default: 30 seconds
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

// WithDialTimeout sets the maximum time allowed for establishing a connection.
// Default: 10 seconds
func WithDialTimeout(d time.Duration) Option {
	return func(c *config) {
		c.dialTimeout = d
	}
}

// WithTLSHandshakeTimeout sets the maximum time allowed for the TLS handshake.
// Default: 10 seconds
func WithTLSHandshakeTimeout(d time.Duration) Option {
	return func(c *config) {
		c.tlsHandshakeTimeout = d
	}
}

// WithResponseHeaderTimeout sets the maximum time to wait for a server's
// response headers after fully writing the request.
// Default: 10 seconds
func WithResponseHeaderTimeout(d time.Duration) Option {
	return func(c *config) {
		c.responseHeaderTimeout = d
	}
}

// ============================================================================
// Connection Pool Options
// ============================================================================

// WithMaxIdleConns sets the maximum number of idle (keep-alive) connections
// across all hosts. Zero means no limit.
// Default: 100
func WithMaxIdleConns(n int) Option {
	return func(c *config) {
		c.transport.maxIdleConns = n
	}
}

// WithMaxIdleConnsPerHost sets the maximum idle (keep-alive) connections
// to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used.
// Default: 10
func WithMaxIdleConnsPerHost(n int) Option {
	return func(c *config) {
		c.transport.maxIdleConnsPerHost = n
	}
}

// WithMaxConnsPerHost optionally limits the total number of connections
// per host, including connections in the dialing, active, and idle states.
// Zero means no limit.
// Default: 100
func WithMaxConnsPerHost(n int) Option {
	return func(c *config) {
		c.transport.maxConnsPerHost = n
	}
}

// WithIdleConnTimeout sets the maximum amount of time an idle (keep-alive)
// connection will remain idle before closing itself.
// Zero means no limit.
// Default: 90 seconds
func WithIdleConnTimeout(d time.Duration) Option {
	return func(c *config) {
		c.transport.idleConnTimeout = d
	}
}

