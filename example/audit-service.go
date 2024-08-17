package main

import (
	"clienthttp/pkg/structs"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type AuditoryService struct{}

func NewAuditoryService() *AuditoryService {
	return &AuditoryService{}
}

type RequestAudit struct {
	Url     string              `json:"url"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
	Params  string              `json:"params"`
	Cookies []*http.Cookie      `json:"cookies"`
	Body    string              `json:"body"`
}

type ResponseAudit struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func (as AuditoryService) Save(ctx context.Context, request *structs.Request, response *structs.Response) {
	req := RequestAudit{
		Url:     request.Url,
		Method:  request.Method,
		Headers: request.Headers,
		Params:  request.Params,
		Cookies: request.Cookies,
		Body:    string(request.Body),
	}

	res := ResponseAudit{
		StatusCode: response.StatusCode,
		Body:       string(response.Body),
	}

	payload, err := json.Marshal(map[string]any{"request": req, "response": res})
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile("auditory_payload.json", payload, 0644)
}
