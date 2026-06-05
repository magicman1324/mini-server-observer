package repository

import (
	"testing"

	"github.com/pingan/hydra/internal/model"
)

func TestMemRuleRepo_CRUD(t *testing.T) {
	repo := NewMemRuleRepo()

	// Create
	rule := &model.Rule{
		Name:     "High CPU",
		Metric:   "cpu_usage_percent",
		Operator: ">",
		Threshold: 80.0,
		DurationSec: 30,
		Severity: "warning",
		Enabled:  true,
	}
	if err := repo.Create(rule); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if rule.ID == "" {
		t.Fatal("expected non-empty ID after create")
	}

	// GetByID
	got, err := repo.GetByID(rule.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Name != "High CPU" {
		t.Errorf("name = %q, want 'High CPU'", got.Name)
	}

	// Update
	rule.Name = "High CPU Updated"
	if err := repo.Update(rule); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	got, _ = repo.GetByID(rule.ID)
	if got.Name != "High CPU Updated" {
		t.Errorf("name = %q after update", got.Name)
	}

	// List
	rules, total, err := repo.List(1, 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(rules) != 1 {
		t.Errorf("len = %d, want 1", len(rules))
	}

	// ListEnabled
	enabled, _ := repo.ListEnabled()
	if len(enabled) != 1 {
		t.Errorf("enabled = %d, want 1", len(enabled))
	}

	// Delete
	if err := repo.Delete(rule.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, err = repo.GetByID(rule.ID)
	if err == nil {
		t.Error("expected error getting deleted rule")
	}
}

func TestMemRuleRepo_Validation(t *testing.T) {
	tests := []struct {
		name    string
		rule    model.Rule
		wantErr string
	}{
		{"empty name", model.Rule{Name: "", Metric: "cpu", Operator: ">"}, "name is required"},
		{"empty metric", model.Rule{Name: "test", Metric: "", Operator: ">"}, "metric is required"},
		{"invalid operator", model.Rule{Name: "test", Metric: "cpu", Operator: "??"}, "invalid operator: ??"},
		{"default severity", model.Rule{Name: "test", Metric: "cpu", Operator: ">", Severity: "critical"}, ""},
		{"valid rule", model.Rule{Name: "test", Metric: "cpu", Operator: ">=", Threshold: 90, Severity: "warning"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.rule.Validate()
			if tt.wantErr == "" && msg != "" {
				t.Errorf("expected valid, got: %s", msg)
			}
			if tt.wantErr != "" && msg != tt.wantErr {
				t.Errorf("expected %q, got %q", tt.wantErr, msg)
			}
		})
	}
}
