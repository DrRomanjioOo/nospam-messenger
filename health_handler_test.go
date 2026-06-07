package domain

import "time"

type User struct {
	ID        int64
	Login     string
	CreatedAt time.Time
}

type Message struct {
	ID             int64
	UserID         int64
	AuthorLogin    string
	Content        string
	CreatedAt      time.Time
	DeletedByUser  bool
	DeletedByAI    bool
}

func (m Message) IsDeleted() bool {
	return m.DeletedByUser || m.DeletedByAI
}

type SpamAuditEntry struct {
	MessageID   int64
	CheckType   string
	Verdict     string
	Model       string
	RawResponse string
	LatencyMs   int
	ErrorText   string
}
