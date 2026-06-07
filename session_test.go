package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/middleware"
)

func TestCORS_Preflight(t *testing.T) {
	allowed := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	h := middleware.CORS(allowed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:3000" {
		t.Fatalf("headers=%v", rr.Header())
	}
}

func TestCORS_PassThrough(t *testing.T) {
	called := false
	h := middleware.CORS([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if !called {
		t.Fatal("handler not called")
	}
}
