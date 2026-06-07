package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionData struct {
	UserID    int64  `json:"user_id"`
	Login     string `json:"login"`
	CreatedAt int64  `json:"created_at"`
}

type SessionRepository struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewSessionRepository(rdb *redis.Client, ttl time.Duration) *SessionRepository {
	return &SessionRepository{rdb: rdb, ttl: ttl}
}

func (r *SessionRepository) key(id string) string {
	return "session:" + id
}

func (r *SessionRepository) Create(ctx context.Context, userID int64, login string) (string, error) {
	id := uuid.NewString()
	data := SessionData{
		UserID:    userID,
		Login:     login,
		CreatedAt: time.Now().Unix(),
	}
	b, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal session: %w", err)
	}
	if err := r.rdb.Set(ctx, r.key(id), b, r.ttl).Err(); err != nil {
		return "", fmt.Errorf("set session: %w", err)
	}
	return id, nil
}

func (r *SessionRepository) Get(ctx context.Context, id string) (SessionData, error) {
	b, err := r.rdb.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return SessionData{}, ErrSessionNotFound
		}
		return SessionData{}, fmt.Errorf("get session: %w", err)
	}
	var data SessionData
	if err := json.Unmarshal(b, &data); err != nil {
		return SessionData{}, fmt.Errorf("unmarshal session: %w", err)
	}
	return data, nil
}

func (r *SessionRepository) Touch(ctx context.Context, id string) error {
	return r.rdb.Expire(ctx, r.key(id), r.ttl).Err()
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	return r.rdb.Del(ctx, r.key(id)).Err()
}
