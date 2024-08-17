package clienthttp

import "net/http"

type config struct {
	authCallback         func(r *http.Request)
	cookies              []http.Cookie
	contentType          string
	enabledCorrelationID bool
}

type Option func(*config)

func newConfig(opts ...Option) *config {
	c := new(config)
	c.defaults()

	for i := range opts {
		opts[i](c)
	}

	return c
}

const (
	JsonContentType = "application/json"
)

func (c *config) defaults() {
	c.authCallback = nil
	c.contentType = JsonContentType
	c.cookies = make([]http.Cookie, 0)
}
