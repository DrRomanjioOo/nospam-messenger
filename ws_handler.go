package handler

import (
	"context"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
)

type AuthAPI interface {
	Register(ctx context.Context, login, password string) (domain.User, string, error)
	Login(ctx context.Context, login, password string) (domain.User, string, error)
	ResolveSession(ctx context.Context, sessionID string) (repository.SessionData, error)
}

type MessageAPI interface {
	List(ctx context.Context, beforeID int64, limit int) ([]domain.Message, error)
	Send(ctx context.Context, userID int64, login, content string) (domain.Message, error)
	Get(ctx context.Context, id int64) (domain.Message, error)
	DeleteByUser(ctx context.Context, messageID, userID int64) (domain.Message, error)
}
