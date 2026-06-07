package config_test

import (
	"os"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/config"
)

func TestLoad_MissingRequired(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("SESSION_SECRET", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_Success(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("SESSION_SECRET", "secret")
	t.Setenv("CORS_ORIGIN", "http://localhost:3000")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.CORSOrigin != "http://localhost:3000" {
		t.Fatalf("cfg=%+v", cfg)
	}
	if len(cfg.StopWords) == 0 {
		t.Fatal("expected default stop words")
	}
}

func TestEnvIntOr_InvalidUsesFallback(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("REDIS_URL", "redis://x")
	t.Setenv("SESSION_SECRET", "s")
	t.Setenv("AI_WORKER_COUNT", "not-a-number")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AIWorkerCount != 4 {
		t.Fatalf("got %d", cfg.AIWorkerCount)
	}
	_ = os.Unsetenv("AI_WORKER_COUNT")
}
