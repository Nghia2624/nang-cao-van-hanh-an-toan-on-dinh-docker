package httpapi

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nghia/dockerai/backend/internal/domain"
	"github.com/nghia/dockerai/backend/internal/pkg/logger"
	"github.com/nghia/dockerai/backend/internal/repository"
	"github.com/nghia/dockerai/backend/internal/service"
)

type Handler struct {
	Docker                 service.DockerService
	Metrics                service.MetricsService
	Logs                   service.LogProcessor
	AI                     service.AIAnalyzer
	Alerts                 service.AlertEngine
	SSE                    *SSEHub
	Log                    *logger.Logger
	AlertRepo              repository.AlertRepository
	LogRepo                repository.LogRepository
	AIRepo                 repository.AIRepository
	AIChatRepo             repository.AIChatRepository
	PredictionFeedbackRepo repository.PredictionFeedbackRepository
}

func (h *Handler) listContainers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.Log.Infow("listContainers called")
	res, err := h.Docker.ListContainers(ctx)
	if err != nil {
		h.Log.Errorw("failed to list containers", "error", err)
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	h.Log.Infow("listContainers success", "count", len(res))
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) systemMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := h.Metrics.GetSystemOverview(ctx)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *Handler) getLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	containerID := q.Get("container")
	level := domain.LogLevel(q.Get("level"))
	limit := 200
	if s := q.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 2000 {
			limit = n
		}
	}

	now := time.Now().UTC()
	from := now.Add(-15 * time.Minute)
	to := now
	if s := q.Get("from"); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			from = time.Unix(ts/1000, 0).UTC()
		} else if t, err := time.Parse(time.RFC3339, s); err == nil {
			from = t
		}
	}
	if s := q.Get("to"); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			to = time.Unix(ts/1000, 0).UTC()
		} else if t, err := time.Parse(time.RFC3339, s); err == nil {
			to = t
		}
	}
	if to.Before(from) {
		writeErr(w, http.StatusBadRequest, "INVALID_RANGE", "`to` must be after `from`")
		return
	}
	items, err := h.LogRepo.Query(r.Context(), containerID, from, to, level, limit)
	if err != nil {
		h.Log.Errorw("failed to query logs", "error", err)
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	// Ensure we always return an array, never nil
	if items == nil {
		items = []*domain.LogEntry{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) streamLogsSSE(w http.ResponseWriter, r *http.Request) {
	containerID := r.URL.Query().Get("container")
	if containerID == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "container required")
		return
	}
	ctx := r.Context()

	// Look up container name for enrichment
	containerName := ""
	if c, err := h.Docker.GetContainer(ctx, containerID); err == nil {
		containerName = c.Name
	}

	ch, err := h.Docker.StreamLogs(ctx, containerID, true, 100)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	for entry := range ch {
		if containerName != "" {
			entry.Container = containerName
		}
		_ = h.Logs.ProcessLog(ctx, &entry)
		payload, _ := json.Marshal(entry)
		w.Write([]byte("data: " + string(payload) + "\n\n"))
		flusher.Flush()
	}
}

func (h *Handler) listAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	statusFilter := domain.AlertStatus(r.URL.Query().Get("status"))
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	items, err := h.AlertRepo.List(ctx, statusFilter, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) aiAnalyze(w http.ResponseWriter, r *http.Request) {
	var logEntry domain.LogEntry
	if err := json.NewDecoder(r.Body).Decode(&logEntry); err != nil {
		h.Log.Warnw("failed to decode AI analyze request", "error", err)
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body: "+err.Error())
		return
	}
	// Validate required fields
	if logEntry.Message == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "message field is required")
		return
	}
	if logEntry.Level == "" {
		logEntry.Level = domain.LogLevelError // Default to ERROR if not specified
	}
	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now().UTC()
	}
	// Generate fingerprint if not provided
	if logEntry.Fingerprint == "" {
		logEntry.Fingerprint = fingerprintLogEntry(&logEntry)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	h.Log.Infow("analyzing log with AI", "fingerprint", logEntry.Fingerprint, "level", logEntry.Level)
	analysis, err := h.AI.AnalyzeError(ctx, &logEntry, nil)
	if err != nil {
		h.Log.Errorw("AI analysis failed", "error", err, "fingerprint", logEntry.Fingerprint)
		writeErr(w, http.StatusInternalServerError, "AI_ERROR", "failed to analyze log: "+err.Error())
		return
	}
	if analysis == nil {
		h.Log.Warnw("AI analysis returned nil", "fingerprint", logEntry.Fingerprint)
		writeErr(w, http.StatusInternalServerError, "AI_ERROR", "analysis returned empty result")
		return
	}
	h.Log.Infow("AI analysis completed", "fingerprint", analysis.Fingerprint, "severity", analysis.SeverityLabel)
	writeJSON(w, http.StatusOK, analysis)
}

func fingerprintLogEntry(log *domain.LogEntry) string {
	normalized := strings.ToLower(string(log.Level) + "|" + log.Message)
	hash := sha1.Sum([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

func (h *Handler) getContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "container id required")
		return
	}
	container, err := h.Docker.GetContainer(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, container)
}

func (h *Handler) getContainerCPU(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "container id required")
		return
	}
	q := r.URL.Query()
	from := time.Now().Add(-1 * time.Hour)
	to := time.Now()
	if s := q.Get("from"); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			from = time.Unix(ts/1000, 0)
		} else if t, err := time.Parse(time.RFC3339, s); err == nil {
			from = t
		}
	}
	if s := q.Get("to"); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			to = time.Unix(ts/1000, 0)
		} else if t, err := time.Parse(time.RFC3339, s); err == nil {
			to = t
		}
	}
	points, err := h.Metrics.GetContainerCPU(r.Context(), id, from, to)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, points)
}

func (h *Handler) getContainerMemory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "container id required")
		return
	}
	q := r.URL.Query()
	from := time.Now().Add(-1 * time.Hour)
	to := time.Now()
	if s := q.Get("from"); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			from = time.Unix(ts/1000, 0)
		} else if t, err := time.Parse(time.RFC3339, s); err == nil {
			from = t
		}
	}
	if s := q.Get("to"); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			to = time.Unix(ts/1000, 0)
		} else if t, err := time.Parse(time.RFC3339, s); err == nil {
			to = t
		}
	}
	points, err := h.Metrics.GetContainerMemory(r.Context(), id, from, to)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, points)
}

func (h *Handler) updateAlertStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	action := chi.URLParam(r, "action")
	if id == "" || action == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "id and action required")
		return
	}
	var status domain.AlertStatus
	switch action {
	case "acknowledged":
		status = domain.AlertStatusAcknowledged
	case "resolved":
		status = domain.AlertStatusResolved
	default:
		writeErr(w, http.StatusBadRequest, "INVALID_ACTION", "action must be 'acknowledged' or 'resolved'")
		return
	}
	if err := h.AlertRepo.UpdateStatus(r.Context(), id, status); err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listAIAnalyses(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	analyses, err := h.AIRepo.List(r.Context(), limit)
	if err != nil {
		h.Log.Errorw("failed to list AI analyses", "error", err)
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	// Ensure we always return an array, never nil
	if analyses == nil {
		analyses = []*domain.AIAnalysis{}
	}
	writeJSON(w, http.StatusOK, analyses)
}

// parsePastedLogs parses raw pasted log text into structured LogEntry objects.
// Used by chat_sessions.go for AI chat with pasted logs.
func parsePastedLogs(text string) []*domain.LogEntry {
	lines := strings.Split(text, "\n")
	var logs []*domain.LogEntry
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ts := time.Now().UTC()
		level := domain.LogLevelInfo
		message := line

		// Try pattern: timestamp [LEVEL] message
		if strings.Contains(line, " [") && strings.Contains(line, "] ") {
			parts := strings.SplitN(line, "] ", 2)
			if len(parts) == 2 {
				header := parts[0]
				message = parts[1]
				if idx := strings.LastIndex(header, " ["); idx >= 0 {
					level = domain.LogLevel(strings.ToUpper(header[idx+2:]))
					tsStr := header[:idx]
					if t, err := time.Parse(time.RFC3339, tsStr); err == nil {
						ts = t
					} else if t, err := time.Parse("2006-01-02 15:04:05", tsStr); err == nil {
						ts = t
					}
				}
			}
		}

		entry := &domain.LogEntry{Timestamp: ts, Level: level, Message: message}
		entry.Fingerprint = fingerprintLogEntry(entry)
		logs = append(logs, entry)
	}
	return logs
}

func (h *Handler) getContainerPrediction(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "id")
	if containerID == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "container id required")
		return
	}

	// Get container info
	container, err := h.Docker.GetContainer(r.Context(), containerID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	// Get recent metrics for prediction (last 24h)
	metricsTo := time.Now()
	metricsFrom := metricsTo.Add(-24 * time.Hour)
	cpuPoints, err := h.Metrics.GetContainerCPU(r.Context(), containerID, metricsFrom, metricsTo)
	if err != nil {
		h.Log.Errorw("failed to get CPU metrics for prediction", "error", err, "container", containerID)
		cpuPoints = []domain.MetricPoint{}
	}

	memoryPoints, err := h.Metrics.GetContainerMemory(r.Context(), containerID, metricsFrom, metricsTo)
	if err != nil {
		h.Log.Errorw("failed to get memory metrics for prediction", "error", err, "container", containerID)
		memoryPoints = []domain.MetricPoint{}
	}

	// Get recent logs for context
	to := time.Now().UTC()
	from := to.Add(-1 * time.Hour)
	logs, err := h.LogRepo.Query(r.Context(), containerID, from, to, domain.LogLevelError, 50)
	if err != nil {
		h.Log.Errorw("failed to get logs for prediction", "error", err, "container", containerID)
		logs = []*domain.LogEntry{}
	}

	// Prediction logic based on metrics and logs
	prediction := h.generatePrediction(container, cpuPoints, memoryPoints, logs)
	// enrich with docker restart/uptime
	prediction.Signals.RestartCount = container.RestartCount
	prediction.Signals.UptimeSec = int(container.UptimeSec)
	writeJSON(w, http.StatusOK, prediction)
}

func (h *Handler) submitPredictionFeedback(w http.ResponseWriter, r *http.Request) {
	containerID := chi.URLParam(r, "id")
	if containerID == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "container id required")
		return
	}

	var feedback struct {
		PredictionID string `json:"prediction_id"`
		IsCorrect    bool   `json:"is_correct"`
		Comment      string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body: "+err.Error())
		return
	}

	fb := &domain.ChatPredictionFeedback{
		PredictionID: feedback.PredictionID,
		ContainerID:  containerID,
		IsCorrect:    feedback.IsCorrect,
		Comment:      feedback.Comment,
		CreatedAt:    time.Now().UTC(),
	}

	if err := h.PredictionFeedbackRepo.SaveChatFeedback(r.Context(), fb); err != nil {
		h.Log.Errorw("failed to save prediction feedback", "error", err)
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) generatePrediction(container *domain.Container, cpuPoints, memoryPoints []domain.MetricPoint, logs []*domain.LogEntry) *domain.ContainerPrediction {
	prediction := &domain.ContainerPrediction{
		ContainerID:     container.ID,
		Name:            container.Name,
		Status:          container.Status,
		PredictedAt:     time.Now().UTC(),
		RiskFactors:     []string{},
		Recommendations: []string{},
		ETAMinutes:      -1, // Default to N/A
	}

	// --- Signal Analysis ---
	var avgCPU, maxCPU, avgMem, maxMem float64
	if len(cpuPoints) > 10 {
		for _, p := range cpuPoints {
			avgCPU += p.Value
			if p.Value > maxCPU {
				maxCPU = p.Value
			}
		}
		avgCPU /= float64(len(cpuPoints))
	}
	if len(memoryPoints) > 10 {
		for _, p := range memoryPoints {
			avgMem += p.Value
			if p.Value > maxMem {
				maxMem = p.Value
			}
		}
		avgMem /= float64(len(memoryPoints))
	}
	errorCount1H := len(logs)

	prediction.Signals.CPUAvg = avgCPU
	prediction.Signals.CPUMax = maxCPU
	prediction.Signals.MemAvg = avgMem
	prediction.Signals.MemMax = maxMem
	prediction.Signals.ErrorCount1H = errorCount1H

	// --- Risk Score Calculation (0-100) ---
	// Normalize: metrics come as percentage (0-100), convert to ratio (0-1)
	cpuNorm := avgCPU / 100.0
	cpuMaxNorm := maxCPU / 100.0
	memNorm := avgMem / 100.0
	memMaxNorm := maxMem / 100.0

	var riskScore int
	riskScore += int(cpuNorm * 25)           // High avg CPU contributes up to 25 points
	riskScore += int(cpuMaxNorm * 15)         // CPU spikes contribute up to 15 points
	riskScore += int(memNorm * 25)            // High avg Memory contributes up to 25 points
	riskScore += int(memMaxNorm * 15)         // Memory spikes contribute up to 15 points
	riskScore += errorCount1H * 2             // Each error adds 2 points, capped below
	riskScore += container.RestartCount * 5   // Each restart adds 5 points

	if riskScore > 100 {
		riskScore = 100
	}
	prediction.RiskScore = riskScore

	// --- Determine Risk Level & ETA ---
	if riskScore > 75 {
		prediction.RiskLevel = domain.RiskLevelHigh
		prediction.ETAMinutes = 30 // High risk, potential failure within 30 mins
		prediction.Summary = "Container is at high risk of failure due to sustained high resource usage or frequent errors."
	} else if riskScore > 40 {
		prediction.RiskLevel = domain.RiskLevelMedium
		prediction.ETAMinutes = 240 // Medium risk, potential issues within 4 hours
		prediction.Summary = "Container is showing signs of stress. Performance may be degraded."
	} else {
		prediction.RiskLevel = domain.RiskLevelLow
		prediction.Summary = "Container appears to be running normally."
	}

	// --- Generate Factors & Recommendations ---
	// Use percentage values (0-100) for thresholds
	if avgCPU > 80 {
		prediction.RiskFactors = append(prediction.RiskFactors, "High Average CPU Usage (>80%)")
		prediction.Recommendations = append(prediction.Recommendations, "Investigate process-level CPU usage or consider scaling up.")
	}
	if maxCPU > 95 {
		prediction.RiskFactors = append(prediction.RiskFactors, "CPU Spikes Detected (>95%)")
	}
	if avgMem > 80 {
		prediction.RiskFactors = append(prediction.RiskFactors, "High Average Memory Usage (>80%)")
		prediction.Recommendations = append(prediction.Recommendations, "Check for memory leaks or consider increasing memory limits.")
	}
	if maxMem > 95 {
		prediction.RiskFactors = append(prediction.RiskFactors, "Memory Spikes Detected (>95%)")
	}
	if errorCount1H > 10 {
		prediction.RiskFactors = append(prediction.RiskFactors, "High Error Rate in Logs (>10/hr)")
		prediction.Recommendations = append(prediction.Recommendations, "Analyze error logs to identify the root cause.")
	}
	if container.RestartCount > 3 {
		prediction.RiskFactors = append(prediction.RiskFactors, "Frequent Restarts")
	}
	if len(prediction.RiskFactors) == 0 {
		prediction.RiskFactors = append(prediction.RiskFactors, "No significant risk factors detected.")
	}
	if len(prediction.Recommendations) == 0 {
		prediction.Recommendations = append(prediction.Recommendations, "Continue monitoring.")
	}

	// --- Confidence Score (clamped to 0.0 - 1.0) ---
	confidence := 0.7
	if len(cpuPoints) < 10 || len(memoryPoints) < 10 {
		confidence -= 0.2
	}
	if errorCount1H == 0 {
		confidence += 0.1
	}
	if container.UptimeSec < 3600 {
		confidence -= 0.1 // Less confident for newly started containers
	}
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	prediction.ConfidenceScore = confidence

	return prediction
}
