package api

import "time"

type ContainerDTO struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	Status       string            `json:"status"`
	UptimeSec    int64             `json:"uptime_sec"`
	RestartCount int               `json:"restart_count"`
	Labels       map[string]string `json:"labels,omitempty"`
}

type MetricPointDTO struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type SystemMetricsDTO struct {
	CPUPercent       float64 `json:"cpu_percent"`
	MemoryPercent    float64 `json:"memory_percent"`
	ContainerRunning int     `json:"container_running"`
	ContainerTotal   int     `json:"container_total"`
}

type AIAnalyzeRequest struct {
	LogMessage string `json:"log_message"`
	Level      string `json:"level"`
	Container  string `json:"container"`
}

type AIAnalyzeResponse struct {
	RootCause          string   `json:"root_cause"`
	Severity           int      `json:"severity"`
	SeverityLabel      string   `json:"severity_label"`
	ImpactAnalysis     string   `json:"impact_analysis"`
	AffectedComponents []string `json:"affected_components"`
	RecommendedActions []string `json:"recommended_actions"`
	PreventionMeasures []string `json:"prevention_measures"`
	RelatedIssues      []string `json:"related_issues"`
	ConfidenceScore    float64  `json:"confidence_score"`
}
