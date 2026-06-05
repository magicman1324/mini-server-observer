package scraper

import (
	"strings"
	"testing"
)

const sampleMetrics = `# HELP cpu_usage_percent Hydra agent collected metric
# TYPE cpu_usage_percent gauge
cpu_usage_percent{hostname="web-01",region="ap-sh-1"} 42.5
# HELP memory_usage_bytes Hydra agent collected metric
# TYPE memory_usage_bytes gauge
memory_usage_bytes{hostname="web-01",region="ap-sh-1"} 8.589934592e+09
# HELP disk_usage_percent Hydra agent collected metric
# TYPE disk_usage_percent gauge
disk_usage_percent{hostname="web-01",region="ap-sh-1",mount="/"} 72.3
# HELP network_rx_bytes_total Hydra agent collected metric
# TYPE network_rx_bytes_total gauge
network_rx_bytes_total{hostname="web-01",region="ap-sh-1",iface="eth0"} 50000000
`

func TestParsePrometheusText(t *testing.T) {
	samples := ParsePrometheusText(sampleMetrics, "web-01", "ap-sh-1")

	if len(samples) != 4 {
		t.Fatalf("expected 4 samples, got %d", len(samples))
	}

	// Check CPU
	if samples[0].Name != "cpu_usage_percent" {
		t.Errorf("sample 0 name = %q, want cpu_usage_percent", samples[0].Name)
	}
	if samples[0].Value != 42.5 {
		t.Errorf("sample 0 value = %g, want 42.5", samples[0].Value)
	}
	if samples[0].Host != "web-01" {
		t.Errorf("sample 0 host = %q, want web-01", samples[0].Host)
	}

	// Check memory (scientific notation)
	if samples[1].Name != "memory_usage_bytes" {
		t.Errorf("sample 1 name = %q", samples[1].Name)
	}
	if samples[1].Value < 8e9 {
		t.Errorf("sample 1 value = %g, want ~8.59e9", samples[1].Value)
	}

	// Check disk with mount label
	if samples[2].Name != "disk_usage_percent" {
		t.Errorf("sample 2 name = %q", samples[2].Name)
	}

	// Check network
	if samples[3].Name != "network_rx_bytes_total" {
		t.Errorf("sample 3 name = %q", samples[3].Name)
	}
}

func TestParsePrometheusTextEmpty(t *testing.T) {
	samples := ParsePrometheusText("", "host", "region")
	if len(samples) != 0 {
		t.Errorf("expected 0 samples from empty input, got %d", len(samples))
	}
}

func TestParsePrometheusTextComments(t *testing.T) {
	input := `# HELP test_metric a test
# TYPE test_metric gauge
`
	samples := ParsePrometheusText(input, "h", "r")
	if len(samples) != 0 {
		t.Errorf("expected 0 samples from comments-only input, got %d", len(samples))
	}
}

func TestParseMetricLine(t *testing.T) {
	tests := []struct {
		line     string
		wantName string
		wantVal  float64
	}{
		{`cpu_usage_percent{hostname="web-01"} 42.5`, "cpu_usage_percent", 42.5},
		{`memory_bytes 1.073741824e+09`, "memory_bytes", 1.073741824e+09},
		{`gauge_no_labels 0`, "gauge_no_labels", 0},
	}

	for _, tt := range tests {
		name, labels, val, ok := parseMetricLine(tt.line)
		if !ok {
			t.Errorf("parseMetricLine(%q) failed", tt.line)
			continue
		}
		if name != tt.wantName {
			t.Errorf("name = %q, want %q", name, tt.wantName)
		}
		if val != tt.wantVal {
			t.Errorf("val = %g, want %g", val, tt.wantVal)
		}
		_ = labels
	}
}

func TestParseLabels(t *testing.T) {
	result := parseLabels(`hostname="web-01",region="ap-sh-1",mount="/data"`)
	if result["hostname"] != "web-01" {
		t.Errorf("hostname = %q", result["hostname"])
	}
	if result["region"] != "ap-sh-1" {
		t.Errorf("region = %q", result["region"])
	}
	if result["mount"] != "/data" {
		t.Errorf("mount = %q", result["mount"])
	}
}

func TestSafeFloat64(t *testing.T) {
	if v := SafeFloat64(42.5); v != 42.5 {
		t.Errorf("SafeFloat64(42.5) = %g", v)
	}
	if v := SafeFloat64(0); v != 0 {
		t.Errorf("SafeFloat64(0) = %g", v)
	}
}

func TestCleanLabels(t *testing.T) {
	labels := map[string]string{
		"hostname":  "web-01",
		"region":    "ap-sh-1",
		"__host__":  "internal",
		"__region__": "internal",
	}
	cleaned := cleanLabels(labels)
	if _, ok := cleaned["__host__"]; ok {
		t.Error("__host__ should have been cleaned")
	}
	if cleaned["hostname"] != "web-01" {
		t.Error("hostname should remain")
	}
	if len(cleaned) != 2 {
		t.Errorf("expected 2 clean labels, got %d", len(cleaned))
	}
}

func TestParsePrometheusText_InvalidLines(t *testing.T) {
	input := `this is not a valid metric line
also_invalid
cpu_usage_percent{hostname="ok"} 99.9
`
	samples := ParsePrometheusText(input, "host", "region")
	if len(samples) != 1 {
		t.Fatalf("expected 1 valid sample, got %d", len(samples))
	}
	if samples[0].Name != "cpu_usage_percent" {
		t.Errorf("name = %q", samples[0].Name)
	}
}

// TestParsePrometheusTextWithoutLabels tests metric lines without label braces.
func TestParsePrometheusTextWithoutLabels(t *testing.T) {
	input := `simple_metric 123.45
`
	samples := ParsePrometheusText(input, "myhost", "myregion")
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(samples))
	}
	if samples[0].Name != "simple_metric" {
		t.Errorf("name = %q, want simple_metric", samples[0].Name)
	}
	if samples[0].Value != 123.45 {
		t.Errorf("value = %g, want 123.45", samples[0].Value)
	}
	if samples[0].Host != "myhost" {
		t.Errorf("host = %q, want myhost", samples[0].Host)
	}

	// Verify labels don't leak internal keys
	body := strings.Builder{}
	for k := range samples[0].Labels {
		body.WriteString(k + "=" + samples[0].Labels[k] + " ")
	}
	if strings.Contains(body.String(), "__") {
		t.Errorf("internal label leaked in output: %s", body.String())
	}
}
