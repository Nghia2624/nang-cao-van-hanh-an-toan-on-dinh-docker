package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/nghia/dockerai/backend/internal/domain"
	"github.com/nghia/dockerai/backend/internal/pkg/logger"
	"github.com/nghia/dockerai/backend/internal/repository"
)

type logProcessor struct {
	repo      repository.LogRepository
	ai        AIAnalyzer
	alerts    AlertEngine
	log       *logger.Logger
	sse       SSEBroadcaster
	aiQueue   chan *domain.LogEntry
	maxQueue  int
	samplePct int // percentage for INFO
	stopCh    chan struct{}
}

type SSEBroadcaster interface {
	Broadcast(msg string)
}

func NewLogProcessor(repo repository.LogRepository, ai AIAnalyzer, alerts AlertEngine, log *logger.Logger, sse SSEBroadcaster) LogProcessor {
	lp := &logProcessor{
		repo:      repo,
		ai:        ai,
		alerts:    alerts,
		log:       log,
		sse:       sse,
		aiQueue:   make(chan *domain.LogEntry, 200),
		maxQueue:  200,
		samplePct: 10,
		stopCh:    make(chan struct{}),
	}
	go lp.runQueue()
	return lp
}

// Stop gracefully stops the log processor, flushing any remaining batch.
func (p *logProcessor) Stop() {
	close(p.stopCh)
}

func (p *logProcessor) ProcessLog(ctx context.Context, logEntry *domain.LogEntry) error {
	// Prevent recursion: ignore logs from our own components
	if p.isInternalLog(logEntry) {
		return nil
	}
	logEntry.Fingerprint = fingerprint(logEntry)
	if !p.shouldStore(logEntry) {
		return nil
	}
	// Deduplication: check if we already have this fingerprint recently
	if p.isDuplicate(ctx, logEntry.Fingerprint) {
		return nil
	}
	if err := p.repo.Insert(ctx, logEntry); err != nil {
		return err
	}
	if p.ShouldAnalyzeWithAI(logEntry) {
		_ = p.QueueForAI(logEntry)
	}
	if alert, err := p.alerts.EvaluateLog(ctx, logEntry); err == nil && alert != nil {
		p.log.Infow("alert from log", "alert", alert)
		p.broadcastEvent("alert", alert)
	}
	return nil
}

func (p *logProcessor) ProcessBatch(ctx context.Context, logs []*domain.LogEntry) error {
	for _, l := range logs {
		_ = p.ProcessLog(ctx, l)
	}
	return nil
}

func (p *logProcessor) ShouldAnalyzeWithAI(log *domain.LogEntry) bool {
	if log.Level == domain.LogLevelError || log.Level == domain.LogLevelFatal {
		return true
	}
	return false
}

func (p *logProcessor) QueueForAI(log *domain.LogEntry) error {
	select {
	case p.aiQueue <- log:
	default:
		p.log.Warn("ai queue full, dropping log")
	}
	return nil
}

func (p *logProcessor) runQueue() {
	batch := make([]*domain.LogEntry, 0, 20)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case l := <-p.aiQueue:
			batch = append(batch, l)
			if len(batch) >= 20 {
				p.flushBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				p.flushBatch(batch)
				batch = batch[:0]
			}
		case <-p.stopCh:
			// Flush remaining batch before exiting
			if len(batch) > 0 {
				p.flushBatch(batch)
			}
			p.log.Info("log processor queue stopped")
			return
		}
	}
}

func (p *logProcessor) flushBatch(batch []*domain.LogEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if len(batch) == 0 {
		return
	}
	if len(batch) == 1 {
		analysis, err := p.ai.AnalyzeError(ctx, batch[0], nil)
		if err != nil {
			p.log.Warnw("ai analysis failed", "error", err, "log", batch[0].Message[:min(100, len(batch[0].Message))])
		} else if analysis != nil {
			p.log.Infow("ai analysis completed", "fingerprint", analysis.Fingerprint, "severity", analysis.SeverityLabel)
			p.broadcastEvent("ai_insight", analysis)
			if alert, err := p.alerts.EvaluateAI(ctx, analysis); err != nil {
				p.log.Warnw("failed to evaluate AI alert", "error", err)
			} else if alert != nil {
				p.broadcastEvent("alert", alert)
			}
		}
	} else {
		analysis, err := p.ai.AnalyzeLogs(ctx, batch)
		if err != nil {
			p.log.Warnw("ai batch analysis failed", "error", err, "count", len(batch))
		} else if analysis != nil {
			p.log.Infow("ai batch analysis completed", "fingerprint", analysis.Fingerprint, "severity", analysis.SeverityLabel, "logs", len(batch))
			p.broadcastEvent("ai_insight", analysis)
			if alert, err := p.alerts.EvaluateAI(ctx, analysis); err != nil {
				p.log.Warnw("failed to evaluate AI alert", "error", err)
			} else if alert != nil {
				p.broadcastEvent("alert", alert)
			}
		}
	}
}

func (p *logProcessor) broadcastEvent(eventType string, payload interface{}) {
	if p.sse == nil {
		return
	}
	b, err := json.Marshal(map[string]interface{}{
		"type":    eventType,
		"payload": payload,
		"ts":      time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return
	}
	p.sse.Broadcast(string(b))
}

func (p *logProcessor) shouldStore(log *domain.LogEntry) bool {
	if log.Level == domain.LogLevelError || log.Level == domain.LogLevelFatal || log.Level == domain.LogLevelWarn {
		return true
	}
	if log.Level == domain.LogLevelInfo {
		hash := sha1.Sum([]byte(log.Message))
		val := int(hash[0]) % 100
		return val < p.samplePct
	}
	return false
}

func (p *logProcessor) isInternalLog(entry *domain.LogEntry) bool {
	// Skip logs containing our own component names / generated alert markers to prevent recursion
	msg := strings.ToLower(entry.Message)
	return strings.Contains(msg, "alert from log") ||
		strings.Contains(msg, "service/log_processor_impl.go") ||
		strings.Contains(msg, "service/alert_engine_impl.go") ||
		strings.Contains(msg, "service/ai_analyzer_impl.go") ||
		strings.Contains(msg, "service/docker_service_impl.go") ||
		strings.Contains(msg, "middleware/logger.go") ||
		strings.Contains(msg, "middleware/rate_limiter.go") ||
		strings.Contains(msg, "context canceled") ||
		strings.Contains(msg, "context deadline exceeded")
}

func (p *logProcessor) isDuplicate(ctx context.Context, fingerprint string) bool {
	if fingerprint == "" {
		return false
	}
	since := time.Now().Add(-5 * time.Minute) // dedup window
	exists, err := p.repo.ExistsByFingerprint(ctx, fingerprint, since)
	if err != nil {
		p.log.Warnw("failed to check duplicate fingerprint", "error", err, "fingerprint", fingerprint)
		return false
	}
	return exists
}

func fingerprint(l *domain.LogEntry) string {
	normalized := strings.ToLower(string(l.Level) + "|" + l.Message)
	hash := sha1.Sum([]byte(normalized))
	return hex.EncodeToString(hash[:])
}
