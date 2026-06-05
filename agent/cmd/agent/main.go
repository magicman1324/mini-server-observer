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

	"github.com/pingan/hydra-agent/internal/collector"
	"github.com/pingan/hydra-agent/internal/config"
	"github.com/pingan/hydra-agent/internal/expose"
	"github.com/pingan/hydra-agent/internal/model"
)

func main() {
	cfg := config.Load()

	hostname, _ := os.Hostname()
	log.Printf("Hydra Agent starting on %s (hostname: %s)", cfg.ListenAddr, hostname)

	collectors := []collector.Collector{
		collector.NewCPUCollector(),
		collector.NewMemoryCollector(),
		collector.NewDiskCollector(),
		collector.NewNetworkCollector(),
	}

	registry := model.NewMetricRegistry(hostname, cfg.Region)

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", expose.MetricsHandler(collectors, registry))
	mux.HandleFunc("/health", expose.HealthHandler())

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Agent metrics endpoint: http://%s/metrics", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown failed: %v", err)
	}
	fmt.Println("Agent stopped")
}
