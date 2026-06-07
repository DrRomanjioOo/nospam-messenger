package spam

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultOpenRouterURL = "https://openrouter.ai/api/v1/chat/completions"

var ErrAINotConfigured = errors.New("openrouter api key not configured")

type OpenRouterClient struct {
	apiKey  string
	model   string
	prompt  string
	client  *http.Client
	baseURL string
}

func NewOpenRouterClient(apiKey, model, prompt string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey: strings.TrimSpace(apiKey),
		model:  model,
		prompt: prompt,
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

// Enabled reports whether a usable OpenRouter API key is configured.
func (o *OpenRouterClient) Enabled() bool {
	return IsValidOpenRouterAPIKey(o.apiKey)
}

func (o *OpenRouterClient) SetHTTPClientForTest(c *http.Client) { o.client = c }
func (o *OpenRouterClient) SetBaseURLForTest(u string)         { o.baseURL = u }

// ExtractOpenRouterTextForTest exposes response parsing for unit tests.
func ExtractOpenRouterTextForTest(raw []byte) (string, error) {
	return extractOpenRouterText(raw)
}

type aiVerdict struct {
	IsSpam bool   `json:"is_spam"`
	Reason string `json:"reason"`
}

// CheckSpam returns isSpam, raw model text, error.
func (o *OpenRouterClient) CheckSpam(ctx context.Context, messageText string) (bool, string, error) {
	if !o.Enabled() {
		return false, "", ErrAINotConfigured
	}

	systemPrompt := o.prompt
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = `You are a spam moderator for a chat. Reply ONLY with JSON: {"is_spam":true|false,"reason":"..."}.
Mark spam if the message is advertising, scam, harassment, or obvious bot flood. Otherwise is_spam=false.`
	}

	body := map[string]any{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "Message:\n" + messageText},
		},
		"temperature": 0,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}
	b, _ := json.Marshal(body)

	url := o.baseURL
	if url == "" {
		url = defaultOpenRouterURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/noname-group7630520/nospam-messenger")
	req.Header.Set("X-Title", "nospam-messenger")

	resp, err := o.client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return false, string(raw), &APIError{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	text, err := extractOpenRouterText(raw)
	if err != nil {
		return false, string(raw), err
	}

	var verdict aiVerdict
	if err := json.Unmarshal([]byte(text), &verdict); err != nil {
		lower := strings.ToLower(text)
		if strings.Contains(lower, `"is_spam":true`) || strings.Contains(lower, "is_spam: true") {
			return true, text, nil
		}
		return false, text, fmt.Errorf("parse verdict: %w", err)
	}
	return verdict.IsSpam, text, nil
}

func extractOpenRouterText(raw []byte) (string, error) {
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty openrouter response")
	}
	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("empty openrouter response")
	}
	return text, nil
}
