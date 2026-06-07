package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/handler"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/middleware"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/service"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/testutil"
)

func TestAuthHandler_Register(t *testing.T) {
	h := handler.NewAuthHandler(&testutil.MockAuthAPI{
		RegisterFn: func(ctx context.Context, login, password string) (domain.User, string, error) {
			return domain.User{ID: 1, Login: login}, "sid", nil
		},
	}, "session_id", false, 0)

	body := bytes.NewBufferString(`{"login":"alice1","password":"password1"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	c := rr.Result().Cookies()
	if len(c) == 0 || c[0].Name != "session_id" {
		t.Fatalf("cookies=%v", c)
	}
}

func TestAuthHandler_Register_LoginTaken(t *testing.T) {
	h := handler.NewAuthHandler(&testutil.MockAuthAPI{
		RegisterFn: func(ctx context.Context, login, password string) (domain.User, string, error) {
			return domain.User{}, "", service.ErrLoginTaken
		},
	}, "session_id", false, 0)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"login":"alice1","password":"password1"}`))
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAuthHandler_Login(t *testing.T) {
	h := handler.NewAuthHandler(&testutil.MockAuthAPI{
		LoginFn: func(ctx context.Context, login, password string) (domain.User, string, error) {
			return domain.User{ID: 2, Login: login}, "sid2", nil
		},
	}, "session_id", false, 0)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"login":"alice1","password":"password1"}`))
	rr := httptest.NewRecorder()
	h.Login(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAuthHandler_Register_BadJSON(t *testing.T) {
	h := handler.NewAuthHandler(&testutil.MockAuthAPI{}, "session_id", false, 0)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{`))
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAuthHandler_Me(t *testing.T) {
	h := handler.NewAuthHandler(&testutil.MockAuthAPI{}, "session_id", false, 0)
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(middleware.WithSessionForTest(req.Context(), repository.SessionData{
		UserID: 7,
		Login:  "alice1",
	}))
	rr := httptest.NewRecorder()
	h.Me(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAuthHandler_mapAuthError_ViaRegister(t *testing.T) {
	h := handler.NewAuthHandler(&testutil.MockAuthAPI{
		RegisterFn: func(ctx context.Context, login, password string) (domain.User, string, error) {
			return domain.User{}, "", errors.New("boom")
		},
	}, "session_id", false, 0)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"login":"alice1","password":"password1"}`))
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rr.Code)
	}
}

