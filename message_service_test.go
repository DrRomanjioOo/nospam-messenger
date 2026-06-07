package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitRepository struct {
	rdb *redis.Client
}

func NewRateLimitRepository(rdb *redis.Client) *RateLimitRepository {
	return &RateLimitRepository{rdb: rdb}
}

func (r *RateLimitRepository) key(userID int64) string {
	return fmt.Sprintf("rate:msg:%d", userID)
}

// AllowMessage returns false if user sent a message within the last second.
func (r *RateLimitRepository) AllowMessage(ctx context.Context, userID int64) (bool, error) {
	ok, err := r.rdb.SetNX(ctx, r.key(userID), "1", time.Second).Result()
	if err != nil {
		return false, fmt.Errorf("rate limit: %w", err)
	}
	return ok, nil
}
