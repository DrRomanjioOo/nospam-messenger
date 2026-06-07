package service

import (
	"context"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
)

type UserStore interface {
	Create(ctx context.Context, login, passwordHash string) (domain.User, error)
	GetByLogin(ctx context.Context, login string) (domain.User, string, error)
	GetByID(ctx context.Context, id int64) (domain.User, error)
}

type SessionStore interface {
	Create(ctx context.Context, userID int64, login string) (string, error)
	Get(ctx context.Context, id string) (repository.SessionData, error)
	Touch(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

type MessageStore interface {
	Create(ctx context.Context, userID int64, content string) (domain.Message, error)
	GetByID(ctx context.Context, id int64) (domain.Message, error)
	List(ctx context.Context, beforeID int64, limit int) ([]domain.Message, error)
	SoftDeleteByUser(ctx context.Context, id, userID int64) (domain.Message, error)
	SoftDeleteByAI(ctx context.Context, id int64) (domain.Message, error)
	AttachAuthor(ctx context.Context, m domain.Message) (domain.Message, error)
}

type RateLimiter interface {
	AllowMessage(ctx context.Context, userID int64) (bool, error)
}
