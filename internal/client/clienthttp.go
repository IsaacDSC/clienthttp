package client

import (
	"bytes"
	"clienthttp/internal/adapter"
	"clienthttp/internal/structs"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type ClientHttp struct {
	httpClient *http.Client
	baseUrl    string
	config     Config

	auditory    adapter.AuditoryAdapter
	correlation adapter.CorrelationIDAdapter
}

func NewClientHttp(baseUrl string, auditory adapter.AuditoryAdapter, correlation adapter.CorrelationIDAdapter, opts ...Option) (*ClientHttp, error) {
	cfg := newConfig(opts...)

	if !isValidBaseUrl(baseUrl) {
		return nil, errors.New("invalid http url")
	}

	baseUrl = fmtBaseUrl(baseUrl)

	httpClient := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: cfg.buildTransport(),
	}

	return &ClientHttp{
		httpClient:  httpClient,
		baseUrl:     baseUrl,
		config:      *cfg,
		auditory:    auditory,
		correlation: correlation,
	}, nil
}

// buildTransport creates an http.Transport with the configured timeout, connection pool, and TLS settings.
func (c *Config) buildTransport() *http.Transport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   c.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   c.TLSHandshakeTimeout,
		ResponseHeaderTimeout: c.ResponseHeaderTimeout,
		MaxIdleConns:          c.Transport.MaxIdleConns,
		MaxIdleConnsPerHost:   c.Transport.MaxIdleConnsPerHost,
		MaxConnsPerHost:       c.Transport.MaxConnsPerHost,
		IdleConnTimeout:       c.Transport.IdleConnTimeout,
		ForceAttemptHTTP2:     true,
	}

	// Apply TLS configuration if enabled
	if c.TLS.Enabled {
		transport.TLSClientConfig = c.buildTLSConfig()
	}

	return transport
}

// buildTLSConfig creates a tls.Config based on the configured TLS options.
func (c *Config) buildTLSConfig() *tls.Config {
	// If a custom TLS config was provided, use it directly
	if c.TLS.CustomConfig != nil {
		return c.TLS.CustomConfig
	}

	// Build TLS config from individual options
	return &tls.Config{
		InsecureSkipVerify: c.TLS.InsecureSkipVerify,
		RootCAs:            c.TLS.RootCAs,
		Certificates:       c.TLS.Certificates,
		MinVersion:         c.TLS.MinVersion,
		MaxVersion:         c.TLS.MaxVersion,
	}
}

func (c ClientHttp) DoRequest(
	ctx context.Context,
	method string,
	endpoint string,
	queryParams map[string]string,
	body []byte,
	headers map[string]string,
	options ...structs.NewRequestModifier,
) (*structs.Response, error) {
	var (
		auditInput  *structs.Request
		auditOutput *structs.Response
	)

	defer func() {
		if c.auditory != nil {
			c.auditory.Save(ctx, auditInput, auditOutput)
		}
	}()

	endpoint = fmtEndpoint(endpoint)
	urlReq := fmt.Sprintf("%s/%s", c.baseUrl, endpoint)

	httpReq, err := http.NewRequest(method, urlReq, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	c.setHeaders(ctx, httpReq, headers)
	c.setQueryParams(ctx, httpReq, queryParams)
	c.setCookies(ctx, httpReq)

	auditInput = structs.NewRequest(urlReq, method, httpReq.Header, httpReq.URL.String(), httpReq.Cookies(), body)

	for i := range options {
		options[i](httpReq)
	}

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	auditOutput, err = c.toResponse(ctx, httpRes)
	if err != nil {
		return nil, err
	}

	return auditOutput, nil
}

func (c ClientHttp) DoFormRequest(ctx context.Context, endpoint string, data map[string]string) (*structs.Response, error) {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}

	urlReq := fmt.Sprintf("%s/%s", c.baseUrl, endpoint)
	httpRes, err := c.httpClient.PostForm(urlReq, form)
	if err != nil {
		return nil, err
	}

	response, err := c.toResponse(ctx, httpRes)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c ClientHttp) setHeaders(ctx context.Context, req *http.Request, headers map[string]string) {
	req.Header.Set("Content-Type", c.config.ContentType)
	if c.config.AuthCallback != nil {
		c.config.AuthCallback(req)
	}

	if c.config.EnabledCorrelationID {
		req.Header.Set("correlation_id", c.correlation(ctx))
	} else {
		req.Header.Set("correlation_id", uuid.New().String())
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if req.Header.Get("request_id") == "" {
		req.Header.Set("request_id", uuid.New().String())
	}

}

func (c ClientHttp) setQueryParams(ctx context.Context, req *http.Request, queryParamsMap map[string]string) {
	queryParams := req.URL.Query()
	for k, v := range queryParamsMap {
		queryParams.Add(k, v)
	}

	req.URL.RawQuery = queryParams.Encode()
}

func (c ClientHttp) setCookies(ctx context.Context, req *http.Request) {
	for i := range c.config.Cookies {
		req.AddCookie(&c.config.Cookies[i])
	}
}

func (c ClientHttp) toResponse(ctx context.Context, httpResponse *http.Response) (res *structs.Response, err error) {
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	res = &structs.Response{
		StatusCode: httpResponse.StatusCode,
		Body:       body,
		Headers:    httpResponse.Header,
	}

	if err := c.verifyResponseError(ctx, res, httpResponse); err != nil {
		return nil, err
	}

	return res, nil
}

func (c ClientHttp) verifyResponseError(ctx context.Context, res *structs.Response, response *http.Response) error {
	if !res.IsStatusSuccessfully() {
		urlReq := response.Request.URL.String()
		return fmt.Errorf("http:request to %s failed with status %d, body %s",
			urlReq, res.StatusCode, string(res.Body),
		)
	}
	return nil
}
