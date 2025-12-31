package clienthttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Client is an HTTP client with configurable options.
type Client struct {
	httpClient *http.Client
	baseURL    string
	config     *config
}

// New creates a new HTTP client with the given base URL and options.
func New(baseURL string, opts ...Option) (*Client, error) {
	if !isValidURL(baseURL) {
		return nil, ErrInvalidURL
	}

	baseURL = normalizeURL(baseURL)
	cfg := newConfig(opts...)

	httpClient := &http.Client{
		Timeout:   cfg.timeout,
		Transport: buildTransport(cfg),
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		config:     cfg,
	}, nil
}

// Get performs an HTTP GET request.
func (c *Client) Get(ctx context.Context, endpoint string, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodGet, endpoint, nil, opts...)
}

// Post performs an HTTP POST request.
func (c *Client) Post(ctx context.Context, endpoint string, body []byte, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodPost, endpoint, body, opts...)
}

// Put performs an HTTP PUT request.
func (c *Client) Put(ctx context.Context, endpoint string, body []byte, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodPut, endpoint, body, opts...)
}

// Patch performs an HTTP PATCH request.
func (c *Client) Patch(ctx context.Context, endpoint string, body []byte, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodPatch, endpoint, body, opts...)
}

// Delete performs an HTTP DELETE request.
func (c *Client) Delete(ctx context.Context, endpoint string, opts ...RequestOption) (*Response, error) {
	return c.Do(ctx, http.MethodDelete, endpoint, nil, opts...)
}

// Do performs an HTTP request with the given method.
func (c *Client) Do(ctx context.Context, method, endpoint string, body []byte, opts ...RequestOption) (*Response, error) {
	rc := newRequestConfig(opts...)

	endpoint = normalizeEndpoint(endpoint)
	reqURL := fmt.Sprintf("%s/%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, newError(method, reqURL, 0, nil, err)
	}

	c.applyHeaders(ctx, req, rc.headers)
	c.applyQueryParams(req, rc.queryParams)
	c.applyCookies(req)
	c.applyBasicAuth(req, rc.headers)

	// Audit request
	var auditReq *AuditRequest
	if c.config.auditor != nil {
		auditReq = &AuditRequest{
			URL:     req.URL.String(),
			Method:  method,
			Headers: req.Header.Clone(),
			Body:    body,
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, newError(method, reqURL, 0, nil, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, newError(method, reqURL, 0, nil, err)
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}

	// Audit response
	if c.config.auditor != nil {
		auditResp := &AuditResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			Body:       respBody,
		}
		c.config.auditor.Log(ctx, auditReq, auditResp)
	}

	if !response.OK() {
		return response, newError(method, reqURL, resp.StatusCode, respBody, ErrRequestFailed)
	}

	return response, nil
}

// PostForm performs a POST request with form data.
func (c *Client) PostForm(ctx context.Context, endpoint string, data map[string]string, opts ...RequestOption) (*Response, error) {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}

	endpoint = normalizeEndpoint(endpoint)
	reqURL := fmt.Sprintf("%s/%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, newError(http.MethodPost, reqURL, 0, nil, err)
	}

	rc := newRequestConfig(opts...)
	c.applyHeaders(ctx, req, rc.headers)
	c.applyQueryParams(req, rc.queryParams)
	c.applyCookies(req)

	// Override Content-Type for form data (must be after applyHeaders)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, newError(http.MethodPost, reqURL, 0, nil, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, newError(http.MethodPost, reqURL, 0, nil, err)
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}

	if !response.OK() {
		return response, newError(http.MethodPost, reqURL, resp.StatusCode, respBody, ErrRequestFailed)
	}

	return response, nil
}

// ============================================================================
// Internal helpers
// ============================================================================

func (c *Client) applyHeaders(ctx context.Context, req *http.Request, headers map[string]string) {
	req.Header.Set("Content-Type", c.config.contentType)

	if c.config.authCallback != nil {
		c.config.authCallback(req)
	}

	// Set correlation ID
	if c.config.correlationFn != nil {
		req.Header.Set("X-Correlation-ID", c.config.correlationFn(ctx))
	} else {
		req.Header.Set("X-Correlation-ID", uuid.New().String())
	}

	// Set request ID
	if req.Header.Get("X-Request-ID") == "" {
		req.Header.Set("X-Request-ID", uuid.New().String())
	}

	// Apply custom headers
	for k, v := range headers {
		if !strings.HasPrefix(k, "_") { // Skip internal markers
			req.Header.Set(k, v)
		}
	}
}

func (c *Client) applyQueryParams(req *http.Request, params map[string]string) {
	if len(params) == 0 {
		return
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
}

func (c *Client) applyCookies(req *http.Request) {
	for i := range c.config.cookies {
		req.AddCookie(&c.config.cookies[i])
	}
}

func (c *Client) applyBasicAuth(req *http.Request, headers map[string]string) {
	user, hasUser := headers["_basic_auth_user"]
	pass, hasPass := headers["_basic_auth_pass"]
	if hasUser && hasPass {
		req.SetBasicAuth(user, pass)
	}
}

func buildTransport(cfg *config) *http.Transport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   cfg.dialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   cfg.tlsHandshakeTimeout,
		ResponseHeaderTimeout: cfg.responseHeaderTimeout,
		MaxIdleConns:          cfg.transport.maxIdleConns,
		MaxIdleConnsPerHost:   cfg.transport.maxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.transport.maxConnsPerHost,
		IdleConnTimeout:       cfg.transport.idleConnTimeout,
		ForceAttemptHTTP2:     true,
	}

	if cfg.tls.enabled {
		transport.TLSClientConfig = buildTLSConfig(cfg)
	}

	return transport
}

func buildTLSConfig(cfg *config) *tls.Config {
	if cfg.tls.customConfig != nil {
		return cfg.tls.customConfig
	}

	return &tls.Config{
		InsecureSkipVerify: cfg.tls.insecureSkipVerify,
		RootCAs:            cfg.tls.rootCAs,
		Certificates:       cfg.tls.certificates,
		MinVersion:         cfg.tls.minVersion,
		MaxVersion:         cfg.tls.maxVersion,
	}
}

func isValidURL(u string) bool {
	return strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")
}

func normalizeURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")
	return u
}

func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimPrefix(endpoint, "/")
	endpoint = strings.TrimSuffix(endpoint, "?")
	return endpoint
}
