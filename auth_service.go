package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
)

var ErrMessageNotFound = errors.New("message not found")

type MessageRepository struct {
	pg *Postgres
}

func NewMessageRepository(pg *Postgres) *MessageRepository {
	return &MessageRepository{pg: pg}
}

func (r *MessageRepository) Create(ctx context.Context, userID int64, content string) (domain.Message, error) {
	const q = `
		INSERT INTO messages (user_id, content)
		VALUES ($1, $2)
		RETURNING id, user_id, content, created_at, deleted_by_user, deleted_by_ai`

	var m domain.Message
	err := r.pg.pool.QueryRow(ctx, q, userID, content).Scan(
		&m.ID, &m.UserID, &m.Content, &m.CreatedAt, &m.DeletedByUser, &m.DeletedByAI,
	)
	if err != nil {
		return domain.Message{}, fmt.Errorf("insert message: %w", err)
	}
	return m, nil
}

func (r *MessageRepository) GetByID(ctx context.Context, id int64) (domain.Message, error) {
	const q = `
		SELECT m.id, m.user_id, u.login, m.content, m.created_at, m.deleted_by_user, m.deleted_by_ai
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = $1`

	var m domain.Message
	err := r.pg.pool.QueryRow(ctx, q, id).Scan(
		&m.ID, &m.UserID, &m.AuthorLogin, &m.Content, &m.CreatedAt, &m.DeletedByUser, &m.DeletedByAI,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Message{}, ErrMessageNotFound
		}
		return domain.Message{}, fmt.Errorf("get message: %w", err)
	}
	return m, nil
}

func (r *MessageRepository) List(ctx context.Context, beforeID int64, limit int) ([]domain.Message, error) {
	if limit <= 0 || limit > 50 {
		limit = 50
	}

	var rows pgx.Rows
	var err error

	if beforeID > 0 {
		const q = `
			SELECT m.id, m.user_id, u.login, m.content, m.created_at, m.deleted_by_user, m.deleted_by_ai
			FROM messages m
			JOIN users u ON u.id = m.user_id
			WHERE m.id < $1
			ORDER BY m.id DESC
			LIMIT $2`
		rows, err = r.pg.pool.Query(ctx, q, beforeID, limit)
	} else {
		const q = `
			SELECT m.id, m.user_id, u.login, m.content, m.created_at, m.deleted_by_user, m.deleted_by_ai
			FROM messages m
			JOIN users u ON u.id = m.user_id
			ORDER BY m.id DESC
			LIMIT $1`
		rows, err = r.pg.pool.Query(ctx, q, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var out []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.AuthorLogin, &m.Content, &m.CreatedAt, &m.DeletedByUser, &m.DeletedByAI,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MessageRepository) SoftDeleteByUser(ctx context.Context, id, userID int64) (domain.Message, error) {
	const q = `
		UPDATE messages
		SET content = '', deleted_by_user = TRUE
		WHERE id = $1 AND user_id = $2 AND deleted_by_user = FALSE AND deleted_by_ai = FALSE
		RETURNING id, user_id, content, created_at, deleted_by_user, deleted_by_ai`

	var m domain.Message
	err := r.pg.pool.QueryRow(ctx, q, id, userID).Scan(
		&m.ID, &m.UserID, &m.Content, &m.CreatedAt, &m.DeletedByUser, &m.DeletedByAI,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Message{}, ErrMessageNotFound
		}
		return domain.Message{}, fmt.Errorf("soft delete by user: %w", err)
	}
	return m, nil
}

func (r *MessageRepository) SoftDeleteByAI(ctx context.Context, id int64) (domain.Message, error) {
	const q = `
		UPDATE messages
		SET content = '', deleted_by_ai = TRUE
		WHERE id = $1 AND deleted_by_ai = FALSE
		RETURNING id, user_id, content, created_at, deleted_by_user, deleted_by_ai`

	var m domain.Message
	err := r.pg.pool.QueryRow(ctx, q, id).Scan(
		&m.ID, &m.UserID, &m.Content, &m.CreatedAt, &m.DeletedByUser, &m.DeletedByAI,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Message{}, ErrMessageNotFound
		}
		return domain.Message{}, fmt.Errorf("soft delete by ai: %w", err)
	}
	return m, nil
}

func (r *MessageRepository) AttachAuthor(ctx context.Context, m domain.Message) (domain.Message, error) {
	u, err := NewUserRepository(r.pg).GetByID(ctx, m.UserID)
	if err != nil {
		return domain.Message{}, err
	}
	m.AuthorLogin = u.Login
	return m, nil
}
