package alert

import (
	"time"

	"github.com/pingan/hydra/internal/model"
)

// LifecycleManager handles the alert lifecycle: pending → firing → resolved.
type LifecycleManager struct {
	states   map[string]*AlertState // key = ruleID:host
	evalFn   func(state *AlertState, rule *model.Rule, value float64, now time.Time) State
}

// NewLifecycleManager creates a lifecycle manager.
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		states: make(map[string]*AlertState),
	}
}

// Evaluate checks the current value against a rule and returns the resulting state.
// Returns: (newState, shouldEmitAlert, alertMessage)
func (m *LifecycleManager) Evaluate(rule *model.Rule, host, region string, value float64, now time.Time) (State, bool, string) {
	key := rule.ID + ":" + host
	state, exists := m.states[key]
	if !exists {
		state = &AlertState{RuleID: rule.ID, Host: host, State: StateOK}
		m.states[key] = state
	}

	exceeded := compareOp(value, rule.Operator, rule.Threshold)

	switch state.State {
	case StateOK:
		if exceeded {
			state.State = StatePending
			state.StartedAt = now
			state.LastValue = value
			state.LastTS = now
			return StatePending, false, ""
		}

	case StatePending:
		if !exceeded {
			state.State = StateOK
			return StateOK, false, ""
		}

		if now.Sub(state.StartedAt) >= time.Duration(rule.DurationSec)*time.Second {
			state.State = StateFiring
			state.LastValue = value
			state.LastTS = now
			msg := FormatAlertMessage(rule, host, region, value)
			return StateFiring, true, msg
		}
		state.LastValue = value
		return StatePending, false, ""

	case StateFiring:
		if !exceeded {
			state.State = StateOK
			return StateOK, true, FormatResolvedMessage(rule, host, region, value)
		}
		state.LastValue = value
		return StateFiring, false, ""
	}

	return StateOK, false, ""
}

// Reset clears state for a given rule:host key.
func (m *LifecycleManager) Reset(ruleID, host string) {
	delete(m.states, ruleID+":"+host)
}

// compareOp evaluates a binary comparison between a value and threshold.
func compareOp(value float64, op string, threshold float64) bool {
	switch op {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}
