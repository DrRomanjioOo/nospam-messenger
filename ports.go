package repository_test

import (
	"context"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
)

func TestRateLimitRepository_AllowMessage(t *testing.T) {
	rdb, cleanup := setupRedis(t)
	defer cleanup()
	repo := repository.NewRateLimitRepository(rdb)
	ctx := context.Background()

	ok, err := repo.AllowMessage(ctx, 99)
	if err != nil || !ok {
		t.Fatalf("first: ok=%v err=%v", ok, err)
	}
	ok, err = repo.AllowMessage(ctx, 99)
	if err != nil || ok {
		t.Fatalf("second: ok=%v err=%v", ok, err)
	}
}
