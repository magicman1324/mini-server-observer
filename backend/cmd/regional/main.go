package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pingan/hydra/internal/buffer"
	"github.com/pingan/hydra/internal/config"
	"github.com/pingan/hydra/internal/discovery"
	"github.com/pingan/hydra/internal/grpcstream"
	"github.com/pingan/hydra/internal/scraper"
)

func main() {
	cfg := config.LoadRegional()
	log.Printf("Hydra Regional Collector starting [%s] region=%s", cfg.CollectorID, cfg.Region)

	// Service discovery
	d := discovery.NewStatic(cfg.Targets)
	targets := d.Targets()
	log.Printf("Discovered %d targets: %v", len(targets), cfg.Targets)

	// Ring buffer for metric batching
	ringBuf := buffer.NewRingBuffer(cfg.BufferSize)

	// gRPC stream client (toward Central Aggregator)
	streamClient := grpcstream.NewClient(cfg.CentralAddr, cfg.Region, cfg.CollectorID, ringBuf)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamClient.Start(ctx)

	// Scraper
	scraper := scraper.New(cfg.Region)

	// Main scrape loop
	ticker := time.NewTicker(time.Duration(cfg.ScrapeInterval) * time.Second)
	defer ticker.Stop()

	// Immediate first scrape
	runScrapeCycle(scraper, targets, ringBuf)

	go func() {
		for range ticker.C {
			runScrapeCycle(scraper, targets, ringBuf)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down regional collector...")
	ticker.Stop()
	cancel()
	streamClient.Stop()
	fmt.Println("Regional collector stopped")
}

func runScrapeCycle(s *scraper.Scraper, targets []*scraper.Target, buf *buffer.RingBuffer) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	samples := s.ScrapeAll(ctx, targets)
	accepted := 0
	for _, sample := range samples {
		if buf.Push(sample) {
			accepted++
		}
	}

	log.Printf("Scrape cycle: %d samples from %d targets in %s (buffered: %d/%d)",
		len(samples), len(targets), time.Since(start).Round(time.Millisecond),
		accepted, buf.Len())
}
