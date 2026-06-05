package collector

import (
	"testing"
)

func TestCPUCollectorParse(t *testing.T) {
	c := NewCPUCollector()

	// Simulated /proc/stat content
	content := `cpu  4705 0 2356 4497890 1234 0 567 0 0 0
cpu0 1200 0 600 1124000 300 0 150 0 0 0
cpu1 1200 0 600 1124000 300 0 150 0 0 0
`

	total, idle := c.parseCPU(content)
	if total == 0 {
		t.Fatal("expected non-zero total CPU ticks")
	}
	if idle == 0 {
		t.Fatal("expected non-zero idle CPU ticks")
	}
	// idle includes idle+iowait
	if idle >= total {
		t.Errorf("idle (%d) should be less than total (%d)", idle, total)
	}
}

func TestCPUCollectorName(t *testing.T) {
	c := NewCPUCollector()
	if c.Name() != "cpu" {
		t.Errorf("expected name 'cpu', got '%s'", c.Name())
	}
}
