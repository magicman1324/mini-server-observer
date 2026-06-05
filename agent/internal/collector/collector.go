package collector

import (
	"context"

	"github.com/pingan/hydra-agent/internal/model"
)

// Collector defines the interface for a metric collection module.
// Each collector gathers a specific category of system metrics (CPU, memory, disk, network).
type Collector interface {
	// Name returns a unique identifier for this collector (e.g. "cpu", "memory").
	Name() string

	// Collect gathers metrics and appends them to the provided batch.
	Collect(ctx context.Context, batch *model.MetricsBatch)
}
