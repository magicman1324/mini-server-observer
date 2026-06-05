package alert

import (
	"testing"
	"time"

	"github.com/pingan/hydra/internal/model"
)

func TestLifecycle_OK_to_Firing(t *testing.T) {
	lm := NewLifecycleManager()
	rule := &model.Rule{
		ID:          "rule-1",
		Name:        "High CPU",
		Metric:      "cpu_usage_percent",
		Operator:    ">",
		Threshold:   80.0,
		DurationSec: 2,
		Severity:    "critical",
	}
	host := "web-01"
	region := "ap-sh-1"

	now := time.Now()

	// Step 1: value above threshold → pending
	state, emit, msg := lm.Evaluate(rule, host, region, 85.0, now)
	if state != StatePending {
		t.Fatalf("step 1: expected pending, got %s", state)
	}
	if emit {
		t.Fatal("step 1: should not emit yet")
	}

	// Step 2: value still above threshold but not enough time → still pending
	state, emit, _ = lm.Evaluate(rule, host, region, 90.0, now.Add(1*time.Second))
	if state != StatePending {
		t.Fatalf("step 2: expected pending, got %s", state)
	}
	if emit {
		t.Fatal("step 2: should not emit yet (duration not met)")
	}

	// Step 3: value still above threshold, duration exceeded → firing
	state, emit, msg = lm.Evaluate(rule, host, region, 95.0, now.Add(3*time.Second))
	if state != StateFiring {
		t.Fatalf("step 3: expected firing, got %s", state)
	}
	if !emit {
		t.Fatal("step 3: should emit alert")
	}
	if msg == "" {
		t.Error("step 3: message should not be empty")
	}
}

func TestLifecycle_OK_to_Recover(t *testing.T) {
	lm := NewLifecycleManager()
	rule := &model.Rule{
		ID:          "rule-1",
		Name:        "High CPU",
		Metric:      "cpu_usage_percent",
		Operator:    ">",
		Threshold:   80.0,
		DurationSec: 5,
		Severity:    "warning",
	}
	host := "web-01"
	now := time.Now()

	// Step 1: exceed → pending
	state, _, _ := lm.Evaluate(rule, host, "r", 85.0, now)
	if state != StatePending {
		t.Fatalf("step 1: expected pending, got %s", state)
	}

	// Step 2: recover before duration → ok
	state, emit, _ := lm.Evaluate(rule, host, "r", 50.0, now.Add(1*time.Second))
	if state != StateOK {
		t.Fatalf("step 2: expected ok, got %s", state)
	}
	if emit {
		t.Fatal("step 2: should not emit on recovery from pending")
	}
}

func TestLifecycle_Firing_to_Resolved(t *testing.T) {
	lm := NewLifecycleManager()
	rule := &model.Rule{
		ID:          "rule-1",
		Name:        "High CPU",
		Metric:      "cpu_usage_percent",
		Operator:    ">",
		Threshold:   80.0,
		DurationSec: 1,
		Severity:    "critical",
	}
	host := "web-01"
	now := time.Now()

	// Fire the alert
	lm.Evaluate(rule, host, "r", 85.0, now)
	state, emit, _ := lm.Evaluate(rule, host, "r", 85.0, now.Add(2*time.Second))
	if state != StateFiring || !emit {
		t.Fatal("expected firing alert")
	}

	// Now resolve
	state, emit, msg := lm.Evaluate(rule, host, "r", 50.0, now.Add(3*time.Second))
	if state != StateOK {
		t.Fatalf("expected ok after resolution, got %s", state)
	}
	if !emit {
		t.Fatal("should emit resolution message")
	}
	if msg == "" {
		t.Error("resolution message should not be empty")
	}
}

func TestCompareOp(t *testing.T) {
	tests := []struct {
		op        string
		val       float64
		threshold float64
		expected  bool
	}{
		{">", 90, 80, true},
		{">", 80, 80, false},
		{">=", 80, 80, true},
		{">=", 79, 80, false},
		{"<", 70, 80, true},
		{"<", 90, 80, false},
		{"<=", 80, 80, true},
		{"<=", 81, 80, false},
		{"==", 80, 80, true},
		{"!=", 90, 80, true},
	}

	for _, tt := range tests {
		got := compareOp(tt.val, tt.op, tt.threshold)
		if got != tt.expected {
			t.Errorf("compareOp(%g %s %g) = %v, want %v",
				tt.val, tt.op, tt.threshold, got, tt.expected)
		}
	}
}

func TestLifecycle_Reset(t *testing.T) {
	lm := NewLifecycleManager()
	rule := &model.Rule{ID: "r", Name: "test", Metric: "cpu", Operator: ">", Threshold: 80, DurationSec: 5}
	host := "h"

	lm.Evaluate(rule, host, "r", 85, time.Now())
	lm.Reset(rule.ID, host)

	// Should be fresh state
	state, _, _ := lm.Evaluate(rule, host, "r", 85, time.Now())
	if state != StatePending {
		t.Errorf("expected fresh pending after reset, got %s", state)
	}
}
