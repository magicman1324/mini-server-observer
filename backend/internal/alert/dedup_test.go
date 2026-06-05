package alert

import (
	"testing"
	"time"
)

func TestDedupTracker_ShouldFire(t *testing.T) {
	dt := NewDedupTracker(100 * time.Millisecond)

	// First fire should be allowed
	if !dt.ShouldFire("rule-1", "web-01") {
		t.Fatal("first fire should be allowed")
	}

	// Second fire within window should be blocked
	if dt.ShouldFire("rule-1", "web-01") {
		t.Fatal("second fire within window should be blocked")
	}

	// Different host should be allowed
	if !dt.ShouldFire("rule-1", "web-02") {
		t.Fatal("different host should be allowed")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)
	if !dt.ShouldFire("rule-1", "web-01") {
		t.Fatal("fire after window should be allowed")
	}
}

func TestDedupTracker_Reset(t *testing.T) {
	dt := NewDedupTracker(time.Hour)

	dt.ShouldFire("rule-1", "web-01")
	dt.Reset("rule-1", "web-01")

	// After reset, should fire again
	if !dt.ShouldFire("rule-1", "web-01") {
		t.Fatal("should fire after reset")
	}
}

func TestDedupTracker_DefaultWindow(t *testing.T) {
	dt := NewDedupTracker(0)
	if dt.window != defaultDedupWindow {
		t.Errorf("expected default window %v, got %v", defaultDedupWindow, dt.window)
	}
}
