package clienthttp

import (
	"bytes"
	"clienthttp/pkg/adapter"
	"clienthttp/pkg/structs"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	return &ClientHttp{
		httpClient:  &http.Client{},
		baseUrl:     baseUrl,
		config:      *cfg,
		auditory:    auditory,
		correlation: correlation,
	}, nil
}

func (c ClientHttp) DoRequest(ctx context.Context, method string, endpoint string, queryParams map[string]string, body []byte, options ...structs.NewRequestModifier) (*structs.Response, error) {
	var (
		auditInput  *structs.Request
		auditOutput *structs.Response
	)

	endpoint = fmtEndpoint(endpoint)
	urlReq := fmt.Sprintf("%s/%s", c.baseUrl, endpoint)

	httpReq, err := http.NewRequest(method, urlReq, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	c.setHeaders(ctx, httpReq)
	c.setQueryParams(nil, httpReq, queryParams)
	c.setCookies(nil, httpReq)

	auditInput = structs.NewRequest(urlReq, method, httpReq.Header, httpReq.URL.String(), httpReq.Cookies(), body)

	for i := range options {
		options[i](httpReq)
	}

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	response, err := c.toResponse(ctx, httpRes)
	if err != nil {
		return nil, err
	}

	auditOutput = response
	c.auditory.Save(ctx, auditInput, auditOutput)
	return response, nil
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

func (c ClientHttp) setHeaders(ctx context.Context, req *http.Request) {
	req.Header.Set("Content-Type", c.config.contentType)
	if c.config.authCallback != nil {
		c.config.authCallback(req)
	}

	if c.config.enabledCorrelationID {
		req.Header.Set("correlation_id", c.correlation(ctx))
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
