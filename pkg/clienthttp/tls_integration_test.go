package clienthttp

import (
	"clienthttp/pkg/structs"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ============================================================================
// TLS Integration Tests
// ============================================================================

// testCA holds CA certificate and key for testing
type testCA struct {
	cert    *x509.Certificate
	key     *rsa.PrivateKey
	certPEM []byte
	keyPEM  []byte
}

// testCert holds a certificate and key pair signed by a CA
type testCert struct {
	cert    *x509.Certificate
	key     *rsa.PrivateKey
	certPEM []byte
	keyPEM  []byte
}

// generateTestCA creates a self-signed CA certificate for testing
func generateTestCA() (*testCA, error) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
			CommonName:   "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)})

	return &testCA{
		cert:    caCert,
		key:     caKey,
		certPEM: caCertPEM,
		keyPEM:  caKeyPEM,
	}, nil
}

// generateServerCert creates a server certificate signed by the CA
func generateServerCert(ca *testCA, hosts ...string) (*testCert, error) {
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server key: %w", err)
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Test Server"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			serverTemplate.IPAddresses = append(serverTemplate.IPAddresses, ip)
		} else {
			serverTemplate.DNSNames = append(serverTemplate.DNSNames, h)
		}
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, ca.cert, &serverKey.PublicKey, ca.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	serverCert, err := x509.ParseCertificate(serverCertDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server certificate: %w", err)
	}

	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})

	return &testCert{
		cert:    serverCert,
		key:     serverKey,
		certPEM: serverCertPEM,
		keyPEM:  serverKeyPEM,
	}, nil
}

// generateClientCert creates a client certificate signed by the CA for mTLS
func generateClientCert(ca *testCA) (*testCert, error) {
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate client key: %w", err)
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization: []string{"Test Client"},
			CommonName:   "test-client",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, ca.cert, &clientKey.PublicKey, ca.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create client certificate: %w", err)
	}

	clientCert, err := x509.ParseCertificate(clientCertDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client certificate: %w", err)
	}

	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})
	clientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})

	return &testCert{
		cert:    clientCert,
		key:     clientKey,
		certPEM: clientCertPEM,
		keyPEM:  clientKeyPEM,
	}, nil
}

// createTLSServer creates an HTTPS test server with the given certificates
func createTLSServer(serverCert *testCert, ca *testCA, requireClientCert bool) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	server := httptest.NewUnstartedServer(handler)

	cert, err := tls.X509KeyPair(serverCert.certPEM, serverCert.keyPEM)
	if err != nil {
		panic(fmt.Sprintf("failed to load server certificate: %v", err))
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(ca.cert)

	clientAuth := tls.NoClientCert
	if requireClientCert {
		clientAuth = tls.RequireAndVerifyClientCert
	}

	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   clientAuth,
		MinVersion:   tls.VersionTLS12,
	}

	server.StartTLS()
	return server
}

// TestTLSIntegration_WithRootCA tests connection to HTTPS server with custom CA
func TestTLSIntegration_WithRootCA(t *testing.T) {
	// Generate CA and server certificate
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	// Start TLS server
	server := createTLSServer(serverCert, ca, false)
	defer server.Close()

	// Create client with custom CA
	client, err := NewClientHttp(server.URL, nil, nil,
		WithRootCAFromPEM(ca.certPEM),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request
	resp, err := client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestTLSIntegration_WithRootCAFile tests connection using CA from file
func TestTLSIntegration_WithRootCAFile(t *testing.T) {
	// Generate CA and server certificate
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	// Write CA to temp file
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	if err := os.WriteFile(caFile, ca.certPEM, 0644); err != nil {
		t.Fatalf("Failed to write CA file: %v", err)
	}

	// Start TLS server
	server := createTLSServer(serverCert, ca, false)
	defer server.Close()

	// Create client with CA file
	client, err := NewClientHttp(server.URL, nil, nil,
		WithRootCA(caFile),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request
	resp, err := client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestTLSIntegration_WithInsecureSkipVerify tests connection with certificate verification disabled
func TestTLSIntegration_WithInsecureSkipVerify(t *testing.T) {
	// Generate CA and server certificate
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	// Start TLS server
	server := createTLSServer(serverCert, ca, false)
	defer server.Close()

	// Create client with InsecureSkipVerify (without providing CA)
	client, err := NewClientHttp(server.URL, nil, nil,
		WithInsecureSkipVerify(),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request - should succeed despite unknown CA
	resp, err := client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestTLSIntegration_FailsWithUntrustedCA tests that connection fails without proper CA
func TestTLSIntegration_FailsWithUntrustedCA(t *testing.T) {
	// Generate CA and server certificate
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	// Start TLS server
	server := createTLSServer(serverCert, ca, false)
	defer server.Close()

	// Generate a different CA (untrusted)
	untrustedCA, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate untrusted CA: %v", err)
	}

	// Create client with wrong CA
	client, err := NewClientHttp(server.URL, nil, nil,
		WithRootCAFromPEM(untrustedCA.certPEM),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request - should fail due to untrusted certificate
	_, err = client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err == nil {
		t.Error("Expected error due to untrusted CA, but request succeeded")
	}
}

// TestTLSIntegration_mTLS tests mutual TLS authentication
func TestTLSIntegration_mTLS(t *testing.T) {
	// Generate CA, server cert, and client cert
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	clientCert, err := generateClientCert(ca)
	if err != nil {
		t.Fatalf("Failed to generate client certificate: %v", err)
	}

	// Start TLS server requiring client certificate
	server := createTLSServer(serverCert, ca, true)
	defer server.Close()

	// Create client with CA and client certificate
	client, err := NewClientHttp(server.URL, nil, nil,
		WithRootCAFromPEM(ca.certPEM),
		WithClientCertificateFromPEM(clientCert.certPEM, clientCert.keyPEM),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request - should succeed with client certificate
	resp, err := client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestTLSIntegration_mTLSFromFiles tests mTLS with certificate files
func TestTLSIntegration_mTLSFromFiles(t *testing.T) {
	// Generate CA, server cert, and client cert
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	clientCert, err := generateClientCert(ca)
	if err != nil {
		t.Fatalf("Failed to generate client certificate: %v", err)
	}

	// Write certificates to temp files
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca.pem")
	clientCertFile := filepath.Join(tmpDir, "client.pem")
	clientKeyFile := filepath.Join(tmpDir, "client-key.pem")

	if err := os.WriteFile(caFile, ca.certPEM, 0644); err != nil {
		t.Fatalf("Failed to write CA file: %v", err)
	}
	if err := os.WriteFile(clientCertFile, clientCert.certPEM, 0644); err != nil {
		t.Fatalf("Failed to write client cert file: %v", err)
	}
	if err := os.WriteFile(clientKeyFile, clientCert.keyPEM, 0600); err != nil {
		t.Fatalf("Failed to write client key file: %v", err)
	}

	// Start TLS server requiring client certificate
	server := createTLSServer(serverCert, ca, true)
	defer server.Close()

	// Create client with file-based certificates
	client, err := NewClientHttp(server.URL, nil, nil,
		WithRootCA(caFile),
		WithClientCertificate(clientCertFile, clientKeyFile),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request
	resp, err := client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestTLSIntegration_mTLSFailsWithoutClientCert tests that mTLS fails without client certificate
func TestTLSIntegration_mTLSFailsWithoutClientCert(t *testing.T) {
	// Generate CA and server cert
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	// Start TLS server requiring client certificate
	server := createTLSServer(serverCert, ca, true)
	defer server.Close()

	// Create client without client certificate
	client, err := NewClientHttp(server.URL, nil, nil,
		WithRootCAFromPEM(ca.certPEM),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request - should fail due to missing client certificate
	_, err = client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err == nil {
		t.Error("Expected error due to missing client certificate, but request succeeded")
	}
}

// TestTLSIntegration_TLSVersionControl tests TLS version constraints
func TestTLSIntegration_TLSVersionControl(t *testing.T) {
	// Generate CA and server certificate
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	// Start TLS server with TLS 1.2 minimum
	server := createTLSServer(serverCert, ca, false)
	defer server.Close()

	// Create client with TLS 1.2 minimum
	client, err := NewClientHttp(server.URL, nil, nil,
		WithRootCAFromPEM(ca.certPEM),
		WithTLSMinVersion(tls.VersionTLS12),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request - should succeed with TLS 1.2+
	resp, err := client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestTLSIntegration_WithCustomTLSConfig tests using a fully custom TLS config
func TestTLSIntegration_WithCustomTLSConfig(t *testing.T) {
	// Generate CA and server certificate
	ca, err := generateTestCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCert, err := generateServerCert(ca, "127.0.0.1", "localhost")
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	// Start TLS server
	server := createTLSServer(serverCert, ca, false)
	defer server.Close()

	// Create custom TLS config
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca.certPEM)

	customTLSConfig := &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS12,
	}

	// Create client with custom TLS config
	client, err := NewClientHttp(server.URL, nil, nil,
		WithTLSConfig(customTLSConfig),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make request
	resp, err := client.Get(context.Background(), structs.GetRequest{
		BaseInput: structs.BaseInput{Endpoint: "/test"},
	})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

