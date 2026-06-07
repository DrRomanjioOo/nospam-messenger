package repository

import (
	"context"
	"fmt"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
)

type SpamAuditRepository struct {
	pg *Postgres
}

func NewSpamAuditRepository(pg *Postgres) *SpamAuditRepository {
	return &SpamAuditRepository{pg: pg}
}

func (r *SpamAuditRepository) Insert(ctx context.Context, e domain.SpamAuditEntry) error {
	const q = `
		INSERT INTO spam_audit (message_id, check_type, verdict, model, raw_response, latency_ms, error_text)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pg.pool.Exec(ctx, q,
		e.MessageID, e.CheckType, e.Verdict, e.Model, e.RawResponse, e.LatencyMs, e.ErrorText,
	)
	if err != nil {
		return fmt.Errorf("insert spam audit: %w", err)
	}
	return nil
}
