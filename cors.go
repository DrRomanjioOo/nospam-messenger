package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/handler"
)

type healthMock struct {
	pgErr error
	rErr  error
}

func (h healthMock) PingPostgres(r *http.Request) error { return h.pgErr }
func (h healthMock) PingRedis(r *http.Request) error    { return h.rErr }

func TestHealthHandler_OK(t *testing.T) {
	h := handler.NewHealthHandler(healthMock{})
	rr := httptest.NewRecorder()
	h.Serve(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestHealthHandler_Degraded(t *testing.T) {
	h := handler.NewHealthHandler(healthMock{pgErr: errors.New("pg down")})
	rr := httptest.NewRecorder()
	h.Serve(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
