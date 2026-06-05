package rest

import (
	"net/http"
	"strings"

	"github.com/pingan/hydra/internal/ingest"
	"github.com/pingan/hydra/internal/repository"
)

// CORSMiddleware wraps an http.Handler with CORS headers.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RegisterRoutes sets up all REST API routes on the given ServeMux.
func RegisterRoutes(mux *http.ServeMux, metricRepo repository.MetricRepository,
	ruleRepo repository.RuleRepository, alertRepo repository.AlertRepository,
	streamSrv *ingest.StreamServer) {

	h := &Handler{
		metricRepo: metricRepo,
		ruleRepo:   ruleRepo,
		alertRepo:  alertRepo,
		streamSrv:  streamSrv,
	}

	// Auth (public)
	mux.HandleFunc("/api/v1/auth/login", h.Login)

	// Health (public)
	mux.HandleFunc("/api/v1/health", h.Health)

	// Metrics
	mux.HandleFunc("/api/v1/metrics/query", h.QueryMetrics)
	mux.HandleFunc("/api/v1/metrics/names", h.ListMetricNames)

	// Hosts
	mux.HandleFunc("/api/v1/hosts", h.ListHosts)

	// Rules CRUD
	mux.HandleFunc("/api/v1/rules", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListRules(w, r)
		case http.MethodPost:
			h.CreateRule(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	// /api/v1/rules/{id}
	mux.HandleFunc("/api/v1/rules/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path: /api/v1/rules/{id}
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/rules/")
		if path == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetRule(w, r, path)
		case http.MethodPut:
			h.UpdateRule(w, r, path)
		case http.MethodDelete:
			h.DeleteRule(w, r, path)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Alerts
	mux.HandleFunc("/api/v1/alerts/stats", h.AlertStats)
	mux.HandleFunc("/api/v1/alerts", h.ListAlerts)
	// /api/v1/alerts/{id}
	mux.HandleFunc("/api/v1/alerts/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/alerts/")
		if path == "" || path == "stats" {
			return
		}
		h.GetAlert(w, r, path)
	})

	// Silences
	mux.HandleFunc("/api/v1/silences", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListSilences(w, r)
		case http.MethodPost:
			h.CreateSilence(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/silences/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/silences/")
		if path == "" {
			return
		}
		h.DeleteSilence(w, r, path)
	})

	// Ingest
	mux.HandleFunc("/api/v1/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.IngestMetrics(w, r)
	})
}

// Handler holds all dependencies for REST endpoints.
type Handler struct {
	metricRepo repository.MetricRepository
	ruleRepo   repository.RuleRepository
	alertRepo  repository.AlertRepository
	streamSrv  *ingest.StreamServer
}
