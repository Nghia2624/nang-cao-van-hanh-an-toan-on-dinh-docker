package domain

import "time"

type PredictionFeedback struct {
	ID             string    `json:"id,omitempty"`
	ContainerID    string    `json:"container_id"`
	RiskLevel      string    `json:"risk_level"`
	RiskScore      int       `json:"risk_score"`
	ETAMinutes     int       `json:"eta_minutes"`
	Accurate       bool      `json:"accurate"`
	Note           string    `json:"note,omitempty"`
	PredictionFrom time.Time `json:"prediction_from"`
	PredictionTo   time.Time `json:"prediction_to"`
	CreatedAt      time.Time `json:"created_at"`
}

// Additional fields for chat feedback
type ChatPredictionFeedback struct {
	ID           string    `json:"id,omitempty"`
	PredictionID string    `json:"prediction_id"`
	ContainerID  string    `json:"container_id"`
	IsCorrect    bool      `json:"is_correct"`
	Comment      string    `json:"comment,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

type ContainerPrediction struct {
	ID              string    `json:"id,omitempty"`
	ContainerID     string    `json:"container_id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	PredictedAt     time.Time `json:"predicted_at"`
	RiskLevel       RiskLevel `json:"risk_level"`
	RiskScore       int       `json:"risk_score"`
	ETAMinutes      int       `json:"eta_minutes"`
	RiskFactors     []string  `json:"risk_factors"`
	Recommendations []string  `json:"recommendations"`
	ConfidenceScore float64   `json:"confidence_score"`
	Signals         struct {
		CPUAvg       float64 `json:"cpu_avg"`
		CPUMax       float64 `json:"cpu_max"`
		MemAvg       float64 `json:"mem_avg"`
		MemMax       float64 `json:"mem_max"`
		ErrorCount1H int     `json:"error_count_1h"`
		RestartCount int     `json:"restart_count"`
		UptimeSec    int     `json:"uptime_sec"`
	} `json:"signals"`
	Summary string `json:"summary"`
}
