package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nghia/dockerai/backend/internal/domain"
)

type createChatSessionRequest struct {
	Title string `json:"title"`
}

type createChatSessionResponse struct {
	Session *domain.AIChatSession `json:"session"`
}

type listChatSessionsResponse struct {
	Sessions []*domain.AIChatSession `json:"sessions"`
}

type sendChatMessageRequest struct {
	Message     string `json:"message"`
	LogsText    string `json:"logs_text"`
	ContainerID string `json:"container_id"`
	From        int64  `json:"from"`
	To          int64  `json:"to"`
	MaxLogs     int    `json:"max_logs"`
}

type sendChatMessageResponse struct {
	Session  *domain.AIChatSession   `json:"session"`
	Answer   string                  `json:"answer"`
	Evidence []domain.AIChatEvidence `json:"evidence"`
	Insight  *domain.AIAnalysis      `json:"insight,omitempty"`
}

func (h *Handler) chatCreateSession(w http.ResponseWriter, r *http.Request) {
	if h.AIChatRepo == nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", "ai chat repo not configured")
		return
	}
	var req createChatSessionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "New chat"
	}
	s, err := h.AIChatRepo.CreateSession(r.Context(), title)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, createChatSessionResponse{Session: s})
}

func (h *Handler) chatListSessions(w http.ResponseWriter, r *http.Request) {
	if h.AIChatRepo == nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", "ai chat repo not configured")
		return
	}
	limit := 20
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n > 0 && n <= 100 {
				limit = n
			}
		}
	}
	sessions, err := h.AIChatRepo.ListSessions(r.Context(), limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	if sessions == nil {
		sessions = []*domain.AIChatSession{}
	}
	writeJSON(w, http.StatusOK, listChatSessionsResponse{Sessions: sessions})
}

func (h *Handler) chatGetSession(w http.ResponseWriter, r *http.Request) {
	if h.AIChatRepo == nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", "ai chat repo not configured")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "session id required")
		return
	}
	s, err := h.AIChatRepo.GetSession(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) chatSendMessage(w http.ResponseWriter, r *http.Request) {
	if h.AIChatRepo == nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", "ai chat repo not configured")
		return
	}
	var req sendChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body: "+err.Error())
		return
	}
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "session id required")
		return
	}

	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		writeErr(w, http.StatusBadRequest, "INVALID_BODY", "message is required")
		return
	}

	maxLogs := req.MaxLogs
	if maxLogs <= 0 {
		maxLogs = 120
	}
	if maxLogs > 300 {
		maxLogs = 300
	}

	// resolve logs
	logs := []*domain.LogEntry{}
	logsText := strings.TrimSpace(req.LogsText)
	if logsText != "" {
		logs = parsePastedLogs(logsText)
	}
	if len(logs) == 0 {
		if req.ContainerID != "" {
			to := time.Now().UTC()
			if req.To != 0 {
				to = time.Unix(req.To/1000, 0).UTC()
			}
			from := to.Add(-1 * time.Hour)
			if req.From != 0 {
				from = time.Unix(req.From/1000, 0).UTC()
			}
			if to.Before(from) {
				writeErr(w, http.StatusBadRequest, "INVALID_RANGE", "`to` must be after `from`")
				return
			}
			items, err := h.LogRepo.Query(r.Context(), req.ContainerID, from, to, domain.LogLevel(""), maxLogs)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
				return
			}
			logs = items
		}
	}
	if len(logs) == 0 {
		if req.ContainerID != "" {
			hints := []string{
				"Không tìm thấy log trong khoảng thời gian đã chọn.",
				"Gợi ý:",
				"- Mở rộng time range (ví dụ 6h hoặc 24h)",
				"- Tăng max_logs (tối đa 300)",
				"- Kiểm tra lại container_id có đúng không",
			}
			writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", strings.Join(hints, " "))
			return
		}
		writeErr(w, http.StatusBadRequest, "INVALID_ARGUMENT", "no logs provided (use logs_text or container_id+from/to)")
		return
	}

	userMsg := domain.AIChatMessage{
		Role:        domain.AIChatRoleUser,
		Content:     msg,
		CreatedAt:   time.Now().UTC(),
		ContainerID: strings.TrimSpace(req.ContainerID),
		FromMillis:  req.From,
		ToMillis:    req.To,
	}
	_ = h.AIChatRepo.AppendMessage(r.Context(), sessionID, userMsg)

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	answer, evidence, insight, err := h.AI.Chat(ctx, msg, logs)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "AI_ERROR", err.Error())
		return
	}

	// build assistant message with evidence
	aiEvidence := []domain.AIChatEvidence{}
	for _, e := range evidence {
		aiEvidence = append(aiEvidence, domain.AIChatEvidence{Timestamp: e.Timestamp, Level: string(e.Level), Message: e.Message, ContainerID: e.ContainerID})
	}
	assistantMsg := domain.AIChatMessage{
		Role:      domain.AIChatRoleAssistant,
		Content:   answer,
		CreatedAt: time.Now().UTC(),
		Evidence:  aiEvidence,
	}
	_ = h.AIChatRepo.AppendMessage(r.Context(), sessionID, assistantMsg)

	// update title if still default
	if s, err := h.AIChatRepo.GetSession(r.Context(), sessionID); err == nil {
		if strings.TrimSpace(s.Title) == "" || s.Title == "New chat" {
			newTitle := msg
			if len(newTitle) > 60 {
				newTitle = newTitle[:60]
			}
			_ = h.AIChatRepo.SetTitle(r.Context(), sessionID, newTitle)
			s.Title = newTitle
		}
		writeJSON(w, http.StatusOK, sendChatMessageResponse{Session: s, Answer: answer, Evidence: aiEvidence, Insight: insight})
		return
	}

	// fallback response
	s := &domain.AIChatSession{ID: sessionID}
	writeJSON(w, http.StatusOK, sendChatMessageResponse{Session: s, Answer: answer, Evidence: aiEvidence, Insight: insight})
}

