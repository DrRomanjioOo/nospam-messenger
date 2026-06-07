package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/handler"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/testutil"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/ws"
)

func TestWSHandler_Unauthorized_NoCookie(t *testing.T) {
	h := handler.NewWSHandler(ws.NewHub(), &testutil.MockAuthAPI{}, &testutil.MockMessageAPI{}, []string{"http://localhost:3000"})
	rr := httptest.NewRecorder()
	h.Serve(rr, httptest.NewRequest(http.MethodGet, "/ws", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestWSHandler_Unauthorized_InvalidSession(t *testing.T) {
	h := handler.NewWSHandler(ws.NewHub(), &testutil.MockAuthAPI{
		ResolveSessionFn: func(ctx context.Context, sessionID string) (repository.SessionData, error) {
			return repository.SessionData{}, repository.ErrSessionNotFound
		},
	}, &testutil.MockMessageAPI{}, []string{"http://localhost:3000"})
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "bad"})
	rr := httptest.NewRecorder()
	h.Serve(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}
