package scraper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	pb "github.com/pingan/hydra/internal/proto"
)

// Scraper is responsible for HTTP-scraping metrics from agents.
type Scraper struct {
	client  *http.Client
	region  string
	timeout time.Duration
}

// New creates a Scraper with default settings.
func New(region string) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		region:  region,
		timeout: 10 * time.Second,
	}
}

// ScrapeTarget fetches /metrics from a single target and parses the response.
func (s *Scraper) ScrapeTarget(ctx context.Context, target *Target) ([]*pb.MetricSample, error) {
	url := fmt.Sprintf("http://%s/metrics", target.Addr())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB limit
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	samples := ParsePrometheusText(string(body), target.Host, s.region)
	return samples, nil
}

// ScrapeAll fetches metrics from all targets concurrently.
func (s *Scraper) ScrapeAll(ctx context.Context, targets []*Target) []*pb.MetricSample {
	var allSamples []*pb.MetricSample

	type result struct {
		samples []*pb.MetricSample
		err     error
		target  string
	}

	results := make(chan result, len(targets))

	for _, t := range targets {
		go func(target *Target) {
			samples, err := s.ScrapeTarget(ctx, target)
			results <- result{samples: samples, err: err, target: target.Addr()}
		}(t)
	}

	for range targets {
		r := <-results
		if r.err != nil {
			log.Printf("scrape error [%s]: %v", r.target, r.err)
			continue
		}
		allSamples = append(allSamples, r.samples...)
	}

	return allSamples
}
