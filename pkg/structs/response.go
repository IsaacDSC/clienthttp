package structs

import "net/http"

type NewRequestModifier func(r *http.Request)
type Response struct {
	StatusCode int
	Body       []byte
}

func (r Response) IsStatusSuccessfully() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
