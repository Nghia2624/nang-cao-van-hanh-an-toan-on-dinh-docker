package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nghia/dockerai/backend/internal/domain"
	"github.com/nghia/dockerai/backend/internal/pkg/logger"
	"github.com/nghia/dockerai/backend/internal/repository"
)


type aiAnalyzer struct {
	endpoint       string
	apiKeys        []string
	keyIndex       int
	model          string
	cacheTTL       time.Duration
	cache          map[string]cachedAI
	mu             sync.Mutex
	repo           repository.AIRepository
	log            *logger.Logger
	client         *http.Client
	rateLimitUntil time.Time
}

type cachedAI struct {
	data      *domain.AIAnalysis
	expiresAt time.Time
}

func NewAIAnalyzer(endpoint string, keys []string, model string, repo repository.AIRepository, log *logger.Logger) AIAnalyzer {
	a := &aiAnalyzer{
		endpoint: endpoint,
		apiKeys:  append([]string{}, keys...),
		keyIndex: 0,
		model:    model,
		cacheTTL: time.Hour,
		cache:    make(map[string]cachedAI),
		repo:     repo,
		log:      log,
		client:   &http.Client{Timeout: 90 * time.Second},
	}
	go a.cleanupCache()
	return a
}

// cleanupCache periodically removes expired entries from the in-memory cache.
func (a *aiAnalyzer) cleanupCache() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		a.mu.Lock()
		now := time.Now()
		for k, v := range a.cache {
			if now.After(v.expiresAt) {
				delete(a.cache, k)
			}
		}
		a.mu.Unlock()
	}
}

func (a *aiAnalyzer) AnalyzeLogs(ctx context.Context, logs []*domain.LogEntry) (*domain.AIAnalysis, error) {
	if len(logs) == 0 {
		return nil, errors.New("empty logs")
	}
	fp := logs[0].Fingerprint
	if cached := a.getCache(fp); cached != nil {
		return cached, nil
	}
	if a.isRateLimited() {
		return nil, errors.New("AI rate limited, try later")
	}
	parsed, err := a.callGeminiWithRotation(ctx, buildPrompt(logs))
	if err != nil {
		if isRateLimitError(err) {
			a.setRateLimit()
		}
		return nil, err
	}
	analysis := &domain.AIAnalysis{
		Fingerprint:        fp,
		RootCause:          parsed.RootCause,
		Severity:           parsed.Severity,
		SeverityLabel:      parsed.SeverityLabel,
		ImpactAnalysis:     parsed.ImpactAnalysis,
		AffectedComponents: parsed.AffectedComponents,
		RecommendedActions: parsed.RecommendedActions,
		PreventionMeasures: parsed.PreventionMeasures,
		RelatedIssues:      parsed.RelatedIssues,
		ConfidenceScore:    parsed.ConfidenceScore,
		CreatedAt:          time.Now(),
	}
	_ = a.repo.Save(context.Background(), analysis)
	a.setCache(fp, analysis)
	return analysis, nil
}

// callGeminiWithRotation tries each API key once until one succeeds, then
// remembers the working key index for subsequent calls.
func (a *aiAnalyzer) callGeminiWithRotation(ctx context.Context, prompt string) (*struct {
	RootCause          string          `json:"root_cause"`
	Severity           int             `json:"severity"`
	SeverityLabel      string          `json:"severity_label"`
	ImpactAnalysis     string          `json:"impact_analysis"`
	AffectedComponents []string        `json:"affected_components"`
	RecommendedActions []domain.Action `json:"recommended_actions"`
	PreventionMeasures []string        `json:"prevention_measures"`
	RelatedIssues      []string        `json:"related_issues"`
	ConfidenceScore    float64         `json:"confidence_score"`
	Answer             string          `json:"answer"`
}, error) {
	if len(a.apiKeys) == 0 {
		return nil, errors.New("no AI API keys configured")
	}

	type geminiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	var lastErr error
	for i := 0; i < len(a.apiKeys); i++ {
		key := a.currentKey()
		payload := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]string{
						{"text": prompt},
					},
				},
			},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint+"?key="+key, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			lastErr = err
			a.advanceKey()
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
			// Quota/rate limited on this key → rotate to next
			resp.Body.Close()
			lastErr = fmt.Errorf("ai quota/rate-limit on current key")
			a.advanceKey()
			continue
		}
		if resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("ai request failed: status %d", resp.StatusCode)
			resp.Body.Close()
			a.advanceKey()
			continue
		}

		var gr geminiResponse
		if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
			resp.Body.Close()
			lastErr = err
			a.advanceKey()
			continue
		}
		resp.Body.Close()

		if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
			lastErr = errors.New("ai response missing candidates")
			a.advanceKey()
			continue
		}
		text := gr.Candidates[0].Content.Parts[0].Text
		// Extract JSON from text (AI might return markdown code blocks)
		jsonText := text
		if idx := strings.Index(text, "```json"); idx >= 0 {
			jsonText = text[idx+7:]
			if idx2 := strings.Index(jsonText, "```"); idx2 >= 0 {
				jsonText = jsonText[:idx2]
			}
		} else if idx := strings.Index(text, "```"); idx >= 0 {
			jsonText = text[idx+3:]
			if idx2 := strings.Index(jsonText, "```"); idx2 >= 0 {
				jsonText = jsonText[:idx2]
			}
		}
		jsonText = strings.TrimSpace(jsonText)

		var parsed struct {
			RootCause          string          `json:"root_cause"`
			Severity           int             `json:"severity"`
			SeverityLabel      string          `json:"severity_label"`
			ImpactAnalysis     string          `json:"impact_analysis"`
			AffectedComponents []string        `json:"affected_components"`
			RecommendedActions []domain.Action `json:"recommended_actions"`
			PreventionMeasures []string        `json:"prevention_measures"`
			RelatedIssues      []string        `json:"related_issues"`
			ConfidenceScore    float64         `json:"confidence_score"`
			Answer             string          `json:"answer"`
		}
		// Ensure steps field is initialized for each action
		if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
			a.log.Warnw("failed to parse AI response", "error", err, "text", text[:min(500, len(text))], "jsonText", jsonText[:min(500, len(jsonText))])
			lastErr = fmt.Errorf("failed to parse AI JSON response: %w", err)
			a.advanceKey()
			continue
		}
		// Initialize nil steps slices after successful unmarshal
		for i := range parsed.RecommendedActions {
			if parsed.RecommendedActions[i].Steps == nil {
				parsed.RecommendedActions[i].Steps = []string{}
			}
		}
		// Validate parsed data
		if parsed.RootCause == "" {
			a.log.Warnw("AI response missing root_cause", "text", text[:min(500, len(text))])
			lastErr = errors.New("ai response missing required field: root_cause")
			a.advanceKey()
			continue
		}
		// Validate severity
		if parsed.Severity < 1 || parsed.Severity > 5 {
			a.log.Warnw("AI response has invalid severity", "severity", parsed.Severity)
			parsed.Severity = 3 // Default to MEDIUM
		}
		// Validate severity label
		if parsed.SeverityLabel == "" {
			severityMap := map[int]string{1: "INFO", 2: "LOW", 3: "MEDIUM", 4: "HIGH", 5: "CRITICAL"}
			parsed.SeverityLabel = severityMap[parsed.Severity]
			if parsed.SeverityLabel == "" {
				parsed.SeverityLabel = "MEDIUM"
			}
		}
		// Ensure arrays are not nil
		if parsed.AffectedComponents == nil {
			parsed.AffectedComponents = []string{}
		}
		if parsed.RecommendedActions == nil {
			parsed.RecommendedActions = []domain.Action{}
		}
		if parsed.PreventionMeasures == nil {
			parsed.PreventionMeasures = []string{}
		}
		if parsed.RelatedIssues == nil {
			parsed.RelatedIssues = []string{}
		}
		// success: keep current key index
		return &parsed, nil
	}
	if lastErr == nil {
		lastErr = errors.New("ai request failed for all keys")
	}
	return nil, lastErr
}

func (a *aiAnalyzer) currentKey() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.apiKeys[a.keyIndex%len(a.apiKeys)]
}

func (a *aiAnalyzer) advanceKey() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.keyIndex = (a.keyIndex + 1) % len(a.apiKeys)
}

func (a *aiAnalyzer) AnalyzeError(ctx context.Context, log *domain.LogEntry, _ *domain.SystemMetrics) (*domain.AIAnalysis, error) {
	return a.AnalyzeLogs(ctx, []*domain.LogEntry{log})
}

func (a *aiAnalyzer) ProcessQueue(ctx context.Context) error {
	// In this simplified implementation, queue is owned by LogProcessor; nothing to process here.
	<-ctx.Done()
	return ctx.Err()
}

func (a *aiAnalyzer) Chat(ctx context.Context, message string, logs []*domain.LogEntry) (answer string, evidence []*domain.LogEntry, insight *domain.AIAnalysis, err error) {
	if len(logs) == 0 {
		return "No logs provided to analyze", nil, nil, nil
	}

	// Build chat prompt
	prompt := a.buildChatPrompt(message, logs)

	// Call AI with rotation
	parsed, err := a.callGeminiWithRotation(ctx, prompt)
	if err != nil {
		return "", nil, nil, err
	}

	// Extract answer from AI response
	answer = parsed.Answer

	// Return logs as evidence
	evidence = logs

	// Create insight from analysis
	insight = &domain.AIAnalysis{
		Fingerprint:        logs[0].Fingerprint,
		RootCause:          parsed.RootCause,
		Severity:           parsed.Severity,
		SeverityLabel:      parsed.SeverityLabel,
		ImpactAnalysis:     parsed.ImpactAnalysis,
		AffectedComponents: parsed.AffectedComponents,
		RecommendedActions: parsed.RecommendedActions,
		PreventionMeasures: parsed.PreventionMeasures,
		RelatedIssues:      parsed.RelatedIssues,
		ConfidenceScore:    parsed.ConfidenceScore,
		CreatedAt:          time.Now(),
	}

	return answer, evidence, insight, nil
}

func (a *aiAnalyzer) buildChatPrompt(message string, logs []*domain.LogEntry) string {
	var b strings.Builder
	b.WriteString("Bạn là một chuyên gia DevOps và SRE. Người dùng đặt câu hỏi về logs sau:\n\n")
	b.WriteString("Câu hỏi: ")
	b.WriteString(message)
	b.WriteString("\n\nLogs:\n")
	b.WriteString("---\n")
	for i, l := range logs {
		b.WriteString(fmt.Sprintf("[%d] %s [%s] %s\n", i+1, l.Timestamp.Format(time.RFC3339), string(l.Level), l.Message))
		if l.ContainerID != "" {
			containerID := l.ContainerID
			if len(containerID) > 12 {
				containerID = containerID[:12]
			}
			b.WriteString(fmt.Sprintf("    Container: %s\n", containerID))
		}
	}
	b.WriteString("---\n\n")
	b.WriteString("Hãy trả lời câu hỏi của người dùng bằng tiếng Việt, dựa trên phân tích logs.\n")
	b.WriteString("Trả về CHỈ JSON (không markdown) với format:\n")
	b.WriteString("{\n")
	b.WriteString("  \"answer\": \"Câu trả lời chi tiết cho người dùng (tiếng Việt)\",\n")
	b.WriteString("  \"root_cause\": \"Nguyên nhân gốc rễ nếu có\",\n")
	b.WriteString("  \"severity\": 1-5,\n")
	b.WriteString("  \"severity_label\": \"CRITICAL|HIGH|MEDIUM|LOW|INFO\",\n")
	b.WriteString("  \"impact_analysis\": \"Phân tích tác động\",\n")
	b.WriteString("  \"affected_components\": [\"component1\", \"component2\"],\n")
	b.WriteString("  \"recommended_actions\": [\n")
	b.WriteString("    {\"action\": \"Hành động cụ thể\", \"priority\": 1-5, \"estimated_effort\": \"5 minutes\", \"impact\": \"High\", \"steps\": [\"bước 1\", \"bước 2\"]}\n")
	b.WriteString("  ],\n")
	b.WriteString("  \"prevention_measures\": [\"Biện pháp phòng ngừa 1\", \"Biện pháp phòng ngừa 2\"],\n")
	b.WriteString("  \"related_issues\": [\"Vấn đề liên quan 1\", \"Vấn đề liên quan 2\"],\n")
	b.WriteString("  \"confidence_score\": 0.85\n")
	b.WriteString("}\n")
	return b.String()
}

func (a *aiAnalyzer) getCache(fp string) *domain.AIAnalysis {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c, ok := a.cache[fp]; ok {
		if time.Now().Before(c.expiresAt) {
			return c.data
		}
		delete(a.cache, fp)
	}
	return nil
}

func (a *aiAnalyzer) isRateLimited() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return time.Now().Before(a.rateLimitUntil)
}

func (a *aiAnalyzer) setRateLimit() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rateLimitUntil = time.Now().Add(5 * time.Minute)
}

func isRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "quota/rate-limit") || strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "403")
}

func (a *aiAnalyzer) setCache(fp string, data *domain.AIAnalysis) {
	a.mu.Lock()
	defer a.mu.Unlock()

	const maxCacheSize = 500

	// Evict if cache is at max size
	if len(a.cache) >= maxCacheSize {
		now := time.Now()
		// First pass: remove expired entries
		for k, v := range a.cache {
			if now.After(v.expiresAt) {
				delete(a.cache, k)
			}
		}
		// Second pass: if still over limit, remove oldest entries
		for len(a.cache) >= maxCacheSize {
			var oldestKey string
			var oldestTime time.Time
			first := true
			for k, v := range a.cache {
				if first || v.expiresAt.Before(oldestTime) {
					oldestKey = k
					oldestTime = v.expiresAt
					first = false
				}
			}
			if oldestKey != "" {
				delete(a.cache, oldestKey)
			} else {
				break
			}
		}
	}

	a.cache[fp] = cachedAI{data: data, expiresAt: time.Now().Add(a.cacheTTL)}
}

func buildPrompt(logs []*domain.LogEntry) string {
	var b strings.Builder
	b.WriteString("Bạn là một chuyên gia DevOps và SRE với nhiều năm kinh nghiệm phân tích logs và troubleshooting hệ thống Docker/Kubernetes.\n\n")
	b.WriteString("Nhiệm vụ: Phân tích các log entries sau đây và cung cấp phân tích chi tiết theo format JSON.\n\n")
	b.WriteString("Logs cần phân tích:\n")
	b.WriteString("---\n")
	for i, l := range logs {
		b.WriteString(fmt.Sprintf("[%d] %s [%s] %s\n", i+1, l.Timestamp.Format(time.RFC3339), string(l.Level), l.Message))
		if l.ContainerID != "" {
			containerID := l.ContainerID
			if len(containerID) > 12 {
				containerID = containerID[:12]
			}
			b.WriteString(fmt.Sprintf("    Container: %s\n", containerID))
		}
	}
	b.WriteString("---\n\n")
	b.WriteString("Hãy phân tích và trả về CHỈ JSON (không có markdown, không có text giải thích) với các trường sau:\n")
	b.WriteString("{\n")
	b.WriteString("  \"root_cause\": \"Nguyên nhân gốc rễ của vấn đề (tiếng Việt, chi tiết, dựa trên patterns trong logs)\",\n")
	b.WriteString("  \"severity\": 1-5 (1=INFO, 2=LOW, 3=MEDIUM, 4=HIGH, 5=CRITICAL),\n")
	b.WriteString("  \"severity_label\": \"CRITICAL|HIGH|MEDIUM|LOW|INFO\",\n")
	b.WriteString("  \"impact_analysis\": \"Phân tích tác động đến hệ thống, services, users (tiếng Việt, chi tiết)\",\n")
	b.WriteString("  \"affected_components\": [\"component1\", \"component2\"],\n")
	b.WriteString("  \"recommended_actions\": [\n")
	b.WriteString("    {\"action\": \"Hành động cụ thể\", \"priority\": 1-5, \"estimated_effort\": \"5 minutes\", \"impact\": \"High\", \"steps\": [\"bước 1\", \"bước 2\"]}\n")
	b.WriteString("  ],\n")
	b.WriteString("  LƯU Ý: Trường \"steps\" là bắt buộc trong recommended_actions, phải là mảng các bước cụ thể để thực hiện hành động.\n")
	b.WriteString("  \"prevention_measures\": [\"Biện pháp phòng ngừa 1\", \"Biện pháp phòng ngừa 2\"],\n")
	b.WriteString("  \"related_issues\": [\"Vấn đề liên quan 1\", \"Vấn đề liên quan 2\"],\n")
	b.WriteString("  \"confidence_score\": 0.85\n")
	b.WriteString("}\n\n")
	b.WriteString("QUAN TRỌNG:\n")
	b.WriteString("- Chỉ trả về JSON thuần túy, không có markdown code blocks\n")
	b.WriteString("- Phân tích dựa trên patterns, error messages, timestamps\n")
	b.WriteString("- Đưa ra recommendations cụ thể, actionable, có thể thực hiện ngay\n")
	b.WriteString("- Ưu tiên các actions có thể giải quyết vấn đề nhanh nhất\n")
	b.WriteString("- Prevention measures phải thực tế và khả thi\n")
	b.WriteString("- Confidence score dựa trên độ rõ ràng của error messages và patterns\n")
	return b.String()
}
