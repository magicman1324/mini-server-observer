package discovery

import (
	"testing"
)

func TestStaticDiscovery(t *testing.T) {
	addrs := []string{
		"web-01:9100",
		"web-02:9100",
		"db-01:9101",
	}

	d := NewStatic(addrs)
	targets := d.Targets()

	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	if targets[0].Host != "web-01" || targets[0].Port != "9100" {
		t.Errorf("target[0] = %s:%s", targets[0].Host, targets[0].Port)
	}
	if targets[2].Host != "db-01" || targets[2].Port != "9101" {
		t.Errorf("target[2] = %s:%s", targets[2].Host, targets[2].Port)
	}
}

func TestSplitAddr(t *testing.T) {
	tests := []struct {
		addr     string
		wantHost string
		wantPort string
	}{
		{"web-01:9100", "web-01", "9100"},
		{"192.168.1.1:8080", "192.168.1.1", "8080"},
		{"localhost", "localhost", "9100"},
		{"host:port:extra", "host:port", "extra"},
	}

	for _, tt := range tests {
		host, port := splitAddr(tt.addr)
		if host != tt.wantHost || port != tt.wantPort {
			t.Errorf("splitAddr(%q) = (%q, %q), want (%q, %q)",
				tt.addr, host, port, tt.wantHost, tt.wantPort)
		}
	}
}
