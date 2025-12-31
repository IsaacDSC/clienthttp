package clienthttp

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"
)

type config struct {
	authCallback         func(r *http.Request)
	cookies              []http.Cookie
	contentType          string
	enabledCorrelationID bool

	// Timeout configuration
	timeout               time.Duration // Timeout total da requisição
	dialTimeout           time.Duration // Timeout para estabelecer conexão
	tlsHandshakeTimeout   time.Duration // Timeout para handshake TLS
	responseHeaderTimeout time.Duration // Timeout para receber headers

	// Connection pool configuration
	transport transportConfig

	// TLS configuration
	tls tlsConfig
}

type transportConfig struct {
	maxIdleConns        int
	maxIdleConnsPerHost int
	maxConnsPerHost     int
	idleConnTimeout     time.Duration
}

// tlsConfig holds TLS/SSL configuration options
type tlsConfig struct {
	enabled            bool
	customConfig       *tls.Config      // User-provided full TLS config
	insecureSkipVerify bool             // Skip certificate verification (dev only)
	rootCAs            *x509.CertPool   // Custom CA certificates
	certificates       []tls.Certificate // Client certificates for mTLS
	minVersion         uint16           // Minimum TLS version
	maxVersion         uint16           // Maximum TLS version
}

type Option func(*config)

func newConfig(opts ...Option) *config {
	c := new(config)
	c.defaults()

	for i := range opts {
		opts[i](c)
	}

	return c
}

const (
	JsonContentType = "application/json"

	// Default timeout values
	DefaultTimeout               = 30 * time.Second
	DefaultDialTimeout           = 10 * time.Second
	DefaultTLSHandshakeTimeout   = 10 * time.Second
	DefaultResponseHeaderTimeout = 10 * time.Second

	// Default connection pool values
	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 10
	DefaultMaxConnsPerHost     = 100
	DefaultIdleConnTimeout     = 90 * time.Second

	// Default TLS values
	DefaultTLSMinVersion uint16 = tls.VersionTLS12
	DefaultTLSMaxVersion uint16 = 0 // 0 means use the maximum version available
)

func (c *config) defaults() {
	c.authCallback = nil
	c.contentType = JsonContentType
	c.cookies = make([]http.Cookie, 0)

	// Timeout defaults
	c.timeout = DefaultTimeout
	c.dialTimeout = DefaultDialTimeout
	c.tlsHandshakeTimeout = DefaultTLSHandshakeTimeout
	c.responseHeaderTimeout = DefaultResponseHeaderTimeout

	// Connection pool defaults
	c.transport = transportConfig{
		maxIdleConns:        DefaultMaxIdleConns,
		maxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		maxConnsPerHost:     DefaultMaxConnsPerHost,
		idleConnTimeout:     DefaultIdleConnTimeout,
	}

	// TLS defaults
	c.tls = tlsConfig{
		enabled:            false,
		insecureSkipVerify: false,
		minVersion:         DefaultTLSMinVersion,
		maxVersion:         DefaultTLSMaxVersion,
	}
}
