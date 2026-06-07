package middleware

import (
	"context"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
)

type sessionKey struct{}
type sessionIDKey struct{}

// WithSession attaches session data and the resolved session id to the context.
func WithSession(ctx context.Context, data repository.SessionData, sessionID string) context.Context {
	ctx = context.WithValue(ctx, sessionKey{}, data)
	if sessionID != "" {
		ctx = context.WithValue(ctx, sessionIDKey{}, sessionID)
	}
	return ctx
}

// SessionIDFromContext returns the validated session id for the current request.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionIDKey{}).(string)
	return v, ok
}

// WithSessionForTest attaches session data in tests.
func WithSessionForTest(ctx context.Context, data repository.SessionData) context.Context {
	return context.WithValue(ctx, sessionKey{}, data)
}
