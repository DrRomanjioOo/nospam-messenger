package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/service"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/spam"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/testutil"
)

func newMessageService() (*service.MessageService, *testutil.MockMessageStore, *testutil.MockAIEnqueuer) {
	msgs := &testutil.MockMessageStore{}
	enq := &testutil.MockAIEnqueuer{}
	svc := service.NewMessageService(
		msgs,
		&testutil.MockRateLimiter{AllowFn: func(ctx context.Context, userID int64) (bool, error) {
			return true, nil
		}},
		spam.NewRuleChecker(nil),
		enq,
	)
	return svc, msgs, enq
}

func TestMessageService_Send_Success(t *testing.T) {
	svc, msgs, enq := newMessageService()
	var enqueued bool
	enq.EnqueueFn = func(id int64, content string) { enqueued = true }
	msgs.CreateFn = func(ctx context.Context, userID int64, content string) (domain.Message, error) {
		return domain.Message{ID: 10, UserID: userID, Content: content, CreatedAt: time.Now()}, nil
	}
	m, err := svc.Send(context.Background(), 1, "alice1", "hello")
	if err != nil || m.ID != 10 || !enqueued {
		t.Fatalf("m=%+v enqueued=%v err=%v", m, enqueued, err)
	}
}

func TestMessageService_Send_TooLong(t *testing.T) {
	svc, _, _ := newMessageService()
	_, err := svc.Send(context.Background(), 1, "a", strings.Repeat("x", 2001))
	if !errors.Is(err, service.ErrMessageTooLong) {
		t.Fatalf("got %v", err)
	}
}

func TestMessageService_Send_SpamRule(t *testing.T) {
	svc, _, _ := newMessageService()
	_, err := svc.Send(context.Background(), 1, "a", "   ")
	if !errors.Is(err, spam.ErrSpamRule) {
		t.Fatalf("got %v", err)
	}
}

func TestMessageService_Send_RateLimited(t *testing.T) {
	svc := service.NewMessageService(
		&testutil.MockMessageStore{},
		&testutil.MockRateLimiter{AllowFn: func(ctx context.Context, userID int64) (bool, error) {
			return false, nil
		}},
		spam.NewRuleChecker(nil),
		nil,
	)
	_, err := svc.Send(context.Background(), 1, "a", "ok")
	if !errors.Is(err, spam.ErrRateLimited) {
		t.Fatalf("got %v", err)
	}
}

func TestMessageService_List(t *testing.T) {
	svc, msgs, _ := newMessageService()
	want := []domain.Message{{ID: 1}}
	msgs.ListFn = func(ctx context.Context, beforeID int64, limit int) ([]domain.Message, error) {
		return want, nil
	}
	got, err := svc.List(context.Background(), 0, 50)
	if err != nil || len(got) != 1 {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestMessageService_DeleteByUser(t *testing.T) {
	svc, msgs, _ := newMessageService()
	msgs.SoftDeleteByUserFn = func(ctx context.Context, id, userID int64) (domain.Message, error) {
		return domain.Message{ID: id, UserID: userID, DeletedByUser: true}, nil
	}
	msgs.AttachAuthorFn = func(ctx context.Context, m domain.Message) (domain.Message, error) {
		m.AuthorLogin = "alice1"
		return m, nil
	}
	m, err := svc.DeleteByUser(context.Background(), 3, 1)
	if err != nil || !m.DeletedByUser || m.AuthorLogin != "alice1" {
		t.Fatalf("m=%+v err=%v", m, err)
	}
}

func TestMessageService_ApplyAIDeletion(t *testing.T) {
	svc, msgs, _ := newMessageService()
	msgs.SoftDeleteByAIFn = func(ctx context.Context, id int64) (domain.Message, error) {
		return domain.Message{ID: id, DeletedByAI: true}, nil
	}
	msgs.AttachAuthorFn = func(ctx context.Context, m domain.Message) (domain.Message, error) {
		m.AuthorLogin = "bob1"
		return m, nil
	}
	m, err := svc.ApplyAIDeletion(context.Background(), 7)
	if err != nil || !m.DeletedByAI {
		t.Fatalf("m=%+v err=%v", m, err)
	}
}

func TestMessageService_Get(t *testing.T) {
	svc, msgs, _ := newMessageService()
	msgs.GetByIDFn = func(ctx context.Context, id int64) (domain.Message, error) {
		return domain.Message{ID: id}, nil
	}
	m, err := svc.Get(context.Background(), 9)
	if err != nil || m.ID != 9 {
		t.Fatalf("m=%+v err=%v", m, err)
	}
}

func TestMessageService_DeleteByUser_NotFound(t *testing.T) {
	svc, msgs, _ := newMessageService()
	msgs.SoftDeleteByUserFn = func(ctx context.Context, id, userID int64) (domain.Message, error) {
		return domain.Message{}, repository.ErrMessageNotFound
	}
	_, err := svc.DeleteByUser(context.Background(), 1, 1)
	if !errors.Is(err, repository.ErrMessageNotFound) {
		t.Fatalf("got %v", err)
	}
}
