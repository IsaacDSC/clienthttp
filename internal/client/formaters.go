package client

import "strings"

func fmtEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimPrefix(endpoint, "/")
	endpoint = strings.TrimSuffix(endpoint, "?")
	return endpoint
}

func fmtBaseUrl(baseUrl string) string {
	baseUrl = strings.TrimSpace(baseUrl)
	baseUrl = strings.TrimSuffix(baseUrl, "/")
	return baseUrl
}
