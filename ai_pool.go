package service

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/spam"
)

const maxMessageLen = 2000

var ErrMessageTooLong = errors.New("message too long")
var ErrForbidden = errors.New("forbidden")

type MessageEnqueuer interface {
	Enqueue(messageID int64, content string)
}

type MessageService struct {
	messages  MessageStore
	rateLimit RateLimiter
	rules     *spam.RuleChecker
	aiPool    MessageEnqueuer
}

func (s *MessageService) SetAIPool(p MessageEnqueuer) {
	s.aiPool = p
}

func NewMessageService(
	messages MessageStore,
	rateLimit RateLimiter,
	rules *spam.RuleChecker,
	aiPool MessageEnqueuer,
) *MessageService {
	return &MessageService{
		messages:  messages,
		rateLimit: rateLimit,
		rules:     rules,
		aiPool:    aiPool,
	}
}

func (s *MessageService) List(ctx context.Context, beforeID int64, limit int) ([]domain.Message, error) {
	return s.messages.List(ctx, beforeID, limit)
}

func (s *MessageService) Send(ctx context.Context, userID int64, login, content string) (domain.Message, error) {
	if utf8.RuneCountInString(content) > maxMessageLen {
		return domain.Message{}, ErrMessageTooLong
	}

	if err := s.rules.Check(content); err != nil {
		return domain.Message{}, err
	}

	ok, err := s.rateLimit.AllowMessage(ctx, userID)
	if err != nil {
		return domain.Message{}, err
	}
	if !ok {
		return domain.Message{}, spam.ErrRateLimited
	}

	m, err := s.messages.Create(ctx, userID, content)
	if err != nil {
		return domain.Message{}, err
	}
	m.AuthorLogin = login

	if s.aiPool != nil {
		s.aiPool.Enqueue(m.ID, content)
	}
	return m, nil
}

func (s *MessageService) DeleteByUser(ctx context.Context, messageID, userID int64) (domain.Message, error) {
	m, err := s.messages.SoftDeleteByUser(ctx, messageID, userID)
	if err != nil {
		return domain.Message{}, err
	}
	m, err = s.messages.AttachAuthor(ctx, m)
	if err != nil {
		return domain.Message{}, fmt.Errorf("attach author: %w", err)
	}
	return m, nil
}

func (s *MessageService) ApplyAIDeletion(ctx context.Context, messageID int64) (domain.Message, error) {
	m, err := s.messages.SoftDeleteByAI(ctx, messageID)
	if err != nil {
		return domain.Message{}, err
	}
	return s.messages.AttachAuthor(ctx, m)
}

func (s *MessageService) Get(ctx context.Context, id int64) (domain.Message, error) {
	return s.messages.GetByID(ctx, id)
}
