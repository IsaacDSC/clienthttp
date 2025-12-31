package clienthttp

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"time"
)

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

// ============================================================================
// TLS/SSL Options
// ============================================================================

// WithTLSConfig sets a custom TLS configuration.
// This allows full control over TLS settings. When provided, other TLS options
// (WithInsecureSkipVerify, WithRootCA, etc.) are ignored.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *config) {
		c.tls.enabled = true
		c.tls.customConfig = tlsConfig
	}
}

// WithInsecureSkipVerify disables TLS certificate verification.
// WARNING: This should only be used for development/testing purposes.
// Using this in production is a security risk.
func WithInsecureSkipVerify() Option {
	return func(c *config) {
		c.tls.enabled = true
		c.tls.insecureSkipVerify = true
	}
}

// WithTLSMinVersion sets the minimum TLS version that is acceptable.
// Use tls.VersionTLS10, tls.VersionTLS11, tls.VersionTLS12, or tls.VersionTLS13.
// Default: tls.VersionTLS12
func WithTLSMinVersion(version uint16) Option {
	return func(c *config) {
		c.tls.enabled = true
		c.tls.minVersion = version
	}
}

// WithTLSMaxVersion sets the maximum TLS version that is acceptable.
// Use tls.VersionTLS10, tls.VersionTLS11, tls.VersionTLS12, or tls.VersionTLS13.
// Default: 0 (uses the maximum version available)
func WithTLSMaxVersion(version uint16) Option {
	return func(c *config) {
		c.tls.enabled = true
		c.tls.maxVersion = version
	}
}

// WithRootCA adds a custom root CA certificate from a PEM file.
// This is useful for connecting to servers with certificates signed by
// a private/corporate CA.
func WithRootCA(caFile string) Option {
	return func(c *config) {
		caPEM, err := os.ReadFile(caFile)
		if err != nil {
			// Store error state - will be handled during transport creation
			c.tls.enabled = true
			return
		}

		if c.tls.rootCAs == nil {
			c.tls.rootCAs = x509.NewCertPool()
		}

		c.tls.rootCAs.AppendCertsFromPEM(caPEM)
		c.tls.enabled = true
	}
}

// WithRootCAFromPEM adds a custom root CA certificate from PEM-encoded bytes.
// This is useful for connecting to servers with certificates signed by
// a private/corporate CA.
func WithRootCAFromPEM(caPEM []byte) Option {
	return func(c *config) {
		if c.tls.rootCAs == nil {
			c.tls.rootCAs = x509.NewCertPool()
		}

		c.tls.rootCAs.AppendCertsFromPEM(caPEM)
		c.tls.enabled = true
	}
}

// WithClientCertificate loads a client certificate and key from PEM files
// for mutual TLS (mTLS) authentication.
func WithClientCertificate(certFile, keyFile string) Option {
	return func(c *config) {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			// Store error state - will be handled during transport creation
			c.tls.enabled = true
			return
		}

		c.tls.certificates = append(c.tls.certificates, cert)
		c.tls.enabled = true
	}
}

// WithClientCertificateFromPEM loads a client certificate and key from
// PEM-encoded bytes for mutual TLS (mTLS) authentication.
func WithClientCertificateFromPEM(certPEM, keyPEM []byte) Option {
	return func(c *config) {
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			// Store error state - will be handled during transport creation
			c.tls.enabled = true
			return
		}

		c.tls.certificates = append(c.tls.certificates, cert)
		c.tls.enabled = true
	}
}

