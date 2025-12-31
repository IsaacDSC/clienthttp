package main

import (
	"clienthttp"
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
	URL     string      `json:"url"`
	Method  string      `json:"method"`
	Headers http.Header `json:"headers"`
	Body    string      `json:"body"`
}

type ResponseAudit struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

// Log implements clienthttp.Auditor interface
func (as AuditoryService) Log(ctx context.Context, request *clienthttp.AuditRequest, response *clienthttp.AuditResponse) {
	req := RequestAudit{
		URL:     request.URL,
		Method:  request.Method,
		Headers: request.Headers,
		Body:    string(request.Body),
	}

	res := ResponseAudit{
		StatusCode: response.StatusCode,
		Body:       string(response.Body),
	}

	payload, err := json.MarshalIndent(map[string]any{"request": req, "response": res}, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile("auditory_payload.json", payload, 0644)
}
