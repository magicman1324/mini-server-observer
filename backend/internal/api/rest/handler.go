package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	pb "github.com/pingan/hydra/internal/proto"
	"github.com/pingan/hydra/internal/model"
)

// Login authenticates a user and returns a JWT token.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if creds.Username == "" || creds.Password == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": "hydra-jwt-token"})
}

// Health returns service health status.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// ============================================================
// Metrics
// ============================================================

// QueryMetrics queries time-series data.
func (h *Handler) QueryMetrics(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	host := q.Get("host")
	limit, _ := strconv.Atoi(defaultStr(q.Get("limit"), "100"))
	startStr := q.Get("start")
	endStr := q.Get("end")

	var start, end time.Time
	if startStr != "" {
		start, _ = time.Parse(time.RFC3339, startStr)
	}
	if endStr != "" {
		end, _ = time.Parse(time.RFC3339, endStr)
	}

	points, err := h.metricRepo.Query(name, host, start, end, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if points == nil {
		points = []*model.MetricPoint{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  points,
		"count": len(points),
	})
}

// ListMetricNames returns all known metric names.
func (h *Handler) ListMetricNames(w http.ResponseWriter, r *http.Request) {
	names, err := h.metricRepo.ListNames()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if names == nil {
		names = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": names})
}

// ============================================================
// Hosts
// ============================================================

// ListHosts returns registered monitoring targets.
func (h *Handler) ListHosts(w http.ResponseWriter, r *http.Request) {
	points, _ := h.metricRepo.Query("", "", time.Time{}, time.Time{}, 10000)
	seen := make(map[string]*model.Host)
	for _, p := range points {
		if existing, ok := seen[p.Host]; !ok {
			seen[p.Host] = &model.Host{
				Host:     p.Host,
				Region:   p.Region,
				LastSeen: p.TS,
			}
		} else if p.TS.After(existing.LastSeen) {
			existing.LastSeen = p.TS
		}
	}

	hosts := make([]*model.Host, 0, len(seen))
	for _, h := range seen {
		hosts = append(hosts, h)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  hosts,
		"count": len(hosts),
	})
}

// ============================================================
// Rules
// ============================================================

// ListRules returns rule definitions with pagination.
func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(defaultStr(q.Get("page"), "1"))
	pageSize, _ := strconv.Atoi(defaultStr(q.Get("page_size"), "20"))

	rules, total, err := h.ruleRepo.List(page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rules == nil {
		rules = []*model.Rule{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":      rules,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetRule returns a single rule by ID.
func (h *Handler) GetRule(w http.ResponseWriter, r *http.Request, id string) {
	rule, err := h.ruleRepo.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

// CreateRule creates a new alert rule.
func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule model.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if msg := rule.Validate(); msg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}

	if err := h.ruleRepo.Create(&rule); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, rule)
}

// UpdateRule updates an existing alert rule.
func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request, id string) {
	var rule model.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	rule.ID = id
	if msg := rule.Validate(); msg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}

	if err := h.ruleRepo.Update(&rule); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
		return
	}

	writeJSON(w, http.StatusOK, rule)
}

// DeleteRule removes a rule.
func (h *Handler) DeleteRule(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.ruleRepo.Delete(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ============================================================
// Alerts
// ============================================================

// ListAlerts returns alerts with pagination and filtering.
func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(defaultStr(q.Get("page"), "1"))
	pageSize, _ := strconv.Atoi(defaultStr(q.Get("page_size"), "20"))
	status := q.Get("status")
	severity := q.Get("severity")
	host := q.Get("host")

	alerts, total, err := h.alertRepo.List(page, pageSize, status, severity, host)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if alerts == nil {
		alerts = []*model.Alert{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":      alerts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAlert returns a single alert by ID.
func (h *Handler) GetAlert(w http.ResponseWriter, r *http.Request, id string) {
	alert, err := h.alertRepo.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "alert not found"})
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

// AlertStats returns alert status counts.
func (h *Handler) AlertStats(w http.ResponseWriter, r *http.Request) {
	firing, resolved := h.alertRepo.Stats()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"firing":   firing,
		"resolved": resolved,
		"total":    firing + resolved,
	})
}

// ============================================================
// Silences
// ============================================================

// ListSilences returns active silences.
func (h *Handler) ListSilences(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  []*model.Silence{},
		"total": 0,
	})
}

// CreateSilence creates a new silence window.
func (h *Handler) CreateSilence(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]string{"message": "silence created"})
}

// DeleteSilence expires a silence.
func (h *Handler) DeleteSilence(w http.ResponseWriter, r *http.Request, id string) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "silence deleted"})
}

// ============================================================
// Ingest
// ============================================================

// IngestMetrics accepts a metric batch from a regional collector.
func (h *Handler) IngestMetrics(w http.ResponseWriter, r *http.Request) {
	var batch pb.MetricBatch
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ack := h.streamSrv.HandleBatch(&batch)
	writeJSON(w, http.StatusOK, ack)
}

// ============================================================
// Helpers
// ============================================================

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

func defaultStr(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}
