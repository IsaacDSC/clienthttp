package client

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"
)

// Config holds the client configuration options.
type Config struct {
	AuthCallback         func(r *http.Request)
	Cookies              []http.Cookie
	ContentType          string
	EnabledCorrelationID bool

	// Timeout configuration
	Timeout               time.Duration // Timeout total da requisição
	DialTimeout           time.Duration // Timeout para estabelecer conexão
	TLSHandshakeTimeout   time.Duration // Timeout para handshake TLS
	ResponseHeaderTimeout time.Duration // Timeout para receber headers

	// Connection pool configuration
	Transport TransportConfig

	// TLS configuration
	TLS TLSConfig
}

// TransportConfig holds connection pool configuration options.
type TransportConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
}

// TLSConfig holds TLS/SSL configuration options.
type TLSConfig struct {
	Enabled            bool
	CustomConfig       *tls.Config       // User-provided full TLS config
	InsecureSkipVerify bool              // Skip certificate verification (dev only)
	RootCAs            *x509.CertPool    // Custom CA certificates
	Certificates       []tls.Certificate // Client certificates for mTLS
	MinVersion         uint16            // Minimum TLS version
	MaxVersion         uint16            // Maximum TLS version
}

// Option is a function that configures the client.
type Option func(*Config)

func newConfig(opts ...Option) *Config {
	c := new(Config)
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

func (c *Config) defaults() {
	c.AuthCallback = nil
	c.ContentType = JsonContentType
	c.Cookies = make([]http.Cookie, 0)

	// Timeout defaults
	c.Timeout = DefaultTimeout
	c.DialTimeout = DefaultDialTimeout
	c.TLSHandshakeTimeout = DefaultTLSHandshakeTimeout
	c.ResponseHeaderTimeout = DefaultResponseHeaderTimeout

	// Connection pool defaults
	c.Transport = TransportConfig{
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		MaxConnsPerHost:     DefaultMaxConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,
	}

	// TLS defaults
	c.TLS = TLSConfig{
		Enabled:            false,
		InsecureSkipVerify: false,
		MinVersion:         DefaultTLSMinVersion,
		MaxVersion:         DefaultTLSMaxVersion,
	}
}
