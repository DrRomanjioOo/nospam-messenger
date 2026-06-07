package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/middleware"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/service"
)

type AuthHandler struct {
	auth       AuthAPI
	cookieName string
	secure     bool
	sessionTTL time.Duration
}

func NewAuthHandler(auth AuthAPI, cookieName string, secure bool, sessionTTL time.Duration) *AuthHandler {
	return &AuthHandler{auth: auth, cookieName: cookieName, secure: secure, sessionTTL: sessionTTL}
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authResponse struct {
	User struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	} `json:"user"`
	SessionID string `json:"session_id"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	u, sid, err := h.auth.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		h.mapAuthError(w, err)
		return
	}
	h.setSessionCookie(w, sid)
	var resp authResponse
	resp.User.ID = u.ID
	resp.User.Login = u.Login
	resp.SessionID = sid
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, login, ok := SessionFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	var resp authResponse
	resp.User.ID = userID
	resp.User.Login = login
	if sid, ok := middleware.SessionIDFromContext(r.Context()); ok {
		resp.SessionID = sid
	} else {
		resp.SessionID = middleware.SessionIDFromRequest(r, h.cookieName)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.secure,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	u, sid, err := h.auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		h.mapAuthError(w, err)
		return
	}
	h.setSessionCookie(w, sid)
	var resp authResponse
	resp.User.ID = u.ID
	resp.User.Login = u.Login
	resp.SessionID = sid
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.secure,
		MaxAge:   int(h.sessionTTL.Seconds()),
	})
}

func (h *AuthHandler) mapAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidLogin),
		errors.Is(err, service.ErrInvalidPassword):
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, service.ErrLoginTaken):
		writeError(w, http.StatusConflict, "login_taken", err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

const sessionCookieName = "session_id"

func SessionFromContext(r *http.Request) (int64, string, bool) {
	s, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		return 0, "", false
	}
	return s.UserID, s.Login, true
}
