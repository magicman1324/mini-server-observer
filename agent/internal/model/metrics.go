package model

import "sync"

// MetricsBatch accumulates metric samples during a single collection cycle.
type MetricsBatch struct {
	mu      sync.Mutex
	Samples []MetricSample
}

// MetricSample represents a single metric data point with associated labels.
type MetricSample struct {
	Name   string
	Value  float64
	Labels Labels
}

// Labels is a map of string key-value pairs attached to a metric sample.
type Labels map[string]string

// MetricRegistry holds per-agent metadata (hostname, region).
type MetricRegistry struct {
	Hostname string
	Region   string
	mu       sync.RWMutex
}

func NewMetricRegistry(hostname, region string) *MetricRegistry {
	return &MetricRegistry{
		Hostname: hostname,
		Region:   region,
	}
}

// Add appends a metric sample to the batch in a thread-safe manner.
func (b *MetricsBatch) Add(name string, value float64, labels Labels) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Samples = append(b.Samples, MetricSample{
		Name:   name,
		Value:  value,
		Labels: labels,
	})
}

// Len returns the number of samples in the batch.
func (b *MetricsBatch) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.Samples)
}
