package repository

import (
	"testing"
	"time"

	"github.com/pingan/hydra/internal/model"
)

func TestMemMetricRepo_InsertAndQuery(t *testing.T) {
	repo := NewMemMetricRepo()

	// Insert
	points := []*model.MetricPoint{
		{TS: time.Now(), Name: "cpu", Value: 42.0, Host: "web-01", Region: "ap-sh-1"},
		{TS: time.Now(), Name: "mem", Value: 8e9, Host: "web-01", Region: "ap-sh-1"},
		{TS: time.Now(), Name: "cpu", Value: 45.0, Host: "web-02", Region: "ap-sh-1"},
	}
	if err := repo.InsertBatch(points); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	// Query by name + host
	results, err := repo.Query("cpu", "web-01", time.Time{}, time.Time{}, 100)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 42.0 {
		t.Errorf("value = %g, want 42.0", results[0].Value)
	}

	// Query all
	results, err = repo.Query("", "", time.Time{}, time.Time{}, 100)
	if err != nil {
		t.Fatalf("query all failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// List names
	names, err := repo.ListNames()
	if err != nil {
		t.Fatalf("list names failed: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d: %v", len(names), names)
	}
}

func TestMemMetricRepo_Query_Empty(t *testing.T) {
	repo := NewMemMetricRepo()
	results, err := repo.Query("cpu", "", time.Time{}, time.Time{}, 10)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
