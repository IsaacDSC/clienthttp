// Package main implements a simple HTTPS server with mTLS support
// for demonstrating the clienthttp TLS features.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Command-line flags
	port := flag.Int("port", 8443, "Port to listen on")
	certsDir := flag.String("certs", "../certs", "Directory containing certificates")
	requireClientCert := flag.Bool("mtls", false, "Require client certificate (mTLS)")
	flag.Parse()

	// Certificate paths
	serverCert := filepath.Join(*certsDir, "server.pem")
	serverKey := filepath.Join(*certsDir, "server-key.pem")
	caCert := filepath.Join(*certsDir, "ca.pem")

	// Load server certificate
	cert, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	// Load CA certificate for client verification (mTLS)
	caCertPEM, err := os.ReadFile(caCert)
	if err != nil {
		log.Fatalf("Failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		log.Fatal("Failed to parse CA certificate")
	}

	// Configure TLS
	clientAuth := tls.NoClientCert
	if *requireClientCert {
		clientAuth = tls.RequireAndVerifyClientCert
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   clientAuth,
		MinVersion:   tls.VersionTLS12,
	}

	// HTTP handlers
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		clientInfo := "none"
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			clientInfo = r.TLS.PeerCertificates[0].Subject.CommonName
		}

		fmt.Fprintf(w, `{"status":"ok","message":"Hello from TLS server!","client":"%s"}`, clientInfo)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy"}`)
	})

	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Echo request information
		fmt.Fprintf(w, `{"method":"%s","path":"%s","host":"%s"}`,
			r.Method, r.URL.Path, r.Host)
	})

	// Create server
	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	// Start server
	mode := "TLS"
	if *requireClientCert {
		mode = "mTLS"
	}

	log.Printf("Starting HTTPS server (%s mode) on https://localhost%s", mode, addr)
	log.Printf("Certificates directory: %s", *certsDir)

	if *requireClientCert {
		log.Println("Client certificate required for connections")
	}

	if err := server.ListenAndServeTLS(serverCert, serverKey); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

