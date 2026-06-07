package spam

import "context"

// NoopClassifier skips AI moderation (always not spam).
type NoopClassifier struct{}

func (NoopClassifier) CheckSpam(_ context.Context, _ string) (bool, string, error) {
	return false, "", nil
}
