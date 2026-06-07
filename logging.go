package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/handler"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/middleware"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/service"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/spam"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/testutil"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/ws"
)

func authRequest(req *http.Request, userID int64, login string) *http.Request {
	sess := repository.SessionData{UserID: userID, Login: login}
	return req.WithContext(middleware.WithSessionForTest(req.Context(), sess))
}

func TestMessageHandler_List(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		ListFn: func(ctx context.Context, beforeID int64, limit int) ([]domain.Message, error) {
			return []domain.Message{{ID: 1, AuthorLogin: "a", CreatedAt: time.Now()}}, nil
		},
	}, ws.NewHub())
	req := authRequest(httptest.NewRequest(http.MethodGet, "/messages?limit=10", nil), 1, "a")
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d %s", rr.Code, rr.Body.String())
	}
}

func TestMessageHandler_Send(t *testing.T) {
	hub := ws.NewHub()
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		SendFn: func(ctx context.Context, userID int64, login, content string) (domain.Message, error) {
			return domain.Message{ID: 5, UserID: userID, AuthorLogin: login, Content: content, CreatedAt: time.Now()}, nil
		},
	}, hub)
	req := authRequest(httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBufferString(`{"content":"hi"}`)), 1, "alice1")
	rr := httptest.NewRecorder()
	h.Send(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d %s", rr.Code, rr.Body.String())
	}
}

func TestMessageHandler_Send_Unauthorized(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{}, ws.NewHub())
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBufferString(`{"content":"hi"}`))
	rr := httptest.NewRecorder()
	h.Send(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestMessageHandler_Send_SpamRule(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		SendFn: func(ctx context.Context, userID int64, login, content string) (domain.Message, error) {
			return domain.Message{}, spam.ErrSpamRule
		},
	}, ws.NewHub())
	req := authRequest(httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBufferString(`{"content":" "}`)), 1, "a")
	rr := httptest.NewRecorder()
	h.Send(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestMessageHandler_Delete(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		DeleteByUserFn: func(ctx context.Context, messageID, userID int64) (domain.Message, error) {
			return domain.Message{ID: messageID, DeletedByUser: true, AuthorLogin: "a", CreatedAt: time.Now()}, nil
		},
	}, ws.NewHub())
	req := authRequest(httptest.NewRequest(http.MethodDelete, "/messages/3", nil), 1, "a")
	req.SetPathValue("id", "3")
	rr := httptest.NewRecorder()
	h.Delete(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d %s", rr.Code, rr.Body.String())
	}
}

func TestMessageHandler_Delete_NotFound(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		DeleteByUserFn: func(ctx context.Context, messageID, userID int64) (domain.Message, error) {
			return domain.Message{}, repository.ErrMessageNotFound
		},
	}, ws.NewHub())
	req := authRequest(httptest.NewRequest(http.MethodDelete, "/messages/3", nil), 1, "a")
	req.SetPathValue("id", "3")
	rr := httptest.NewRecorder()
	h.Delete(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestMessageHandler_Delete_InvalidID(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{}, ws.NewHub())
	req := authRequest(httptest.NewRequest(http.MethodDelete, "/messages/x", nil), 1, "a")
	req.SetPathValue("id", "x")
	rr := httptest.NewRecorder()
	h.Delete(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestMessageHandler_Send_RateLimited(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		SendFn: func(ctx context.Context, userID int64, login, content string) (domain.Message, error) {
			return domain.Message{}, spam.ErrRateLimited
		},
	}, ws.NewHub())
	req := authRequest(httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBufferString(`{"content":"ok"}`)), 1, "a")
	rr := httptest.NewRecorder()
	h.Send(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestMessageHandler_Send_TooLong(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		SendFn: func(ctx context.Context, userID int64, login, content string) (domain.Message, error) {
			return domain.Message{}, service.ErrMessageTooLong
		},
	}, ws.NewHub())
	req := authRequest(httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBufferString(`{"content":"x"}`)), 1, "a")
	rr := httptest.NewRecorder()
	h.Send(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestMessageHandler_List_Error(t *testing.T) {
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		ListFn: func(ctx context.Context, beforeID int64, limit int) ([]domain.Message, error) {
			return nil, errors.New("db")
		},
	}, ws.NewHub())
	rr := httptest.NewRecorder()
	h.List(rr, httptest.NewRequest(http.MethodGet, "/messages", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestMessageDTO(t *testing.T) {
	// Indirectly covered by List response JSON
	h := handler.NewMessageHandler(&testutil.MockMessageAPI{
		ListFn: func(ctx context.Context, beforeID int64, limit int) ([]domain.Message, error) {
			return []domain.Message{{
				ID: 1, UserID: 2, AuthorLogin: "u", Content: "c",
				CreatedAt: time.Unix(0, 0), DeletedByUser: true,
			}}, nil
		},
	}, ws.NewHub())
	rr := httptest.NewRecorder()
	h.List(rr, httptest.NewRequest(http.MethodGet, "/messages", nil))
	var out struct {
		Messages []map[string]any `json:"messages"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil || len(out.Messages) != 1 {
		t.Fatalf("body=%s err=%v", rr.Body.String(), err)
	}
}
