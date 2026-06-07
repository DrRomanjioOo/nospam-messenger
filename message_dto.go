package domain_test

import (
	"testing"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
)

func TestMessage_IsDeleted(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		m    domain.Message
		want bool
	}{
		{"active", domain.Message{}, false},
		{"user", domain.Message{DeletedByUser: true}, true},
		{"ai", domain.Message{DeletedByAI: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.m.IsDeleted(); got != tc.want {
				t.Fatalf("IsDeleted() = %v, want %v", got, tc.want)
			}
		})
	}
}
