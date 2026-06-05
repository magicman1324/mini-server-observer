package collector

import (
	"testing"
)

func TestParseNetDev(t *testing.T) {
	content := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 1000000    5000    0    0    0     0          0         0  1000000    5000    0    0    0     0       0          0
  eth0: 50000000   80000   10    0    0     0          0         0 30000000   60000    5    0    0     0       0          0
`

	result := parseNetDev(content)

	// loopback should be skipped
	if _, ok := result["lo"]; ok {
		t.Error("loopback should be skipped")
	}

	eth0, ok := result["eth0"]
	if !ok {
		t.Fatal("eth0 not found")
	}
	if eth0.rxBytes != 50000000 {
		t.Errorf("eth0 rxBytes = %d, want 50000000", eth0.rxBytes)
	}
	if eth0.txBytes != 30000000 {
		t.Errorf("eth0 txBytes = %d, want 30000000", eth0.txBytes)
	}
	if eth0.rxErrors != 10 {
		t.Errorf("eth0 rxErrors = %d, want 10", eth0.rxErrors)
	}
	if eth0.txErrors != 5 {
		t.Errorf("eth0 txErrors = %d, want 5", eth0.txErrors)
	}
}

func TestNetworkCollectorName(t *testing.T) {
	c := NewNetworkCollector()
	if c.Name() != "network" {
		t.Errorf("expected name 'network', got '%s'", c.Name())
	}
}
