package middleware_test

import (
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/middleware"
)

func TestParseAllowedOrigins(t *testing.T) {
	got := middleware.ParseAllowedOrigins("http://localhost:3000, http://127.0.0.1:3000")
	if len(got) != 2 || got[1] != "http://127.0.0.1:3000" {
		t.Fatalf("got=%v", got)
	}
}

func TestIsAllowedOrigin(t *testing.T) {
	allowed := []string{"http://localhost:3000"}
	if !middleware.IsAllowedOrigin("http://localhost:3000", allowed) {
		t.Fatal("expected localhost to be allowed")
	}
	if !middleware.IsAllowedOrigin("http://127.0.0.1:5173", allowed) {
		t.Fatal("expected loopback origin to be allowed")
	}
	if middleware.IsAllowedOrigin("http://evil.test", allowed) {
		t.Fatal("expected evil origin to be rejected")
	}
	if !middleware.IsAllowedOrigin("", allowed) {
		t.Fatal("expected empty origin to be allowed")
	}
}
