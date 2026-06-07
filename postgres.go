package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/middleware"
)

func TestLogging_PassThrough(t *testing.T) {
	called := false
	h := middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/x", nil))
	if !called || rr.Code != http.StatusCreated {
		t.Fatalf("called=%v status=%d", called, rr.Code)
	}
}
