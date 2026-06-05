package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pingan/hydra/internal/api/rest"
	"github.com/pingan/hydra/internal/ingest"
	"github.com/pingan/hydra/internal/model"
	"github.com/pingan/hydra/internal/repository"
)

func main() {
	log.Println("Hydra Central Aggregator starting...")

	metricRepo := repository.NewMemMetricRepo()
	ruleRepo := repository.NewMemRuleRepo()
	alertRepo := repository.NewMemAlertRepo()

	seedRule(ruleRepo)

	alertCh := make(chan []*model.MetricPoint, 100)

	fanout := ingest.NewFanout(metricRepo, alertCh)
	streamSrv := ingest.NewStreamServer(fanout)

	go func() {
		for range alertCh {
			// PR4: evaluate rules against metric points
		}
	}()

	mux := http.NewServeMux()
	rest.RegisterRoutes(mux, metricRepo, ruleRepo, alertRepo, streamSrv)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      rest.CORSMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Central REST API listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	fmt.Println("Central aggregator stopped")
}

func seedRule(repo repository.RuleRepository) {
	rule := &model.Rule{
		Name:        "High CPU Usage",
		Description: "Alert when CPU usage exceeds 80% for more than 30 seconds",
		Metric:      "cpu_usage_percent",
		Operator:    ">",
		Threshold:   80.0,
		DurationSec: 30,
		Severity:    "warning",
		Enabled:     true,
	}
	if err := repo.Create(rule); err != nil {
		log.Printf("Failed to seed rule: %v", err)
	}
}
