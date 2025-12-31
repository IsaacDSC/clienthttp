package clienthttp

import (
	"bytes"
	"clienthttp/pkg/adapter"
	"clienthttp/pkg/structs"
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
	config     config

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
		Timeout:   cfg.timeout,
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
func (c *config) buildTransport() *http.Transport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   c.dialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   c.tlsHandshakeTimeout,
		ResponseHeaderTimeout: c.responseHeaderTimeout,
		MaxIdleConns:          c.transport.maxIdleConns,
		MaxIdleConnsPerHost:   c.transport.maxIdleConnsPerHost,
		MaxConnsPerHost:       c.transport.maxConnsPerHost,
		IdleConnTimeout:       c.transport.idleConnTimeout,
		ForceAttemptHTTP2:     true,
	}

	// Apply TLS configuration if enabled
	if c.tls.enabled {
		transport.TLSClientConfig = c.buildTLSConfig()
	}

	return transport
}

// buildTLSConfig creates a tls.Config based on the configured TLS options.
func (c *config) buildTLSConfig() *tls.Config {
	// If a custom TLS config was provided, use it directly
	if c.tls.customConfig != nil {
		return c.tls.customConfig
	}

	// Build TLS config from individual options
	return &tls.Config{
		InsecureSkipVerify: c.tls.insecureSkipVerify,
		RootCAs:            c.tls.rootCAs,
		Certificates:       c.tls.certificates,
		MinVersion:         c.tls.minVersion,
		MaxVersion:         c.tls.maxVersion,
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
	req.Header.Set("Content-Type", c.config.contentType)
	if c.config.authCallback != nil {
		c.config.authCallback(req)
	}

	if c.config.enabledCorrelationID {
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
	for i := range c.config.cookies {
		req.AddCookie(&c.config.cookies[i])
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
