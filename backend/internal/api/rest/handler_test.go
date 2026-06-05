package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pingan/hydra/internal/ingest"
	"github.com/pingan/hydra/internal/repository"
)

func setupTestMux() *http.ServeMux {
	metricRepo := repository.NewMemMetricRepo()
	ruleRepo := repository.NewMemRuleRepo()
	alertRepo := repository.NewMemAlertRepo()

	fanout := ingest.NewFanout(metricRepo, nil)
	streamSrv := ingest.NewStreamServer(fanout)

	mux := http.NewServeMux()
	RegisterRoutes(mux, metricRepo, ruleRepo, alertRepo, streamSrv)
	return mux
}

func TestHealthEndpoint(t *testing.T) {
	mux := setupTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("status = %v", body["status"])
	}
}

func TestCreateRule(t *testing.T) {
	mux := setupTestMux()

	body := `{"name":"High CPU","metric":"cpu_usage_percent","operator":">","threshold":80,"severity":"warning"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rules", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var rule map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&rule)
	if rule["name"] != "High CPU" {
		t.Errorf("name = %v", rule["name"])
	}
	if rule["id"] == "" {
		t.Error("expected non-empty id")
	}
}

func TestListRules(t *testing.T) {
	mux := setupTestMux()

	// Create a rule first
	body := `{"name":"Test Rule","metric":"mem","operator":">","threshold":90,"severity":"critical"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rules", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Now list
	req = httptest.NewRequest(http.MethodGet, "/api/v1/rules?page=1&page_size=10", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if total, _ := resp["total"].(float64); total < 1 {
		t.Errorf("expected at least 1 rule, got total=%v", total)
	}
}

func TestCreateRule_ValidationError(t *testing.T) {
	mux := setupTestMux()

	body := `{"name":"","metric":"cpu","operator":">"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rules", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestIngestMetrics(t *testing.T) {
	mux := setupTestMux()

	batch := `{
		"region": "ap-sh-1",
		"collector_id": "test-col-01",
		"samples": [
			{"name": "cpu_usage_percent", "value": 42.5, "host": "web-01", "region": "ap-sh-1"},
			{"name": "memory_usage_bytes", "value": 8589934592, "host": "web-01", "region": "ap-sh-1"}
		],
		"batch_seq": 1
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", strings.NewReader(batch))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var ack map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&ack)
	if ack["success"] != true {
		t.Errorf("ack success = %v", ack["success"])
	}
	if ack["accepted"] != float64(2) {
		t.Errorf("accepted = %v, want 2", ack["accepted"])
	}
}

func TestListAlerts_Empty(t *testing.T) {
	mux := setupTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?page=1&page_size=10", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if total, _ := resp["total"].(float64); total != 0 {
		t.Errorf("expected 0 alerts, got total=%v", total)
	}
}

func TestGetRule_NotFound(t *testing.T) {
	mux := setupTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules/nonexistent", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
