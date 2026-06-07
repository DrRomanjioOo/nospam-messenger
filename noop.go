package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
)

func setupRedis(t *testing.T) (*redis.Client, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return rdb, func() {
		_ = rdb.Close()
		mr.Close()
	}
}

func TestSessionRepository_CreateGetTouchDelete(t *testing.T) {
	rdb, cleanup := setupRedis(t)
	defer cleanup()
	repo := repository.NewSessionRepository(rdb, time.Hour)
	ctx := context.Background()

	id, err := repo.Create(ctx, 42, "alice1")
	if err != nil {
		t.Fatal(err)
	}
	data, err := repo.Get(ctx, id)
	if err != nil || data.UserID != 42 || data.Login != "alice1" {
		t.Fatalf("data=%+v err=%v", data, err)
	}
	if err := repo.Touch(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	_, err = repo.Get(ctx, id)
	if !errors.Is(err, repository.ErrSessionNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionRepository_Get_NotFound(t *testing.T) {
	rdb, cleanup := setupRedis(t)
	defer cleanup()
	repo := repository.NewSessionRepository(rdb, time.Hour)
	_, err := repo.Get(context.Background(), "missing")
	if !errors.Is(err, repository.ErrSessionNotFound) {
		t.Fatalf("err=%v", err)
	}
}
