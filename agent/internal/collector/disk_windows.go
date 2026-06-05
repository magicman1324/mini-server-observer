//go:build windows

package collector

import "github.com/pingan/hydra-agent/internal/model"

func collectDiskMetrics(batch *model.MetricsBatch) {
	// Disk collection via statfs is not supported on Windows.
	// The agent is designed for Linux servers. On Windows
	// development machines, disk metrics will be empty.
}
