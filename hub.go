package spam

import (
	"fmt"
	"strings"
)

// APIError is a non-2xx response from the AI provider HTTP API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("openrouter status %d", e.StatusCode)
}

// Retryable reports whether the caller should retry the request later.
func (e *APIError) Retryable() bool {
	switch e.StatusCode {
	case 408, 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

// Permanent reports client/config errors that will not succeed on retry.
func (e *APIError) Permanent() bool {
	switch e.StatusCode {
	case 400, 401, 403, 404:
		return true
	default:
		return false
	}
}

// IsValidOpenRouterAPIKey reports whether key looks like an OpenRouter API key.
func IsValidOpenRouterAPIKey(key string) bool {
	k := strings.TrimSpace(key)
	return strings.HasPrefix(k, "sk-or-v1-") && len(k) >= 30
}
