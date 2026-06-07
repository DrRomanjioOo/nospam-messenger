package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/middleware"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/service"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/testutil"
)

func TestRequireSession_Unauthorized(t *testing.T) {
	svc := service.NewAuthService(nil, &testutil.MockSessionStore{
		GetFn: func(ctx context.Context, id string) (repository.SessionData, error) {
			return repository.SessionData{}, repository.ErrSessionNotFound
		},
	}, 4)
	h := middleware.RequireSession(svc, "session_id")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestRequireSession_OK(t *testing.T) {
	svc := service.NewAuthService(nil, &testutil.MockSessionStore{
		GetFn: func(ctx context.Context, id string) (repository.SessionData, error) {
			return repository.SessionData{UserID: 9, Login: "bob1"}, nil
		},
		TouchFn: func(ctx context.Context, id string) error { return nil },
	}, 4)
	var gotID int64
	h := middleware.RequireSession(svc, "session_id")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, ok := middleware.SessionFromContext(r.Context())
		if !ok {
			t.Fatal("no session")
		}
		gotID = s.UserID
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || gotID != 9 {
		t.Fatalf("status=%d gotID=%d", rr.Code, gotID)
	}
}

func TestRequireSession_Header(t *testing.T) {
	svc := service.NewAuthService(nil, &testutil.MockSessionStore{
		GetFn: func(ctx context.Context, id string) (repository.SessionData, error) {
			if id != "header-sid" {
				t.Fatalf("sid=%q", id)
			}
			return repository.SessionData{UserID: 3, Login: "carol"}, nil
		},
		TouchFn: func(ctx context.Context, id string) error { return nil },
	}, 4)
	var gotLogin string
	h := middleware.RequireSession(svc, "session_id")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, ok := middleware.SessionFromContext(r.Context())
		if !ok {
			t.Fatal("no session")
		}
		gotLogin = s.Login
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Session-ID", "header-sid")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || gotLogin != "carol" {
		t.Fatalf("status=%d login=%q", rr.Code, gotLogin)
	}
}

func TestAuthenticateSession_PrefersValidQueryOverStaleCookie(t *testing.T) {
	svc := service.NewAuthService(nil, &testutil.MockSessionStore{
		GetFn: func(ctx context.Context, id string) (repository.SessionData, error) {
			if id == "good-sid" {
				return repository.SessionData{UserID: 5, Login: "dave1"}, nil
			}
			return repository.SessionData{}, repository.ErrSessionNotFound
		},
		TouchFn: func(ctx context.Context, id string) error { return nil },
	}, 4)
	req := httptest.NewRequest(http.MethodGet, "/ws?session_id=good-sid", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "stale-sid"})
	sess, sid, ok := middleware.AuthenticateSession(req.Context(), req, "session_id", svc)
	if !ok || sid != "good-sid" || sess.UserID != 5 {
		t.Fatalf("ok=%v sid=%q sess=%+v", ok, sid, sess)
	}
}

func TestSessionFromContext_TestHelper(t *testing.T) {
	ctx := middleware.WithSessionForTest(context.Background(), repository.SessionData{UserID: 1})
	s, ok := middleware.SessionFromContext(ctx)
	if !ok || s.UserID != 1 {
		t.Fatalf("s=%+v ok=%v", s, ok)
	}
}
