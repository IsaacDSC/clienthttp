package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/IsaacDSC/clienthttp"
)

func main() {
	// Configurar log
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Demonstrar diferentes estratégias de retry
	fmt.Println("=== ClientHTTP Retry Examples ===")
	fmt.Println()

	// Exemplo 1: Exponential Backoff (recomendado para produção)
	demonstrateExponentialBackoff()

	// Exemplo 2: Constant Backoff
	demonstrateConstantBackoff()

	// Exemplo 3: Per-request retry override
	demonstratePerRequestRetry()

	// Exemplo 4: Custom retry strategy
	demonstrateCustomRetryStrategy()

	fmt.Println()
	fmt.Println("=== All examples completed! ===")
}

func demonstrateExponentialBackoff() {
	fmt.Println("--- Example 1: Exponential Backoff ---")

	// Criar servidor que falha nas primeiras 2 requisições
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		fmt.Printf("  Server received request #%d\n", count)

		if count < 3 {
			fmt.Printf("  -> Returning 503 Service Unavailable\n")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service temporarily unavailable"}`))
			return
		}

		fmt.Printf("  -> Returning 200 OK\n")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success", "attempts": 3}`))
	}))
	defer server.Close()

	// Criar cliente com exponential backoff
	client, err := clienthttp.New(server.URL,
		clienthttp.WithRetryStrategy(&clienthttp.ExponentialBackoff{
			MaxAttempts:  5,                      // até 5 retries
			InitialDelay: 100 * time.Millisecond, // começa com 100ms
			MaxDelay:     2 * time.Second,        // máximo 2s entre retries
			Multiplier:   2.0,                    // dobra o delay a cada retry
			JitterFactor: 0.1,                    // 10% de jitter para evitar thundering herd
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, "/api/data")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	fmt.Printf("  Final response: %s\n", resp.String())
	fmt.Println()
}

func demonstrateConstantBackoff() {
	fmt.Println("--- Example 2: Constant Backoff ---")

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		fmt.Printf("  Server received request #%d\n", count)

		if count < 2 {
			fmt.Printf("  -> Returning 429 Too Many Requests\n")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		fmt.Printf("  -> Returning 200 OK\n")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Usar ConstantBackoff para rate limiting scenarios
	client, err := clienthttp.New(server.URL,
		clienthttp.WithRetryStrategy(clienthttp.NewConstantBackoff(3, 200*time.Millisecond)),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.Get(context.Background(), "/api/resource")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	fmt.Printf("  Final response: %s\n", resp.String())
	fmt.Println()
}

func demonstratePerRequestRetry() {
	fmt.Println("--- Example 3: Per-Request Retry Override ---")

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		endpoint := r.URL.Path
		fmt.Printf("  Server received request #%d to %s\n", count, endpoint)

		if endpoint == "/critical" && count < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"endpoint": endpoint,
			"attempts": count,
		})
	}))
	defer server.Close()

	// Cliente sem retry padrão
	client, err := clienthttp.New(server.URL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Request crítica com retry agressivo
	fmt.Println("  Making critical request with aggressive retry...")
	resp1, err := client.Get(ctx, "/critical",
		clienthttp.WithRequestRetryStrategy(&clienthttp.ExponentialBackoff{
			MaxAttempts:  5,
			InitialDelay: 50 * time.Millisecond,
			MaxDelay:     500 * time.Millisecond,
			Multiplier:   1.5,
			JitterFactor: 0,
		}),
	)
	if err != nil {
		log.Printf("Critical request failed: %v", err)
	} else {
		fmt.Printf("  Critical response: %s\n", resp1.String())
	}

	// Reset counter
	atomic.StoreInt32(&attempts, 0)

	// Request normal sem retry
	fmt.Println("  Making normal request without retry...")
	resp2, err := client.Get(ctx, "/normal", clienthttp.WithNoRetry())
	if err != nil {
		log.Printf("Normal request failed: %v", err)
	} else {
		fmt.Printf("  Normal response: %s\n", resp2.String())
	}

	fmt.Println()
}

func demonstrateCustomRetryStrategy() {
	fmt.Println("--- Example 4: Custom Retry Strategy ---")

	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		fmt.Printf("  Server received request #%d\n", count)

		// Retorna 500 nas primeiras requisições
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": "success with custom retry"}`))
	}))
	defer server.Close()

	// Estratégia customizada que também faz retry em 500
	customStrategy := &clienthttp.ExponentialBackoff{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0,
		RetryableFunc: func(resp *clienthttp.Response, err error) bool {
			// Retry padrão para erros de rede
			if err != nil {
				return true
			}
			// Também retry em 500 Internal Server Error (normalmente não é retryable)
			if resp != nil && resp.StatusCode == 500 {
				fmt.Printf("  -> Custom logic: retrying on 500 error\n")
				return true
			}
			// Retry em status padrão (429, 502, 503, 504)
			if resp != nil {
				switch resp.StatusCode {
				case 429, 502, 503, 504:
					return true
				}
			}
			return false
		},
	}

	client, err := clienthttp.New(server.URL,
		clienthttp.WithRetryStrategy(customStrategy),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.Get(context.Background(), "/api/flaky")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	fmt.Printf("  Final response: %s\n", resp.String())
}
