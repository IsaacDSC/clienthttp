package client

import "strings"

func isValidBaseUrl(baseUrl string) bool {
	return strings.HasPrefix(baseUrl, "https://") || strings.HasPrefix(baseUrl, "http://")
}
