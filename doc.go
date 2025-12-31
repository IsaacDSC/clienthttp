// Package clienthttp provides an idiomatic HTTP client for Go with
// configurable timeouts, connection pooling, TLS support, and auditing.
//
// # Basic Usage
//
// Create a client and make requests:
//
//	client, err := clienthttp.New("https://api.example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Simple GET request
//	resp, err := client.Get(ctx, "/users/123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(resp.String())
//
// # JSON Helpers
//
// Use the JSON helpers for automatic marshaling/unmarshaling:
//
//	// GET with JSON response
//	var user User
//	err := client.GetJSON(ctx, "/users/123", &user)
//
//	// POST with JSON body and response
//	newUser := User{Name: "John"}
//	var created User
//	err := client.PostJSON(ctx, "/users", newUser, &created)
//
// # Request Options
//
// Add query parameters and headers per-request:
//
//	resp, err := client.Get(ctx, "/users",
//	    clienthttp.WithQuery("page", "1"),
//	    clienthttp.WithQuery("limit", "10"),
//	    clienthttp.WithHeader("X-Custom-Header", "value"),
//	    clienthttp.WithBearerToken("my-token"),
//	)
//
// # Client Configuration
//
// Configure the client with functional options:
//
//	client, err := clienthttp.New("https://api.example.com",
//	    clienthttp.WithTimeout(10*time.Second),
//	    clienthttp.WithMaxIdleConns(50),
//	    clienthttp.WithAuditor(myAuditor),
//	)
//
// # TLS Configuration
//
// Configure TLS for secure connections:
//
//	// With custom CA certificate
//	client, err := clienthttp.New("https://internal.example.com",
//	    clienthttp.WithRootCA("/path/to/ca.pem"),
//	)
//
//	// With client certificate (mTLS)
//	client, err := clienthttp.New("https://secure.example.com",
//	    clienthttp.WithClientCertificate("/path/to/cert.pem", "/path/to/key.pem"),
//	)
//
// # Auditing
//
// Implement the Auditor interface to log all HTTP traffic:
//
//	type MyAuditor struct{}
//
//	func (a *MyAuditor) Log(ctx context.Context, req *clienthttp.AuditRequest, resp *clienthttp.AuditResponse) {
//	    log.Printf("%s %s -> %d", req.Method, req.URL, resp.StatusCode)
//	}
//
//	client, err := clienthttp.New("https://api.example.com",
//	    clienthttp.WithAuditor(&MyAuditor{}),
//	)
//
// # Error Handling
//
// Errors contain detailed information about failures:
//
//	resp, err := client.Get(ctx, "/users/123")
//	if err != nil {
//	    var httpErr *clienthttp.Error
//	    if errors.As(err, &httpErr) {
//	        log.Printf("Request failed: %s %s (status: %d)",
//	            httpErr.Op, httpErr.URL, httpErr.StatusCode)
//	    }
//	}
package clienthttp

