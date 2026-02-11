package clienthttp

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"
)

// ============================================================================
// Client Configuration
// ============================================================================

// config holds the client configuration options.
type config struct {
	// Callbacks and adapters
	authCallback  func(r *http.Request)
	auditor       Auditor
	correlationFn CorrelationIDFunc

	// Request defaults
	contentType string
	cookies     []http.Cookie

	// Timeout configuration
	timeout               time.Duration
	dialTimeout           time.Duration
	tlsHandshakeTimeout   time.Duration
	responseHeaderTimeout time.Duration

	// Connection pool configuration
	transport transportConfig

	// TLS configuration
	tls tlsConfig

	// Retry configuration
	retryStrategy RetryStrategy
}

// transportConfig holds connection pool configuration options.
type transportConfig struct {
	maxIdleConns        int
	maxIdleConnsPerHost int
	maxConnsPerHost     int
	idleConnTimeout     time.Duration
}

// tlsConfig holds TLS/SSL configuration options.
type tlsConfig struct {
	enabled            bool
	customConfig       *tls.Config
	insecureSkipVerify bool
	rootCAs            *x509.CertPool
	certificates       []tls.Certificate
	minVersion         uint16
	maxVersion         uint16
}

// Option configures the HTTP client.
type Option func(*config)

// Default values
const (
	// ContentTypeJSON is the default content type for requests.
	ContentTypeJSON = "application/json"

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
)

func newConfig(opts ...Option) *config {
	c := &config{
		contentType:           ContentTypeJSON,
		cookies:               make([]http.Cookie, 0),
		timeout:               DefaultTimeout,
		dialTimeout:           DefaultDialTimeout,
		tlsHandshakeTimeout:   DefaultTLSHandshakeTimeout,
		responseHeaderTimeout: DefaultResponseHeaderTimeout,
		transport: transportConfig{
			maxIdleConns:        DefaultMaxIdleConns,
			maxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
			maxConnsPerHost:     DefaultMaxConnsPerHost,
			idleConnTimeout:     DefaultIdleConnTimeout,
		},
		tls: tlsConfig{
			minVersion: DefaultTLSMinVersion,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ============================================================================
// Client Options - Timeout
// ============================================================================

// WithTimeout sets the total request timeout.
// Default: 30 seconds
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

// WithDialTimeout sets the maximum time for establishing a connection.
// Default: 10 seconds
func WithDialTimeout(d time.Duration) Option {
	return func(c *config) {
		c.dialTimeout = d
	}
}

// WithTLSHandshakeTimeout sets the maximum time for the TLS handshake.
// Default: 10 seconds
func WithTLSHandshakeTimeout(d time.Duration) Option {
	return func(c *config) {
		c.tlsHandshakeTimeout = d
	}
}

// WithResponseHeaderTimeout sets the maximum time to wait for response headers.
// Default: 10 seconds
func WithResponseHeaderTimeout(d time.Duration) Option {
	return func(c *config) {
		c.responseHeaderTimeout = d
	}
}

// ============================================================================
// Client Options - Connection Pool
// ============================================================================

// WithMaxIdleConns sets the maximum number of idle connections across all hosts.
// Default: 100
func WithMaxIdleConns(n int) Option {
	return func(c *config) {
		c.transport.maxIdleConns = n
	}
}

// WithMaxIdleConnsPerHost sets the maximum idle connections per host.
// Default: 10
func WithMaxIdleConnsPerHost(n int) Option {
	return func(c *config) {
		c.transport.maxIdleConnsPerHost = n
	}
}

// WithMaxConnsPerHost limits the total connections per host.
// Default: 100
func WithMaxConnsPerHost(n int) Option {
	return func(c *config) {
		c.transport.maxConnsPerHost = n
	}
}

// WithIdleConnTimeout sets the maximum time an idle connection remains open.
// Default: 90 seconds
func WithIdleConnTimeout(d time.Duration) Option {
	return func(c *config) {
		c.transport.idleConnTimeout = d
	}
}

// ============================================================================
// Client Options - TLS
// ============================================================================

// WithTLSConfig sets a custom TLS configuration.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *config) {
		c.tls.enabled = true
		c.tls.customConfig = tlsConfig
	}
}

// WithInsecureSkipVerify disables TLS certificate verification.
// WARNING: Only use for development/testing.
func WithInsecureSkipVerify() Option {
	return func(c *config) {
		c.tls.enabled = true
		c.tls.insecureSkipVerify = true
	}
}

// WithTLSMinVersion sets the minimum acceptable TLS version.
// Default: tls.VersionTLS12
func WithTLSMinVersion(version uint16) Option {
	return func(c *config) {
		c.tls.enabled = true
		c.tls.minVersion = version
	}
}

// WithTLSMaxVersion sets the maximum acceptable TLS version.
func WithTLSMaxVersion(version uint16) Option {
	return func(c *config) {
		c.tls.enabled = true
		c.tls.maxVersion = version
	}
}

// WithRootCA adds a custom root CA certificate from a PEM file.
func WithRootCA(caFile string) Option {
	return func(c *config) {
		caPEM, err := os.ReadFile(caFile)
		if err != nil {
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

// WithRootCAFromPEM adds a custom root CA certificate from PEM bytes.
func WithRootCAFromPEM(caPEM []byte) Option {
	return func(c *config) {
		if c.tls.rootCAs == nil {
			c.tls.rootCAs = x509.NewCertPool()
		}
		c.tls.rootCAs.AppendCertsFromPEM(caPEM)
		c.tls.enabled = true
	}
}

// WithClientCertificate loads a client certificate for mTLS from PEM files.
func WithClientCertificate(certFile, keyFile string) Option {
	return func(c *config) {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			c.tls.enabled = true
			return
		}
		c.tls.certificates = append(c.tls.certificates, cert)
		c.tls.enabled = true
	}
}

// WithClientCertificateFromPEM loads a client certificate for mTLS from PEM bytes.
func WithClientCertificateFromPEM(certPEM, keyPEM []byte) Option {
	return func(c *config) {
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			c.tls.enabled = true
			return
		}
		c.tls.certificates = append(c.tls.certificates, cert)
		c.tls.enabled = true
	}
}

// ============================================================================
// Client Options - Adapters
// ============================================================================

// WithAuditor sets an auditor for logging requests and responses.
func WithAuditor(a Auditor) Option {
	return func(c *config) {
		c.auditor = a
	}
}

// WithCorrelationID sets a function to extract correlation IDs from context.
func WithCorrelationID(fn CorrelationIDFunc) Option {
	return func(c *config) {
		c.correlationFn = fn
	}
}

// WithAuthCallback sets a callback to add authentication to requests.
func WithAuthCallback(fn func(r *http.Request)) Option {
	return func(c *config) {
		c.authCallback = fn
	}
}

// WithContentType sets the default content type for requests.
// Default: "application/json"
func WithContentType(contentType string) Option {
	return func(c *config) {
		c.contentType = contentType
	}
}

// WithCookies sets default cookies to be sent with every request.
func WithCookies(cookies ...http.Cookie) Option {
	return func(c *config) {
		c.cookies = append(c.cookies, cookies...)
	}
}

// ============================================================================
// Request Options (per-request)
// ============================================================================

// requestConfig holds per-request configuration.
type requestConfig struct {
	headers       map[string]string
	queryParams   map[string]string
	retryStrategy RetryStrategy
}

// RequestOption configures a single HTTP request.
type RequestOption func(*requestConfig)

func newRequestConfig(opts ...RequestOption) *requestConfig {
	rc := &requestConfig{
		headers:     make(map[string]string),
		queryParams: make(map[string]string),
	}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}

// WithQuery adds a query parameter to the request.
func WithQuery(key, value string) RequestOption {
	return func(rc *requestConfig) {
		rc.queryParams[key] = value
	}
}

// WithQueries adds multiple query parameters to the request.
func WithQueries(params map[string]string) RequestOption {
	return func(rc *requestConfig) {
		for k, v := range params {
			rc.queryParams[k] = v
		}
	}
}

// WithHeader adds a header to the request.
func WithHeader(key, value string) RequestOption {
	return func(rc *requestConfig) {
		rc.headers[key] = value
	}
}

// WithHeaders adds multiple headers to the request.
func WithHeaders(headers map[string]string) RequestOption {
	return func(rc *requestConfig) {
		for k, v := range headers {
			rc.headers[k] = v
		}
	}
}

// WithBasicAuth adds Basic Authentication to the request.
func WithBasicAuth(username, password string) RequestOption {
	return func(rc *requestConfig) {
		// Basic auth will be set directly on the request
		rc.headers["_basic_auth_user"] = username
		rc.headers["_basic_auth_pass"] = password
	}
}

// WithBearerToken adds a Bearer token to the Authorization header.
func WithBearerToken(token string) RequestOption {
	return func(rc *requestConfig) {
		rc.headers["Authorization"] = "Bearer " + token
	}
}

// ============================================================================
// Retry Options
// ============================================================================

// WithRetryStrategy sets the retry strategy for the client.
// All requests made with this client will use the provided strategy.
// Use WithRequestRetryStrategy to override for specific requests.
//
// Example:
//
//	client, _ := clienthttp.New("https://api.example.com",
//	    clienthttp.WithRetryStrategy(clienthttp.NewExponentialBackoff()),
//	)
func WithRetryStrategy(strategy RetryStrategy) Option {
	return func(c *config) {
		c.retryStrategy = strategy
	}
}

// WithRequestRetryStrategy sets the retry strategy for a specific request,
// overriding the client-level strategy.
//
// Example:
//
//	resp, err := client.Get(ctx, "/api/data",
//	    clienthttp.WithRequestRetryStrategy(clienthttp.NewConstantBackoff(5, time.Second)),
//	)
func WithRequestRetryStrategy(strategy RetryStrategy) RequestOption {
	return func(rc *requestConfig) {
		rc.retryStrategy = strategy
	}
}

// WithNoRetry disables retry for a specific request.
// This is a convenience function equivalent to WithRequestRetryStrategy(&NoRetry{}).
func WithNoRetry() RequestOption {
	return func(rc *requestConfig) {
		rc.retryStrategy = &NoRetry{}
	}
}
