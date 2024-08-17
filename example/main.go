package main

import (
	"clienthttp/pkg/clienthttp"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

func GetCorrelation(ctx context.Context) string {
	return uuid.New().String()
}

func main() {
	ctx := context.Background()
	baseURL := "http://localhost:8081"

	auditoryService := NewAuditoryService()
	client, err := clienthttp.NewClientHttp(baseURL, auditoryService, GetCorrelation)
	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(map[string]any{"username": "isaacdsc"})
	if err != nil {
		panic(err)
	}

	res, err := client.DoRequest(ctx, http.MethodPost, "/user", nil, b)
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println(string(res.Body))
	fmt.Println()
}
