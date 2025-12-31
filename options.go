package clienthttp

import (
	"clienthttp/internal/client"
	"crypto/tls"
	"time"
)

// Option is a function that configures the HTTP client.
type Option = client.Option

// ============================================================================
// Timeout Options
// ============================================================================

// WithTimeout sets the total request timeout.
// This is the maximum time allowed for the entire request including
// connection, TLS handshake, sending request, and reading response.
// Default: 30 seconds
func WithTimeout(d time.Duration) Option {
	return client.WithTimeout(d)
}

// WithDialTimeout sets the maximum time allowed for establishing a connection.
// Default: 10 seconds
func WithDialTimeout(d time.Duration) Option {
	return client.WithDialTimeout(d)
}

// WithTLSHandshakeTimeout sets the maximum time allowed for the TLS handshake.
// Default: 10 seconds
func WithTLSHandshakeTimeout(d time.Duration) Option {
	return client.WithTLSHandshakeTimeout(d)
}

// WithResponseHeaderTimeout sets the maximum time to wait for a server's
// response headers after fully writing the request.
// Default: 10 seconds
func WithResponseHeaderTimeout(d time.Duration) Option {
	return client.WithResponseHeaderTimeout(d)
}

// ============================================================================
// Connection Pool Options
// ============================================================================

// WithMaxIdleConns sets the maximum number of idle (keep-alive) connections
// across all hosts. Zero means no limit.
// Default: 100
func WithMaxIdleConns(n int) Option {
	return client.WithMaxIdleConns(n)
}

// WithMaxIdleConnsPerHost sets the maximum idle (keep-alive) connections
// to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used.
// Default: 10
func WithMaxIdleConnsPerHost(n int) Option {
	return client.WithMaxIdleConnsPerHost(n)
}

// WithMaxConnsPerHost optionally limits the total number of connections
// per host, including connections in the dialing, active, and idle states.
// Zero means no limit.
// Default: 100
func WithMaxConnsPerHost(n int) Option {
	return client.WithMaxConnsPerHost(n)
}

// WithIdleConnTimeout sets the maximum amount of time an idle (keep-alive)
// connection will remain idle before closing itself.
// Zero means no limit.
// Default: 90 seconds
func WithIdleConnTimeout(d time.Duration) Option {
	return client.WithIdleConnTimeout(d)
}

// ============================================================================
// TLS/SSL Options
// ============================================================================

// WithTLSConfig sets a custom TLS configuration.
// This allows full control over TLS settings. When provided, other TLS options
// (WithInsecureSkipVerify, WithRootCA, etc.) are ignored.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return client.WithTLSConfig(tlsConfig)
}

// WithInsecureSkipVerify disables TLS certificate verification.
// WARNING: This should only be used for development/testing purposes.
// Using this in production is a security risk.
func WithInsecureSkipVerify() Option {
	return client.WithInsecureSkipVerify()
}

// WithTLSMinVersion sets the minimum TLS version that is acceptable.
// Use tls.VersionTLS10, tls.VersionTLS11, tls.VersionTLS12, or tls.VersionTLS13.
// Default: tls.VersionTLS12
func WithTLSMinVersion(version uint16) Option {
	return client.WithTLSMinVersion(version)
}

// WithTLSMaxVersion sets the maximum TLS version that is acceptable.
// Use tls.VersionTLS10, tls.VersionTLS11, tls.VersionTLS12, or tls.VersionTLS13.
// Default: 0 (uses the maximum version available)
func WithTLSMaxVersion(version uint16) Option {
	return client.WithTLSMaxVersion(version)
}

// WithRootCA adds a custom root CA certificate from a PEM file.
// This is useful for connecting to servers with certificates signed by
// a private/corporate CA.
func WithRootCA(caFile string) Option {
	return client.WithRootCA(caFile)
}

// WithRootCAFromPEM adds a custom root CA certificate from PEM-encoded bytes.
// This is useful for connecting to servers with certificates signed by
// a private/corporate CA.
func WithRootCAFromPEM(caPEM []byte) Option {
	return client.WithRootCAFromPEM(caPEM)
}

// WithClientCertificate loads a client certificate and key from PEM files
// for mutual TLS (mTLS) authentication.
func WithClientCertificate(certFile, keyFile string) Option {
	return client.WithClientCertificate(certFile, keyFile)
}

// WithClientCertificateFromPEM loads a client certificate and key from
// PEM-encoded bytes for mutual TLS (mTLS) authentication.
func WithClientCertificateFromPEM(certPEM, keyPEM []byte) Option {
	return client.WithClientCertificateFromPEM(certPEM, keyPEM)
}

// ============================================================================
// Default Values (exported for reference)
// ============================================================================

const (
	// JsonContentType is the default content type for requests.
	JsonContentType = "application/json"

	// DefaultTimeout is the default total request timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultDialTimeout is the default connection establishment timeout.
	DefaultDialTimeout = 10 * time.Second

	// DefaultTLSHandshakeTimeout is the default TLS handshake timeout.
	DefaultTLSHandshakeTimeout = 10 * time.Second

	// DefaultResponseHeaderTimeout is the default response header timeout.
	DefaultResponseHeaderTimeout = 10 * time.Second

	// DefaultMaxIdleConns is the default maximum number of idle connections.
	DefaultMaxIdleConns = 100

	// DefaultMaxIdleConnsPerHost is the default maximum idle connections per host.
	DefaultMaxIdleConnsPerHost = 10

	// DefaultMaxConnsPerHost is the default maximum connections per host.
	DefaultMaxConnsPerHost = 100

	// DefaultIdleConnTimeout is the default idle connection timeout.
	DefaultIdleConnTimeout = 90 * time.Second

	// DefaultTLSMinVersion is the default minimum TLS version (TLS 1.2).
	DefaultTLSMinVersion uint16 = tls.VersionTLS12

	// DefaultTLSMaxVersion is the default maximum TLS version (0 = use maximum available).
	DefaultTLSMaxVersion uint16 = 0
)
