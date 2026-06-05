package collector

import (
	"testing"
)

func TestParseMeminfo(t *testing.T) {
	content := `MemTotal:       16384000 kB
MemFree:         8192000 kB
Buffers:          102400 kB
Cached:          2048000 kB
SReclaimable:     512000 kB
SwapTotal:       2048000 kB
SwapFree:        2048000 kB
`
	result := parseMeminfo(content)

	if result["MemTotal"] != 16384000 {
		t.Errorf("MemTotal = %d, want 16384000", result["MemTotal"])
	}
	if result["MemFree"] != 8192000 {
		t.Errorf("MemFree = %d, want 8192000", result["MemFree"])
	}
	if result["Buffers"] != 102400 {
		t.Errorf("Buffers = %d, want 102400", result["Buffers"])
	}
	if result["Cached"] != 2048000 {
		t.Errorf("Cached = %d, want 2048000", result["Cached"])
	}
}

func TestMemoryCollectorName(t *testing.T) {
	c := NewMemoryCollector()
	if c.Name() != "memory" {
		t.Errorf("expected name 'memory', got '%s'", c.Name())
	}
}
