package repository

import (
	"context"
	"time"

	"github.com/nghia/dockerai/backend/internal/domain"
)

type LogRepository interface {
	Insert(ctx context.Context, entry *domain.LogEntry) error
	InsertBatch(ctx context.Context, entries []*domain.LogEntry) error
	Query(ctx context.Context, containerID string, from, to time.Time, level domain.LogLevel, limit int) ([]*domain.LogEntry, error)
	ExistsByFingerprint(ctx context.Context, fingerprint string, since time.Time) (bool, error)
}

type AIRepository interface {
	Save(ctx context.Context, analysis *domain.AIAnalysis) error
	FindByFingerprint(ctx context.Context, fingerprint string, since time.Time) (*domain.AIAnalysis, error)
	List(ctx context.Context, limit int) ([]*domain.AIAnalysis, error)
}

type AIChatRepository interface {
	CreateSession(ctx context.Context, title string) (*domain.AIChatSession, error)
	GetSession(ctx context.Context, id string) (*domain.AIChatSession, error)
	AppendMessage(ctx context.Context, sessionID string, msg domain.AIChatMessage) error
	ListSessions(ctx context.Context, limit int) ([]*domain.AIChatSession, error)
	SetTitle(ctx context.Context, sessionID string, title string) error
}

type AlertRepository interface {
	Save(ctx context.Context, alert *domain.Alert) error
	UpdateStatus(ctx context.Context, id string, status domain.AlertStatus) error
	List(ctx context.Context, status domain.AlertStatus, limit int) ([]*domain.Alert, error)
	ExistsByFingerprintSince(ctx context.Context, fingerprint string, since time.Time) (bool, error)
}

type MetricsRepository interface {
	QueryRange(ctx context.Context, promql string, from, to time.Time, step time.Duration) ([]domain.MetricPoint, error)
	QueryInstant(ctx context.Context, promql string, ts time.Time) (float64, error)
}

type PredictionFeedbackRepository interface {
	Save(ctx context.Context, fb *domain.PredictionFeedback) error
	SaveChatFeedback(ctx context.Context, fb *domain.ChatPredictionFeedback) error
}
