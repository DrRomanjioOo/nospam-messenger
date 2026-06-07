package middleware

import (
	"context"
	"net/http"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/service"
)

func SessionFromContext(ctx context.Context) (repository.SessionData, bool) {
	v, ok := ctx.Value(sessionKey{}).(repository.SessionData)
	return v, ok
}

const sessionHeaderName = "X-Session-ID"

// SessionIDCandidates returns explicit client tokens first (header, query), then cookie.
func SessionIDCandidates(r *http.Request, cookieName string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, 3)
	add := func(sid string) {
		if sid == "" {
			return
		}
		if _, ok := seen[sid]; ok {
			return
		}
		seen[sid] = struct{}{}
		out = append(out, sid)
	}
	add(r.Header.Get(sessionHeaderName))
	add(r.URL.Query().Get("session_id"))
	if cookie, err := r.Cookie(cookieName); err == nil {
		add(cookie.Value)
	}
	return out
}

// SessionIDFromRequest returns the first candidate token without validating it.
func SessionIDFromRequest(r *http.Request, cookieName string) string {
	candidates := SessionIDCandidates(r, cookieName)
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}

type sessionResolver interface {
	ResolveSession(ctx context.Context, sessionID string) (repository.SessionData, error)
}

// AuthenticateSession validates session candidates in priority order.
func AuthenticateSession(ctx context.Context, r *http.Request, cookieName string, auth sessionResolver) (repository.SessionData, string, bool) {
	for _, sid := range SessionIDCandidates(r, cookieName) {
		sess, err := auth.ResolveSession(ctx, sid)
		if err == nil {
			return sess, sid, true
		}
	}
	return repository.SessionData{}, "", false
}

func RequireSession(auth *service.AuthService, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, sid, ok := AuthenticateSession(r.Context(), r, cookieName, auth)
			if !ok {
				writeUnauthorized(w, "authentication required")
				return
			}
			ctx := WithSession(r.Context(), sess, sid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
