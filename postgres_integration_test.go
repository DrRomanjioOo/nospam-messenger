package middleware

import (
	"net/url"
	"strings"
)

// ParseAllowedOrigins splits a comma-separated origin list and trims entries.
func ParseAllowedOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func isLoopbackOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	switch u.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

// IsAllowedOrigin reports whether origin is in the allowed list.
func IsAllowedOrigin(origin string, allowed []string) bool {
	if origin == "" {
		return true
	}
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	return isLoopbackOrigin(origin)
}
