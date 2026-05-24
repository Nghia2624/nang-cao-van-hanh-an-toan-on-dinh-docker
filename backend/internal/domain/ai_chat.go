package domain

import "time"

type AIChatRole string

const (
	AIChatRoleUser      AIChatRole = "user"
	AIChatRoleAssistant AIChatRole = "assistant"
)

type AIChatMessage struct {
	Role      AIChatRole `json:"role"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`

	// Optional structured context
	ContainerID string `json:"container_id,omitempty"`
	FromMillis  int64  `json:"from,omitempty"`
	ToMillis    int64  `json:"to,omitempty"`

	// Optional attachments
	Evidence []AIChatEvidence `json:"evidence,omitempty"`
}

type AIChatEvidence struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	ContainerID string    `json:"container_id,omitempty"`
}

type AIChatSession struct {
	ID        string          `json:"id,omitempty"`
	Title     string          `json:"title"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Messages  []AIChatMessage `json:"messages"`
	ExpireAt  time.Time       `json:"-"`
}
