package collector

import (
	"testing"
)

func TestIsPseudoFS(t *testing.T) {
	tests := []struct {
		fsType   string
		expected bool
	}{
		{"ext4", false},
		{"xfs", false},
		{"btrfs", false},
		{"proc", true},
		{"sysfs", true},
		{"tmpfs", true},
		{"devpts", true},
		{"cgroup", true},
		{"cgroup2", true},
		{"debugfs", true},
	}

	for _, tt := range tests {
		got := isPseudoFS(tt.fsType)
		if got != tt.expected {
			t.Errorf("isPseudoFS(%q) = %v, want %v", tt.fsType, got, tt.expected)
		}
	}
}

func TestDiskCollectorName(t *testing.T) {
	c := NewDiskCollector()
	if c.Name() != "disk" {
		t.Errorf("expected name 'disk', got '%s'", c.Name())
	}
}
