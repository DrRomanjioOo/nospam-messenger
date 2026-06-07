package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidLogin       = errors.New("invalid login")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrLoginTaken         = errors.New("login taken")
)

var loginPattern = regexp.MustCompile(`^[a-zA-Z0-9]{4,16}$`)

type AuthService struct {
	users      UserStore
	sessions   SessionStore
	bcryptCost int
}

func NewAuthService(users UserStore, sessions SessionStore, bcryptCost int) *AuthService {
	return &AuthService{users: users, sessions: sessions, bcryptCost: bcryptCost}
}

func (s *AuthService) Register(ctx context.Context, login, password string) (domain.User, string, error) {
	if err := validateLogin(login); err != nil {
		return domain.User{}, "", err
	}
	if err := validatePassword(password); err != nil {
		return domain.User{}, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("hash password: %w", err)
	}

	u, err := s.users.Create(ctx, login, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrLoginTaken) {
			return domain.User{}, "", ErrLoginTaken
		}
		return domain.User{}, "", err
	}

	sid, err := s.sessions.Create(ctx, u.ID, u.Login)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, sid, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (domain.User, string, error) {
	if err := validateLogin(login); err != nil {
		return domain.User{}, "", ErrInvalidCredentials
	}
	if err := validatePassword(password); err != nil {
		return domain.User{}, "", ErrInvalidCredentials
	}

	u, hash, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.User{}, "", ErrInvalidCredentials
		}
		return domain.User{}, "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return domain.User{}, "", ErrInvalidCredentials
	}

	sid, err := s.sessions.Create(ctx, u.ID, u.Login)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, sid, nil
}

func (s *AuthService) ResolveSession(ctx context.Context, sessionID string) (repository.SessionData, error) {
	data, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return repository.SessionData{}, err
	}
	_ = s.sessions.Touch(ctx, sessionID)
	return data, nil
}

func validateLogin(login string) error {
	if !loginPattern.MatchString(login) {
		return ErrInvalidLogin
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}
