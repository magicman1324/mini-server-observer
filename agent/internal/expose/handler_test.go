package expose

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pingan/hydra-agent/internal/collector"
	"github.com/pingan/hydra-agent/internal/model"
)

type mockCollector struct {
	name    string
	samples []struct {
		name  string
		value float64
	}
}

func (m *mockCollector) Name() string                                  { return m.name }
func (m *mockCollector) Collect(_ context.Context, b *model.MetricsBatch) {
	for _, s := range m.samples {
		b.Add(s.name, s.value, nil)
	}
}

func TestMetricsHandler(t *testing.T) {
	mc := &mockCollector{
		name: "test",
		samples: []struct {
			name  string
			value float64
		}{
			{name: "cpu_usage_percent", value: 42.5},
			{name: "memory_usage_bytes", value: 8589934592},
		},
	}

	collectors := []collector.Collector{mc}
	registry := model.NewMetricRegistry("test-host", "test-region")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := MetricsHandler(collectors, registry)
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "cpu_usage_percent") {
		t.Error("response missing cpu_usage_percent")
	}
	if !strings.Contains(body, "memory_usage_bytes") {
		t.Error("response missing memory_usage_bytes")
	}
	if !strings.Contains(body, `hostname="test-host"`) {
		t.Error("response missing hostname label")
	}
	if !strings.Contains(body, "TYPE cpu_usage_percent gauge") {
		t.Error("response missing TYPE line")
	}
	if !strings.Contains(body, "HELP cpu_usage_percent") {
		t.Error("response missing HELP line")
	}
	// Verify Prometheus text format contains the value
	if !strings.Contains(body, "42.5") {
		t.Error("response missing metric value 42.5")
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Error("health response missing ok")
	}
}
