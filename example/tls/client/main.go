// Package main demonstrates using clienthttp with TLS/mTLS configuration.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"clienthttp"
)

func main() {
	// Command-line flags
	serverURL := flag.String("url", "https://localhost:8443", "Server URL")
	certsDir := flag.String("certs", "../certs", "Directory containing certificates")
	useMTLS := flag.Bool("mtls", false, "Use mutual TLS (client certificate)")
	insecure := flag.Bool("insecure", false, "Skip certificate verification (insecure)")
	flag.Parse()

	// Certificate paths
	caCert := filepath.Join(*certsDir, "ca.pem")
	clientCert := filepath.Join(*certsDir, "client.pem")
	clientKey := filepath.Join(*certsDir, "client-key.pem")

	// Build client options
	opts := []clienthttp.Option{
		clienthttp.WithTimeout(10 * time.Second),
		clienthttp.WithTLSMinVersion(tls.VersionTLS12),
	}

	if *insecure {
		fmt.Println("WARNING: Using insecure mode (certificate verification disabled)")
		opts = append(opts, clienthttp.WithInsecureSkipVerify())
	} else {
		// Verify CA certificate exists
		if _, err := os.Stat(caCert); os.IsNotExist(err) {
			log.Fatalf("CA certificate not found: %s\nRun './certs/generate.sh' first", caCert)
		}
		opts = append(opts, clienthttp.WithRootCA(caCert))
	}

	if *useMTLS {
		fmt.Println("Using mTLS mode with client certificate")
		// Verify client certificate exists
		if _, err := os.Stat(clientCert); os.IsNotExist(err) {
			log.Fatalf("Client certificate not found: %s\nRun './certs/generate.sh' first", clientCert)
		}
		opts = append(opts, clienthttp.WithClientCertificate(clientCert, clientKey))
	}

	// Create client
	client, err := clienthttp.New(*serverURL, opts...)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Simple GET request
	fmt.Println("\n=== Example 1: GET / ===")
	resp, err := client.Get(ctx, "/")
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Body: %s\n", resp.String())
	}

	// Example 2: Health check
	fmt.Println("\n=== Example 2: GET /health ===")
	resp, err = client.Get(ctx, "/health")
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Body: %s\n", resp.String())
	}

	// Example 3: Echo endpoint with query params
	fmt.Println("\n=== Example 3: GET /echo ===")
	resp, err = client.Get(ctx, "/echo",
		clienthttp.WithQuery("message", "hello"),
		clienthttp.WithHeader("X-Custom-Header", "test-value"),
	)
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		fmt.Printf("Body: %s\n", resp.String())
	}

	// Example 4: Parse JSON from response
	fmt.Println("\n=== Example 4: GET /health with JSON parsing ===")
	resp, err = client.Get(ctx, "/health")
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		var health map[string]interface{}
		if err := resp.JSON(&health); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
		} else {
			fmt.Printf("Health status: %v\n", health)
		}
	}

	fmt.Println("\n=== Done ===")
}
