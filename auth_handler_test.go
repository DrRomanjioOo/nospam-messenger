package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr           string
	DatabaseURL        string
	RedisURL           string
	SessionSecret      string
	SessionTTL         time.Duration
	CORSOrigin         string
	CookieSecure       bool
	OpenRouterAPIKey     string
	OpenRouterModel      string
	OpenRouterSpamPrompt string
	AIWorkerCount      int
	AIQueueSize        int
	AIRetryDelay       time.Duration
	BcryptCost         int
	StopWords          []string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:         envOr("HTTP_ADDR", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		RedisURL:         os.Getenv("REDIS_URL"),
		SessionSecret:      os.Getenv("SESSION_SECRET"),
		SessionTTL:       7 * 24 * time.Hour,
		CORSOrigin:       envOr("CORS_ORIGIN", "http://localhost:3000,http://127.0.0.1:3000"),
		CookieSecure:     envOr("COOKIE_SECURE", "false") == "true",
		OpenRouterAPIKey:     os.Getenv("OPENROUTER_API_KEY"),
		OpenRouterModel:      envOr("OPENROUTER_MODEL", "deepseek/deepseek-chat"),
		OpenRouterSpamPrompt: os.Getenv("OPENROUTER_SPAM_PROMPT"),
		AIWorkerCount:    envIntOr("AI_WORKER_COUNT", 4),
		AIQueueSize:      envIntOr("AI_QUEUE_SIZE", 256),
		AIRetryDelay:     time.Duration(envIntOr("AI_RETRY_DELAY_SEC", 30)) * time.Second,
		BcryptCost:       envIntOr("BCRYPT_COST", 12),
		StopWords:        defaultStopWords(),
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return cfg, fmt.Errorf("REDIS_URL is required")
	}
	if cfg.SessionSecret == "" {
		return cfg, fmt.Errorf("SESSION_SECRET is required")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func defaultStopWords() []string {
	// Hard slurs / insults — extend via STOP_WORDS env (comma-separated) later if needed.
	return []string{
		"fuck", "shit", "bitch", "asshole", "bastard",
		"сука", "блять", "блядь", "хуй", "пизда", "ебать", "ёбать",
	}
}
