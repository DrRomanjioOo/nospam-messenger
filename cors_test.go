package handler

import (
	"time"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/domain"
)

type messageResponse struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	AuthorLogin   string    `json:"author_login"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	DeletedByUser bool      `json:"deleted_by_user"`
	DeletedByAI   bool      `json:"deleted_by_ai"`
}

func toMessageResponse(m domain.Message) messageResponse {
	return messageResponse{
		ID:            m.ID,
		UserID:        m.UserID,
		AuthorLogin:   m.AuthorLogin,
		Content:       m.Content,
		CreatedAt:     m.CreatedAt,
		DeletedByUser: m.DeletedByUser,
		DeletedByAI:   m.DeletedByAI,
	}
}

func toMessageList(msgs []domain.Message) []messageResponse {
	out := make([]messageResponse, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, toMessageResponse(m))
	}
	return out
}
