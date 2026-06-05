package model

import "time"

// RuleType classifies alert rules.
type RuleType string

const (
	RuleAtomic    RuleType = "atomic"
	RuleComposite RuleType = "composite"
)

// Severity levels for alert rules.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// Rule defines an alert evaluation rule.
type Rule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Metric      string            `json:"metric"`
	Operator    string            `json:"operator"`
	Threshold   float64           `json:"threshold"`
	DurationSec uint32            `json:"duration_sec"`
	Severity    string            `json:"severity"`
	Expression  string            `json:"expression,omitempty"`
	Enabled     bool              `json:"enabled"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ValidOperators is the set of allowed comparison operators.
var ValidOperators = map[string]bool{
	">": true, ">=": true, "<": true, "<=": true, "==": true, "!=": true,
}

// Validate returns an error message if the rule has invalid fields, or empty string if valid.
func (r *Rule) Validate() string {
	if r.Name == "" {
		return "name is required"
	}
	if r.Metric == "" {
		return "metric is required"
	}
	if !ValidOperators[r.Operator] {
		return "invalid operator: " + r.Operator
	}
	if r.DurationSec == 0 {
		r.DurationSec = 30
	}
	if r.Severity == "" {
		r.Severity = "warning"
	}
	return ""
}
