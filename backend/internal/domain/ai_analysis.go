package domain

import "time"

type AIAnalysis struct {
	ID                 string    `json:"id,omitempty"`
	Fingerprint        string    `json:"fingerprint"`
	RootCause          string    `json:"root_cause"`
	Severity           int       `json:"severity"`
	SeverityLabel      string    `json:"severity_label"`
	ImpactAnalysis     string    `json:"impact_analysis"`
	AffectedComponents []string  `json:"affected_components"`
	RecommendedActions []Action  `json:"recommended_actions"`
	PreventionMeasures []string  `json:"prevention_measures"`
	RelatedIssues      []string  `json:"related_issues"`
	ConfidenceScore    float64   `json:"confidence_score"`
	CreatedAt          time.Time `json:"created_at"`
}

type Action struct {
	Priority        int      `json:"priority"`
	Action          string   `json:"action"`
	EstimatedEffort string   `json:"estimated_effort"`
	Impact          string   `json:"impact"`
	Steps           []string `json:"steps,omitempty"`
}
