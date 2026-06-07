package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/service"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/testutil"
)

func TestAuthService_Register_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := &testutil.MockUserStore{
		CreateFn: func(ctx context.Context, login, hash string) (domain.User, error) {
			return domain.User{ID: 1, Login: login, CreatedAt: time.Now()}, nil
		},
	}
	sessions := &testutil.MockSessionStore{
		CreateFn: func(ctx context.Context, userID int64, login string) (string, error) {
			return "sid-1", nil
		},
	}
	svc := service.NewAuthService(users, sessions, bcrypt.MinCost)

	u, sid, err := svc.Register(ctx, "alice1", "password1")
	if err != nil {
		t.Fatal(err)
	}
	if u.Login != "alice1" || sid != "sid-1" {
		t.Fatalf("got %+v %q", u, sid)
	}
}

func TestAuthService_Register_Validation(t *testing.T) {
	t.Parallel()
	svc := service.NewAuthService(&testutil.MockUserStore{}, &testutil.MockSessionStore{}, bcrypt.MinCost)
	_, _, err := svc.Register(context.Background(), "ab", "password1")
	if !errors.Is(err, service.ErrInvalidLogin) {
		t.Fatalf("got %v", err)
	}
	_, _, err = svc.Register(context.Background(), "alice1", "short")
	if !errors.Is(err, service.ErrInvalidPassword) {
		t.Fatalf("got %v", err)
	}
}

func TestAuthService_Register_LoginTaken(t *testing.T) {
	svc := service.NewAuthService(&testutil.MockUserStore{
		CreateFn: func(ctx context.Context, login, hash string) (domain.User, error) {
			return domain.User{}, repository.ErrLoginTaken
		},
	}, &testutil.MockSessionStore{}, bcrypt.MinCost)
	_, _, err := svc.Register(context.Background(), "alice1", "password1")
	if !errors.Is(err, service.ErrLoginTaken) {
		t.Fatalf("got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.MinCost)
	svc := service.NewAuthService(&testutil.MockUserStore{
		GetByLoginFn: func(ctx context.Context, login string) (domain.User, string, error) {
			return domain.User{ID: 2, Login: login}, string(hash), nil
		},
	}, &testutil.MockSessionStore{
		CreateFn: func(ctx context.Context, userID int64, login string) (string, error) {
			return "sid-2", nil
		},
	}, bcrypt.MinCost)

	_, sid, err := svc.Login(context.Background(), "alice1", "password1")
	if err != nil || sid != "sid-2" {
		t.Fatalf("err=%v sid=%q", err, sid)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	svc := service.NewAuthService(&testutil.MockUserStore{
		GetByLoginFn: func(ctx context.Context, login string) (domain.User, string, error) {
			return domain.User{}, "", repository.ErrUserNotFound
		},
	}, &testutil.MockSessionStore{}, bcrypt.MinCost)
	_, _, err := svc.Login(context.Background(), "alice1", "password1")
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Fatalf("got %v", err)
	}
}

func TestAuthService_ResolveSession(t *testing.T) {
	touched := false
	svc := service.NewAuthService(nil, &testutil.MockSessionStore{
		GetFn: func(ctx context.Context, id string) (repository.SessionData, error) {
			return repository.SessionData{UserID: 5, Login: "bob1"}, nil
		},
		TouchFn: func(ctx context.Context, id string) error {
			touched = true
			return nil
		},
	}, bcrypt.MinCost)
	data, err := svc.ResolveSession(context.Background(), "sid")
	if err != nil || data.UserID != 5 || !touched {
		t.Fatalf("data=%+v touched=%v err=%v", data, touched, err)
	}
}
