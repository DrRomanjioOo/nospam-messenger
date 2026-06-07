package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")
var ErrLoginTaken = errors.New("login taken")

type UserRepository struct {
	pg *Postgres
}

func NewUserRepository(pg *Postgres) *UserRepository {
	return &UserRepository{pg: pg}
}

func (r *UserRepository) Create(ctx context.Context, login, passwordHash string) (domain.User, error) {
	const q = `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, login, created_at`

	var u domain.User
	err := r.pg.pool.QueryRow(ctx, q, login, passwordHash).Scan(&u.ID, &u.Login, &u.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, ErrLoginTaken
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (domain.User, string, error) {
	const q = `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1`

	var u domain.User
	var hash string
	err := r.pg.pool.QueryRow(ctx, q, login).Scan(&u.ID, &u.Login, &hash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, "", ErrUserNotFound
		}
		return domain.User{}, "", fmt.Errorf("get user by login: %w", err)
	}
	return u, hash, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (domain.User, error) {
	const q = `SELECT id, login, created_at FROM users WHERE id = $1`

	var u domain.User
	err := r.pg.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Login, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}
