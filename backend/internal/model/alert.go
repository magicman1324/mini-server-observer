package model

import "time"

// AlertStatus represents the lifecycle state of an alert.
type AlertStatus string

const (
	AlertFiring   AlertStatus = "firing"
	AlertResolved AlertStatus = "resolved"
)

// Alert represents a triggered alert event.
type Alert struct {
	ID         string            `json:"id"`
	RuleID     string            `json:"rule_id"`
	RuleName   string            `json:"rule_name"`
	Host       string            `json:"host"`
	Region     string            `json:"region"`
	Severity   string            `json:"severity"`
	Metric     string            `json:"metric"`
	Value      float64           `json:"value"`
	Threshold  float64           `json:"threshold"`
	Message    string            `json:"message"`
	Status     string            `json:"status"`
	FiredAt    time.Time         `json:"fired_at"`
	ResolvedAt *time.Time        `json:"resolved_at,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}
