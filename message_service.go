//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/repository"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/migrate"
)

func TestPostgresRepositories(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	if err := migrate.Up(url); err != nil {
		t.Fatal(err)
	}
	pg, err := repository.NewPostgres(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pg.Close()

	users := repository.NewUserRepository(pg)
	msgs := repository.NewMessageRepository(pg)
	audit := repository.NewSpamAuditRepository(pg)

	login := fmt.Sprintf("u%d", time.Now().UnixNano())
	u, err := users.Create(ctx, login, "hash")
	if err != nil {
		t.Fatal(err)
	}
	got, hash, err := users.GetByLogin(ctx, login)
	if err != nil || got.ID != u.ID || hash != "hash" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	byID, err := users.GetByID(ctx, u.ID)
	if err != nil || byID.Login != login {
		t.Fatalf("byID=%+v err=%v", byID, err)
	}

	m, err := msgs.Create(ctx, u.ID, "hello integration")
	if err != nil {
		t.Fatal(err)
	}
	m2, err := msgs.GetByID(ctx, m.ID)
	if err != nil || m2.Content != "hello integration" {
		t.Fatalf("m2=%+v err=%v", m2, err)
	}
	list, err := msgs.List(ctx, 0, 10)
	if err != nil || len(list) == 0 {
		t.Fatalf("list=%v err=%v", list, err)
	}
	deleted, err := msgs.SoftDeleteByUser(ctx, m.ID, u.ID)
	if err != nil || !deleted.DeletedByUser || deleted.Content != "" {
		t.Fatalf("deleted=%+v err=%v", deleted, err)
	}
	m3, err := msgs.Create(ctx, u.ID, "ai delete me")
	if err != nil {
		t.Fatal(err)
	}
	aiDel, err := msgs.SoftDeleteByAI(ctx, m3.ID)
	if err != nil || !aiDel.DeletedByAI {
		t.Fatalf("aiDel=%+v err=%v", aiDel, err)
	}
	withAuthor, err := msgs.AttachAuthor(ctx, aiDel)
	if err != nil || withAuthor.AuthorLogin != login {
		t.Fatalf("author=%+v err=%v", withAuthor, err)
	}
	if err := audit.Insert(ctx, domain.SpamAuditEntry{
		MessageID: m3.ID, CheckType: "ai", Verdict: "ok", Model: "test",
	}); err != nil {
		t.Fatal(err)
	}
}
