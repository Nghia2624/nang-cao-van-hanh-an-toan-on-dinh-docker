package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nghia/dockerai/backend/internal/domain"
	"github.com/nghia/dockerai/backend/internal/pkg/logger"
	"github.com/nghia/dockerai/backend/internal/repository"
)

type alertEngine struct {
	repo repository.AlertRepository
	log  *logger.Logger
}

func NewAlertEngine(repo repository.AlertRepository, log *logger.Logger) AlertEngine {
	return &alertEngine{repo: repo, log: log}
}

func (e *alertEngine) EvaluateMetrics(ctx context.Context, metrics *domain.HealthMetrics, container domain.Container) (*domain.Alert, error) {
	if metrics == nil {
		return nil, nil
	}
	if metrics.Status == "CRITICAL" || metrics.Status == "UNHEALTHY" {
		alert := &domain.Alert{
			Source:      "metrics",
			Severity:    "HIGH",
			Title:       fmt.Sprintf("Container %s health %s", container.Name, metrics.Status),
			Description: metrics.Summary,
			Status:      domain.AlertStatusNew,
			ContainerID: container.ID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := e.repo.Save(ctx, alert); err != nil {
			e.log.Errorw("failed to save metrics alert", "error", err, "container", container.Name)
		}
		return alert, nil
	}
	return nil, nil
}

func (e *alertEngine) EvaluateLog(ctx context.Context, logEntry *domain.LogEntry) (*domain.Alert, error) {
	if logEntry.Level == domain.LogLevelError || logEntry.Level == domain.LogLevelFatal {
		// Dedup: skip if same fingerprint already alerted recently
		since := time.Now().Add(-30 * time.Minute)
		exists, err := e.repo.ExistsByFingerprintSince(ctx, logEntry.Fingerprint, since)
		if err == nil && exists {
			return nil, nil
		}
		alert := &domain.Alert{
			Source:      "logs",
			Severity:    "HIGH",
			Title:       "Error log detected",
			Description: logEntry.Message,
			Status:      domain.AlertStatusNew,
			ContainerID: logEntry.ContainerID,
			Fingerprint: logEntry.Fingerprint,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := e.repo.Save(ctx, alert); err != nil {
			e.log.Errorw("failed to save log alert", "error", err, "fingerprint", logEntry.Fingerprint)
		}
		return alert, nil
	}
	return nil, nil
}

func (e *alertEngine) EvaluateAI(ctx context.Context, analysis *domain.AIAnalysis) (*domain.Alert, error) {
	if analysis == nil {
		return nil, nil
	}
	if analysis.Severity >= 4 {
		// Dedup: skip if same fingerprint already alerted recently
		if analysis.Fingerprint != "" {
			since := time.Now().Add(-30 * time.Minute)
			exists, err := e.repo.ExistsByFingerprintSince(ctx, analysis.Fingerprint, since)
			if err == nil && exists {
				return nil, nil
			}
		}
		alert := &domain.Alert{
			Source:      "ai",
			Severity:    "CRITICAL",
			Title:       "AI detected high severity issue",
			Description: analysis.RootCause,
			Status:      domain.AlertStatusNew,
			Fingerprint: analysis.Fingerprint,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := e.repo.Save(ctx, alert); err != nil {
			e.log.Errorw("failed to save AI alert", "error", err, "fingerprint", analysis.Fingerprint)
		}
		return alert, nil
	}
	return nil, nil
}
