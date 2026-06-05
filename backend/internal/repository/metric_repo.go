package repository

import (
	"sort"
	"sync"
	"time"

	"github.com/pingan/hydra/internal/model"
)

// MetricRepository defines the interface for time-series metric storage.
type MetricRepository interface {
	InsertBatch(points []*model.MetricPoint) error
	Query(name, host string, start, end time.Time, limit int) ([]*model.MetricPoint, error)
	ListNames() ([]string, error)
}

// MemMetricRepo is an in-memory implementation for development/testing.
type MemMetricRepo struct {
	mu     sync.RWMutex
	points []*model.MetricPoint
}

func NewMemMetricRepo() *MemMetricRepo {
	return &MemMetricRepo{points: make([]*model.MetricPoint, 0)}
}

func (r *MemMetricRepo) InsertBatch(points []*model.MetricPoint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.points = append(r.points, points...)
	return nil
}

func (r *MemMetricRepo) Query(name, host string, start, end time.Time, limit int) ([]*model.MetricPoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*model.MetricPoint
	for _, p := range r.points {
		if name != "" && p.Name != name {
			continue
		}
		if host != "" && p.Host != host {
			continue
		}
		if !start.IsZero() && p.TS.Before(start) {
			continue
		}
		if !end.IsZero() && p.TS.After(end) {
			continue
		}
		result = append(result, p)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TS.Before(result[j].TS)
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *MemMetricRepo) ListNames() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	var names []string
	for _, p := range r.points {
		if !seen[p.Name] {
			seen[p.Name] = true
			names = append(names, p.Name)
		}
	}
	return names, nil
}
