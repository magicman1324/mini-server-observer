package alert

import (
	"sync"
	"time"
)

const defaultDedupWindow = 60 * time.Second

// DedupTracker prevents duplicate alert firings within a configurable window.
type DedupTracker struct {
	mu      sync.Mutex
	lastFired map[string]time.Time // key = ruleID + ":" + host
	window  time.Duration
}

// NewDedupTracker creates a dedup tracker with the given window.
func NewDedupTracker(window time.Duration) *DedupTracker {
	if window <= 0 {
		window = defaultDedupWindow
	}
	return &DedupTracker{
		lastFired: make(map[string]time.Time),
		window:    window,
	}
}

// ShouldFire returns true if enough time has passed since the last alert for this key.
func (d *DedupTracker) ShouldFire(ruleID, host string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := ruleID + ":" + host
	last, ok := d.lastFired[key]
	if ok && time.Since(last) < d.window {
		return false
	}
	d.lastFired[key] = time.Now()
	return true
}

// Reset clears the dedup state for a given key (e.g., on resolution).
func (d *DedupTracker) Reset(ruleID, host string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.lastFired, ruleID+":"+host)
}
