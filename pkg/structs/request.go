package structs

import "net/http"

type Request struct {
	Url     string
	Method  string
	Headers map[string][]string
	Params  string
	Cookies []*http.Cookie
	Body    []byte
}

func NewRequest(url string, method string, headers map[string][]string, params string, cookies []*http.Cookie, body []byte) *Request {
	return &Request{Url: url, Method: method, Headers: headers, Params: params, Cookies: cookies, Body: body}
}
