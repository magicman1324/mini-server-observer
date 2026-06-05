package alert

import "time"

// State represents the current state of an alert for a given rule+host combination.
type State string

const (
	StateOK      State = "ok"
	StatePending State = "pending"
	StateFiring  State = "firing"
)

// AlertState tracks the lifecycle state for a specific rule on a specific host.
type AlertState struct {
	RuleID     string
	Host       string
	State      State
	StartedAt  time.Time // when the threshold was first exceeded in this cycle
	LastValue  float64
	LastTS     time.Time
}
