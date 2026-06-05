package expose

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pingan/hydra-agent/internal/collector"
	"github.com/pingan/hydra-agent/internal/model"
)

// MetricsHandler returns an http.HandlerFunc that triggers a full collection
// cycle and renders the results in Prometheus exposition format.
func MetricsHandler(collectors []collector.Collector, registry *model.MetricRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		batch := &model.MetricsBatch{}
		for _, c := range collectors {
			c.Collect(ctx, batch)
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		writePrometheusText(w, batch, registry)
	}
}

// HealthHandler returns a simple liveness check.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

// writePrometheusText renders metric samples in Prometheus exposition format.
func writePrometheusText(w http.ResponseWriter, batch *model.MetricsBatch, registry *model.MetricRegistry) {
	for _, s := range batch.Samples {
		// Build Prometheus HELP/TYPE lines (write once per unique metric name).
		// For simplicity, we write TYPE and the metric on each line for now.
		fmt.Fprintf(w, "# HELP %s Hydra agent collected metric\n", s.Name)
		fmt.Fprintf(w, "# TYPE %s gauge\n", s.Name)

		// Build label string
		labels := fmt.Sprintf(`hostname="%s",region="%s"`, registry.Hostname, registry.Region)
		for k, v := range s.Labels {
			labels += fmt.Sprintf(`,%s="%s"`, k, v)
		}

		fmt.Fprintf(w, "%s{%s} %g\n", s.Name, labels, s.Value)
	}
}
