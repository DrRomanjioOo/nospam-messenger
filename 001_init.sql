package spam_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/spam"
)

func TestExtractOpenRouterText(t *testing.T) {
	t.Parallel()
	raw, _ := json.Marshal(map[string]any{
		"choices": []map[string]any{{
			"message": map[string]string{"content": `{"is_spam":false}`},
		}},
	})
	text, err := spam.ExtractOpenRouterTextForTest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if text != `{"is_spam":false}` {
		t.Fatalf("got %q", text)
	}
}

func TestExtractOpenRouterText_EmptyChoices(t *testing.T) {
	_, err := spam.ExtractOpenRouterTextForTest([]byte(`{"choices":[]}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOpenRouterClient_CheckSpam_NoAPIKey(t *testing.T) {
	c := spam.NewOpenRouterClient("", "deepseek/deepseek-chat", "")
	if c.Enabled() {
		t.Fatal("expected disabled without key")
	}
	_, _, err := c.CheckSpam(context.Background(), "hello")
	if !errors.Is(err, spam.ErrAINotConfigured) {
		t.Fatalf("err=%v", err)
	}
}

func TestOpenRouterClient_Enabled_ValidKeyFormat(t *testing.T) {
	c := spam.NewOpenRouterClient("sk-or-v1-fakeKeyForUnitTestsOnly123456", "deepseek/deepseek-chat", "")
	if !c.Enabled() {
		t.Fatal("expected enabled for sk-or-v1- key")
	}
}

func TestIsValidOpenRouterAPIKey(t *testing.T) {
	if spam.IsValidOpenRouterAPIKey("sk-fakeKeyForUnitTestsOnly123456") {
		t.Fatal("unexpected valid for non-OpenRouter key")
	}
}

func TestOpenRouterClient_CheckSpam_SpamResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer sk-or-v1-testkey1234567890123456" {
			t.Fatalf("auth=%q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{"content": `{"is_spam":true,"reason":"ads"}`},
			}},
		})
	}))
	defer srv.Close()

	c := spam.NewOpenRouterClient("sk-or-v1-testkey1234567890123456", "deepseek/deepseek-chat", "check")
	c.SetHTTPClientForTest(srv.Client())
	c.SetBaseURLForTest(srv.URL)

	isSpam, raw, err := c.CheckSpam(context.Background(), "buy now")
	if err != nil {
		t.Fatal(err)
	}
	if !isSpam {
		t.Fatalf("expected spam, raw=%s", raw)
	}
}

func TestOpenRouterClient_CheckSpam_OKResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{"content": `{"is_spam":false,"reason":"fine"}`},
			}},
		})
	}))
	defer srv.Close()

	c := spam.NewOpenRouterClient("sk-or-v1-testkey1234567890123456", "deepseek/deepseek-chat", "prompt")
	c.SetHTTPClientForTest(srv.Client())
	c.SetBaseURLForTest(srv.URL)

	isSpam, _, err := c.CheckSpam(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if isSpam {
		t.Fatal("expected not spam")
	}
}
