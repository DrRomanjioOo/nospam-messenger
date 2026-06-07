package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/spam"
)

type Job struct {
	MessageID int64
	Content   string
	Attempt   int
}

type MessageUpdater interface {
	ApplyAIDeletion(ctx context.Context, messageID int64) (domain.Message, error)
}

type Broadcaster interface {
	BroadcastMessageUpdated(domain.Message)
}

type SpamClassifier interface {
	CheckSpam(ctx context.Context, messageText string) (bool, string, error)
}

type AuditWriter interface {
	Insert(ctx context.Context, e domain.SpamAuditEntry) error
}

type AIPool struct {
	queue      chan Job
	classifier SpamClassifier
	audit      AuditWriter
	updater    MessageUpdater
	broadcast  Broadcaster
	retryDelay time.Duration
	model      string
	wg         sync.WaitGroup
}

func NewAIPool(
	workers int,
	queueSize int,
	classifier SpamClassifier,
	audit AuditWriter,
	updater MessageUpdater,
	broadcast Broadcaster,
	retryDelay time.Duration,
	model string,
) *AIPool {
	return &AIPool{
		queue:      make(chan Job, queueSize),
		classifier: classifier,
		audit:      audit,
		updater:    updater,
		broadcast:  broadcast,
		retryDelay: retryDelay,
		model:      model,
	}
}

func (p *AIPool) Start(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.loop(ctx)
		}()
	}
}

func (p *AIPool) Stop() {
	close(p.queue)
	p.wg.Wait()
}

func (p *AIPool) Enqueue(messageID int64, content string) {
	select {
	case p.queue <- Job{MessageID: messageID, Content: content}:
	default:
		slog.Warn("ai queue full, dropping job", "message_id", messageID)
	}
}

func (p *AIPool) loop(ctx context.Context) {
	for job := range p.queue {
		p.process(ctx, job)
	}
}

// ProcessForTest runs a single AI job synchronously (unit tests).
func (p *AIPool) ProcessForTest(ctx context.Context, job Job) {
	p.process(ctx, job)
}

func (p *AIPool) process(ctx context.Context, job Job) {
	start := time.Now()
	checkCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	isSpam, raw, err := p.classifier.CheckSpam(checkCtx, job.Content)
	latency := int(time.Since(start).Milliseconds())

	entry := domain.SpamAuditEntry{
		MessageID: job.MessageID,
		CheckType: "ai",
		Model:     p.model,
		LatencyMs: latency,
		RawResponse: raw,
	}

	if err != nil {
		entry.Verdict = "error"
		entry.ErrorText = err.Error()
		_ = p.audit.Insert(ctx, entry)
		if !shouldRetryAI(err) {
			slog.Error("ai check permanent failure", "message_id", job.MessageID, "err", err)
			return
		}
		slog.Warn("ai check failed, will retry", "message_id", job.MessageID, "err", err, "attempt", job.Attempt)
		go p.retryLater(ctx, job)
		return
	}

	if isSpam {
		entry.Verdict = "spam"
	} else {
		entry.Verdict = "ok"
	}
	_ = p.audit.Insert(ctx, entry)

	if !isSpam {
		return
	}

	updated, err := p.updater.ApplyAIDeletion(ctx, job.MessageID)
	if err != nil {
		slog.Error("ai soft delete failed", "message_id", job.MessageID, "err", err)
		return
	}
	p.broadcast.BroadcastMessageUpdated(updated)
}

func shouldRetryAI(err error) bool {
	if errors.Is(err, spam.ErrAINotConfigured) {
		return false
	}
	var apiErr *spam.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Permanent() {
			return false
		}
		return apiErr.Retryable()
	}
	return true
}

func (p *AIPool) retryDelayFor(job Job) time.Duration {
	delay := p.retryDelay
	for i := 0; i < job.Attempt && i < 6; i++ {
		delay *= 2
	}
	const maxDelay = 5 * time.Minute
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}

func (p *AIPool) retryLater(ctx context.Context, job Job) {
	timer := time.NewTimer(p.retryDelayFor(job))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		job.Attempt++
		if job.Attempt > 10 {
			slog.Error("ai check retries exhausted", "message_id", job.MessageID)
			return
		}
		select {
		case p.queue <- job:
		default:
			slog.Warn("ai queue full on retry", "message_id", job.MessageID)
		}
	}
}
