package clienthttp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateSelfSignedCertForTest generates a self-signed certificate and key for testing
func generateSelfSignedCertForTest(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	// Generate private key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-client",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Encode certificate to PEM
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Encode private key to PEM
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return certPEM, keyPEM
}

// ============================================================================
// Timeout Options Tests
// ============================================================================

func TestWithTimeout_SetsValue(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "set 5 seconds timeout",
			timeout:  5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "set 1 minute timeout",
			timeout:  1 * time.Minute,
			expected: 1 * time.Minute,
		},
		{
			name:     "set zero timeout (no limit)",
			timeout:  0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig(WithTimeout(tt.timeout))
			if cfg.timeout != tt.expected {
				t.Errorf("WithTimeout() = %v, want %v", cfg.timeout, tt.expected)
			}
		})
	}
}

func TestWithDialTimeout_SetsValue(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "set 5 seconds dial timeout",
			timeout:  5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "set 30 seconds dial timeout",
			timeout:  30 * time.Second,
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig(WithDialTimeout(tt.timeout))
			if cfg.dialTimeout != tt.expected {
				t.Errorf("WithDialTimeout() = %v, want %v", cfg.dialTimeout, tt.expected)
			}
		})
	}
}

func TestWithTLSHandshakeTimeout_SetsValue(t *testing.T) {
	timeout := 15 * time.Second
	cfg := newConfig(WithTLSHandshakeTimeout(timeout))

	if cfg.tlsHandshakeTimeout != timeout {
		t.Errorf("WithTLSHandshakeTimeout() = %v, want %v", cfg.tlsHandshakeTimeout, timeout)
	}
}

func TestWithResponseHeaderTimeout_SetsValue(t *testing.T) {
	timeout := 20 * time.Second
	cfg := newConfig(WithResponseHeaderTimeout(timeout))

	if cfg.responseHeaderTimeout != timeout {
		t.Errorf("WithResponseHeaderTimeout() = %v, want %v", cfg.responseHeaderTimeout, timeout)
	}
}

func TestTimeout_DefaultValues(t *testing.T) {
	cfg := newConfig()

	tests := []struct {
		name     string
		got      time.Duration
		expected time.Duration
	}{
		{
			name:     "default timeout",
			got:      cfg.timeout,
			expected: DefaultTimeout,
		},
		{
			name:     "default dial timeout",
			got:      cfg.dialTimeout,
			expected: DefaultDialTimeout,
		},
		{
			name:     "default TLS handshake timeout",
			got:      cfg.tlsHandshakeTimeout,
			expected: DefaultTLSHandshakeTimeout,
		},
		{
			name:     "default response header timeout",
			got:      cfg.responseHeaderTimeout,
			expected: DefaultResponseHeaderTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// ============================================================================
// Connection Pool Options Tests
// ============================================================================

func TestWithMaxIdleConns_SetsValue(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{
			name:     "set 50 max idle conns",
			value:    50,
			expected: 50,
		},
		{
			name:     "set 200 max idle conns",
			value:    200,
			expected: 200,
		},
		{
			name:     "set zero (no limit)",
			value:    0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig(WithMaxIdleConns(tt.value))
			if cfg.transport.maxIdleConns != tt.expected {
				t.Errorf("WithMaxIdleConns() = %v, want %v", cfg.transport.maxIdleConns, tt.expected)
			}
		})
	}
}

func TestWithMaxIdleConnsPerHost_SetsValue(t *testing.T) {
	value := 20
	cfg := newConfig(WithMaxIdleConnsPerHost(value))

	if cfg.transport.maxIdleConnsPerHost != value {
		t.Errorf("WithMaxIdleConnsPerHost() = %v, want %v", cfg.transport.maxIdleConnsPerHost, value)
	}
}

func TestWithMaxConnsPerHost_SetsValue(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{
			name:     "set 50 max conns per host",
			value:    50,
			expected: 50,
		},
		{
			name:     "set 200 max conns per host",
			value:    200,
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig(WithMaxConnsPerHost(tt.value))
			if cfg.transport.maxConnsPerHost != tt.expected {
				t.Errorf("WithMaxConnsPerHost() = %v, want %v", cfg.transport.maxConnsPerHost, tt.expected)
			}
		})
	}
}

func TestWithIdleConnTimeout_SetsValue(t *testing.T) {
	timeout := 60 * time.Second
	cfg := newConfig(WithIdleConnTimeout(timeout))

	if cfg.transport.idleConnTimeout != timeout {
		t.Errorf("WithIdleConnTimeout() = %v, want %v", cfg.transport.idleConnTimeout, timeout)
	}
}

func TestConnectionPool_DefaultValues(t *testing.T) {
	cfg := newConfig()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{
			name:     "default max idle conns",
			got:      cfg.transport.maxIdleConns,
			expected: DefaultMaxIdleConns,
		},
		{
			name:     "default max idle conns per host",
			got:      cfg.transport.maxIdleConnsPerHost,
			expected: DefaultMaxIdleConnsPerHost,
		},
		{
			name:     "default max conns per host",
			got:      cfg.transport.maxConnsPerHost,
			expected: DefaultMaxConnsPerHost,
		},
		{
			name:     "default idle conn timeout",
			got:      cfg.transport.idleConnTimeout,
			expected: DefaultIdleConnTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// ============================================================================
// Combined Options Tests
// ============================================================================

func TestMultipleOptions_AppliedCorrectly(t *testing.T) {
	timeout := 5 * time.Second
	dialTimeout := 3 * time.Second
	maxIdleConns := 50
	maxConnsPerHost := 25

	cfg := newConfig(
		WithTimeout(timeout),
		WithDialTimeout(dialTimeout),
		WithMaxIdleConns(maxIdleConns),
		WithMaxConnsPerHost(maxConnsPerHost),
	)

	if cfg.timeout != timeout {
		t.Errorf("timeout = %v, want %v", cfg.timeout, timeout)
	}
	if cfg.dialTimeout != dialTimeout {
		t.Errorf("dialTimeout = %v, want %v", cfg.dialTimeout, dialTimeout)
	}
	if cfg.transport.maxIdleConns != maxIdleConns {
		t.Errorf("maxIdleConns = %v, want %v", cfg.transport.maxIdleConns, maxIdleConns)
	}
	if cfg.transport.maxConnsPerHost != maxConnsPerHost {
		t.Errorf("maxConnsPerHost = %v, want %v", cfg.transport.maxConnsPerHost, maxConnsPerHost)
	}
}

func TestBuildTransport_ReturnsConfiguredTransport(t *testing.T) {
	dialTimeout := 5 * time.Second
	tlsTimeout := 8 * time.Second
	responseTimeout := 12 * time.Second
	maxIdleConns := 75
	maxIdleConnsPerHost := 15
	maxConnsPerHost := 50
	idleConnTimeout := 60 * time.Second

	cfg := newConfig(
		WithDialTimeout(dialTimeout),
		WithTLSHandshakeTimeout(tlsTimeout),
		WithResponseHeaderTimeout(responseTimeout),
		WithMaxIdleConns(maxIdleConns),
		WithMaxIdleConnsPerHost(maxIdleConnsPerHost),
		WithMaxConnsPerHost(maxConnsPerHost),
		WithIdleConnTimeout(idleConnTimeout),
	)

	transport := cfg.buildTransport()

	if transport.TLSHandshakeTimeout != tlsTimeout {
		t.Errorf("TLSHandshakeTimeout = %v, want %v", transport.TLSHandshakeTimeout, tlsTimeout)
	}
	if transport.ResponseHeaderTimeout != responseTimeout {
		t.Errorf("ResponseHeaderTimeout = %v, want %v", transport.ResponseHeaderTimeout, responseTimeout)
	}
	if transport.MaxIdleConns != maxIdleConns {
		t.Errorf("MaxIdleConns = %v, want %v", transport.MaxIdleConns, maxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != maxIdleConnsPerHost {
		t.Errorf("MaxIdleConnsPerHost = %v, want %v", transport.MaxIdleConnsPerHost, maxIdleConnsPerHost)
	}
	if transport.MaxConnsPerHost != maxConnsPerHost {
		t.Errorf("MaxConnsPerHost = %v, want %v", transport.MaxConnsPerHost, maxConnsPerHost)
	}
	if transport.IdleConnTimeout != idleConnTimeout {
		t.Errorf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, idleConnTimeout)
	}
	if !transport.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 should be true")
	}
}

// ============================================================================
// TLS Options Tests
// ============================================================================

func TestTLS_DefaultValues(t *testing.T) {
	cfg := newConfig()

	if cfg.tls.enabled {
		t.Error("TLS should be disabled by default")
	}
	if cfg.tls.insecureSkipVerify {
		t.Error("insecureSkipVerify should be false by default")
	}
	if cfg.tls.minVersion != tls.VersionTLS12 {
		t.Errorf("minVersion = %v, want %v (TLS 1.2)", cfg.tls.minVersion, tls.VersionTLS12)
	}
	if cfg.tls.maxVersion != 0 {
		t.Errorf("maxVersion = %v, want 0 (no limit)", cfg.tls.maxVersion)
	}
}

func TestWithTLSConfig_SetsCustomConfig(t *testing.T) {
	customConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
	}

	cfg := newConfig(WithTLSConfig(customConfig))

	if !cfg.tls.enabled {
		t.Error("TLS should be enabled when custom config is set")
	}
	if cfg.tls.customConfig != customConfig {
		t.Error("customConfig should be set to provided config")
	}
}

func TestWithInsecureSkipVerify_EnablesOption(t *testing.T) {
	cfg := newConfig(WithInsecureSkipVerify())

	if !cfg.tls.enabled {
		t.Error("TLS should be enabled when InsecureSkipVerify is set")
	}
	if !cfg.tls.insecureSkipVerify {
		t.Error("insecureSkipVerify should be true")
	}
}

func TestWithTLSMinVersion_SetsValue(t *testing.T) {
	tests := []struct {
		name    string
		version uint16
	}{
		{"TLS 1.0", tls.VersionTLS10},
		{"TLS 1.1", tls.VersionTLS11},
		{"TLS 1.2", tls.VersionTLS12},
		{"TLS 1.3", tls.VersionTLS13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig(WithTLSMinVersion(tt.version))

			if !cfg.tls.enabled {
				t.Error("TLS should be enabled when MinVersion is set")
			}
			if cfg.tls.minVersion != tt.version {
				t.Errorf("minVersion = %v, want %v", cfg.tls.minVersion, tt.version)
			}
		})
	}
}

func TestWithTLSMaxVersion_SetsValue(t *testing.T) {
	tests := []struct {
		name    string
		version uint16
	}{
		{"TLS 1.2", tls.VersionTLS12},
		{"TLS 1.3", tls.VersionTLS13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newConfig(WithTLSMaxVersion(tt.version))

			if !cfg.tls.enabled {
				t.Error("TLS should be enabled when MaxVersion is set")
			}
			if cfg.tls.maxVersion != tt.version {
				t.Errorf("maxVersion = %v, want %v", cfg.tls.maxVersion, tt.version)
			}
		})
	}
}

func TestWithRootCAFromPEM_AddsCA(t *testing.T) {
	// Valid CA certificate PEM (self-signed for testing)
	caPEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHBfpegPjMCMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBnRl
c3RjYTAeFw0yMzAxMDEwMDAwMDBaFw0zMzAxMDEwMDAwMDBaMBExDzANBgNVBAMM
BnRlc3RjYTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC7o96WsME5mq+5tJMHBjJL
oLvXwFiCgReJ0X5kXJA8uJCJoLvXS8AEvPNeRALd5TxDwCCX6F4NNANBgSmKdN7b
AgMBAAGjUzBRMB0GA1UdDgQWBBQJtOxEaJU3+rzJFcChPv5Yn8tj0DAfBgNVHSME
GDAWgBQJtOxEaJU3+rzJFcChPv5Yn8tj0DAPBgNVHRMBAf8EBTADAQH/MA0GCSqG
SIb3DQEBCwUAA0EAqVrPvOLOb2qPdOLQ8GOQB8gF3rrP8FqP5dYf0LvfK1qLwryO
9t1TbJI1xIeS3GHPFT0FrDl5fX5hPOUAWnZhEA==
-----END CERTIFICATE-----`)

	cfg := newConfig(WithRootCAFromPEM(caPEM))

	if !cfg.tls.enabled {
		t.Error("TLS should be enabled when RootCA is set")
	}
	if cfg.tls.rootCAs == nil {
		t.Error("rootCAs should not be nil")
	}
}

func TestWithRootCA_NonExistentFile(t *testing.T) {
	cfg := newConfig(WithRootCA("/nonexistent/path/ca.pem"))

	// Should still enable TLS even if file doesn't exist
	if !cfg.tls.enabled {
		t.Error("TLS should be enabled even when file doesn't exist")
	}
	// rootCAs should be nil since file doesn't exist
	if cfg.tls.rootCAs != nil {
		t.Error("rootCAs should be nil when file doesn't exist")
	}
}

func TestWithRootCA_ValidFile(t *testing.T) {
	// Create a temporary CA file
	caPEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHBfpegPjMCMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBnRl
c3RjYTAeFw0yMzAxMDEwMDAwMDBaFw0zMzAxMDEwMDAwMDBaMBExDzANBgNVBAMM
BnRlc3RjYTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC7o96WsME5mq+5tJMHBjJL
oLvXwFiCgReJ0X5kXJA8uJCJoLvXS8AEvPNeRALd5TxDwCCX6F4NNANBgSmKdN7b
AgMBAAGjUzBRMB0GA1UdDgQWBBQJtOxEaJU3+rzJFcChPv5Yn8tj0DAfBgNVHSME
GDAWgBQJtOxEaJU3+rzJFcChPv5Yn8tj0DAPBgNVHRMBAf8EBTADAQH/MA0GCSqG
SIb3DQEBCwUAA0EAqVrPvOLOb2qPdOLQ8GOQB8gF3rrP8FqP5dYf0LvfK1qLwryO
9t1TbJI1xIeS3GHPFT0FrDl5fX5hPOUAWnZhEA==
-----END CERTIFICATE-----`)

	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	if err := os.WriteFile(caFile, caPEM, 0644); err != nil {
		t.Fatalf("Failed to write temp CA file: %v", err)
	}

	cfg := newConfig(WithRootCA(caFile))

	if !cfg.tls.enabled {
		t.Error("TLS should be enabled when RootCA file is valid")
	}
	if cfg.tls.rootCAs == nil {
		t.Error("rootCAs should not be nil when file is valid")
	}
}

func TestWithClientCertificateFromPEM_AddsCertificate(t *testing.T) {
	// Generate a valid self-signed certificate for testing
	certPEM, keyPEM := generateSelfSignedCertForTest(t)

	cfg := newConfig(WithClientCertificateFromPEM(certPEM, keyPEM))

	if !cfg.tls.enabled {
		t.Error("TLS should be enabled when client certificate is set")
	}
	if len(cfg.tls.certificates) != 1 {
		t.Errorf("certificates length = %d, want 1", len(cfg.tls.certificates))
	}
}

func TestWithClientCertificateFromPEM_InvalidCert(t *testing.T) {
	invalidCertPEM := []byte("invalid certificate")
	invalidKeyPEM := []byte("invalid key")

	cfg := newConfig(WithClientCertificateFromPEM(invalidCertPEM, invalidKeyPEM))

	// Should still enable TLS even with invalid cert
	if !cfg.tls.enabled {
		t.Error("TLS should be enabled even with invalid certificate")
	}
	// But certificates should be empty
	if len(cfg.tls.certificates) != 0 {
		t.Errorf("certificates length = %d, want 0", len(cfg.tls.certificates))
	}
}

func TestWithClientCertificate_NonExistentFiles(t *testing.T) {
	cfg := newConfig(WithClientCertificate("/nonexistent/cert.pem", "/nonexistent/key.pem"))

	// Should still enable TLS even if files don't exist
	if !cfg.tls.enabled {
		t.Error("TLS should be enabled even when files don't exist")
	}
	// But certificates should be empty
	if len(cfg.tls.certificates) != 0 {
		t.Errorf("certificates length = %d, want 0", len(cfg.tls.certificates))
	}
}

func TestWithClientCertificate_ValidFiles(t *testing.T) {
	// Generate a valid self-signed certificate for testing
	certPEM, keyPEM := generateSelfSignedCertForTest(t)

	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "client.pem")
	keyFile := filepath.Join(tmpDir, "client-key.pem")

	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		t.Fatalf("Failed to write temp cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		t.Fatalf("Failed to write temp key file: %v", err)
	}

	cfg := newConfig(WithClientCertificate(certFile, keyFile))

	if !cfg.tls.enabled {
		t.Error("TLS should be enabled when certificate files are valid")
	}
	if len(cfg.tls.certificates) != 1 {
		t.Errorf("certificates length = %d, want 1", len(cfg.tls.certificates))
	}
}

func TestBuildTLSConfig_WithCustomConfig(t *testing.T) {
	customConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	cfg := newConfig(WithTLSConfig(customConfig))
	tlsConfig := cfg.buildTLSConfig()

	if tlsConfig != customConfig {
		t.Error("buildTLSConfig should return the custom config when set")
	}
}

func TestBuildTLSConfig_WithOptions(t *testing.T) {
	cfg := newConfig(
		WithInsecureSkipVerify(),
		WithTLSMinVersion(tls.VersionTLS13),
		WithTLSMaxVersion(tls.VersionTLS13),
	)

	tlsConfig := cfg.buildTLSConfig()

	if !tlsConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
	if tlsConfig.MinVersion != tls.VersionTLS13 {
		t.Errorf("MinVersion = %v, want %v", tlsConfig.MinVersion, tls.VersionTLS13)
	}
	if tlsConfig.MaxVersion != tls.VersionTLS13 {
		t.Errorf("MaxVersion = %v, want %v", tlsConfig.MaxVersion, tls.VersionTLS13)
	}
}

func TestBuildTransport_WithTLSEnabled(t *testing.T) {
	cfg := newConfig(WithInsecureSkipVerify())
	transport := cfg.buildTransport()

	if transport.TLSClientConfig == nil {
		t.Error("TLSClientConfig should not be nil when TLS is enabled")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true in transport TLS config")
	}
}

func TestBuildTransport_WithoutTLSEnabled(t *testing.T) {
	cfg := newConfig()
	transport := cfg.buildTransport()

	if transport.TLSClientConfig != nil {
		t.Error("TLSClientConfig should be nil when TLS is not enabled")
	}
}

func TestMultipleTLSOptions_AppliedCorrectly(t *testing.T) {
	caPEM := []byte(`-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHBfpegPjMCMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBnRl
c3RjYTAeFw0yMzAxMDEwMDAwMDBaFw0zMzAxMDEwMDAwMDBaMBExDzANBgNVBAMM
BnRlc3RjYTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC7o96WsME5mq+5tJMHBjJL
oLvXwFiCgReJ0X5kXJA8uJCJoLvXS8AEvPNeRALd5TxDwCCX6F4NNANBgSmKdN7b
AgMBAAGjUzBRMB0GA1UdDgQWBBQJtOxEaJU3+rzJFcChPv5Yn8tj0DAfBgNVHSME
GDAWgBQJtOxEaJU3+rzJFcChPv5Yn8tj0DAPBgNVHRMBAf8EBTADAQH/MA0GCSqG
SIb3DQEBCwUAA0EAqVrPvOLOb2qPdOLQ8GOQB8gF3rrP8FqP5dYf0LvfK1qLwryO
9t1TbJI1xIeS3GHPFT0FrDl5fX5hPOUAWnZhEA==
-----END CERTIFICATE-----`)

	cfg := newConfig(
		WithTLSMinVersion(tls.VersionTLS12),
		WithTLSMaxVersion(tls.VersionTLS13),
		WithRootCAFromPEM(caPEM),
	)

	if !cfg.tls.enabled {
		t.Error("TLS should be enabled")
	}
	if cfg.tls.minVersion != tls.VersionTLS12 {
		t.Errorf("minVersion = %v, want %v", cfg.tls.minVersion, tls.VersionTLS12)
	}
	if cfg.tls.maxVersion != tls.VersionTLS13 {
		t.Errorf("maxVersion = %v, want %v", cfg.tls.maxVersion, tls.VersionTLS13)
	}
	if cfg.tls.rootCAs == nil {
		t.Error("rootCAs should not be nil")
	}
}

