package alert

import (
	"testing"
	"time"

	"github.com/pingan/hydra/internal/model"
	"github.com/pingan/hydra/internal/repository"
)

func TestEvaluator_ProcessPoints(t *testing.T) {
	ruleRepo := repository.NewMemRuleRepo()
	alertRepo := repository.NewMemAlertRepo()

	rule := &model.Rule{
		Name:        "High CPU",
		Metric:      "cpu_usage_percent",
		Operator:    ">",
		Threshold:   80.0,
		DurationSec: 1,
		Severity:    "critical",
		Enabled:     true,
	}
	if err := ruleRepo.Create(rule); err != nil {
		t.Fatalf("create rule: %v", err)
	}

	notifyCh := make(chan *model.Alert, 10)
	eval := NewEvaluator(ruleRepo, alertRepo, notifyCh)

	// Process points with CPU above threshold
	points := []*model.MetricPoint{
		{TS: time.Now(), Name: "cpu_usage_percent", Value: 90.0, Host: "web-01", Region: "ap-sh-1"},
	}
	eval.ProcessPoints(points)

	// Should be pending, no alert yet
	firing, resolved := alertRepo.Stats()
	if firing != 0 {
		t.Errorf("expected 0 firing alerts after first point, got %d", firing)
	}
	_ = resolved

	// Wait for duration and process again
	time.Sleep(1100 * time.Millisecond)
	eval.ProcessPoints(points)

	firing, resolved = alertRepo.Stats()
	if firing != 1 {
		t.Errorf("expected 1 firing alert after duration, got %d", firing)
	}
	_ = resolved

	// Verify notification was sent
	select {
	case alert := <-notifyCh:
		if alert.RuleName != "High CPU" {
			t.Errorf("alert name = %q", alert.RuleName)
		}
		if alert.Status != "firing" {
			t.Errorf("alert status = %q", alert.Status)
		}
	default:
		t.Error("expected notification on alert channel")
	}
}

func TestEvaluator_NoTrigger(t *testing.T) {
	ruleRepo := repository.NewMemRuleRepo()
	alertRepo := repository.NewMemAlertRepo()

	rule := &model.Rule{
		Name:     "High CPU",
		Metric:   "cpu_usage_percent",
		Operator: ">",
		Threshold: 80.0,
		DurationSec: 5,
		Severity: "critical",
		Enabled:  true,
	}
	ruleRepo.Create(rule)

	notifyCh := make(chan *model.Alert, 10)
	eval := NewEvaluator(ruleRepo, alertRepo, notifyCh)

	// Process normal CPU (below threshold)
	points := []*model.MetricPoint{
		{TS: time.Now(), Name: "cpu_usage_percent", Value: 42.0, Host: "web-01", Region: "ap-sh-1"},
	}
	eval.ProcessPoints(points)
	time.Sleep(100 * time.Millisecond)
	eval.ProcessPoints(points)

	firing, _ := alertRepo.Stats()
	if firing != 0 {
		t.Errorf("expected 0 firing alerts, got %d", firing)
	}
}
