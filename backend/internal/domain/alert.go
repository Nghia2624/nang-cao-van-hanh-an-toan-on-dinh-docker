package domain

import "time"

type AlertStatus string

const (
	AlertStatusNew         AlertStatus = "NEW"
	AlertStatusAcknowledged AlertStatus = "ACKNOWLEDGED"
	AlertStatusResolved    AlertStatus = "RESOLVED"
)

type Alert struct {
	ID          string      `json:"id,omitempty"`
	Source      string      `json:"source"` // metrics|logs|ai
	Severity    string      `json:"severity"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      AlertStatus `json:"status"`
	ContainerID string      `json:"container_id,omitempty"`
	Fingerprint string      `json:"fingerprint,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
