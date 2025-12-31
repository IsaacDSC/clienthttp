package clienthttp

import (
	"testing"
	"time"
)

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

