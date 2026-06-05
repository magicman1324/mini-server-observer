package collector

import (
	"context"

	"github.com/pingan/hydra-agent/internal/model"
)

// DiskCollector reads disk usage for each mounted filesystem via /proc/mounts + statfs.
type DiskCollector struct{}

func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

func (c *DiskCollector) Name() string { return "disk" }

func (c *DiskCollector) Collect(ctx context.Context, batch *model.MetricsBatch) {
	collectDiskMetrics(batch)
}

func isPseudoFS(fsType string) bool {
	switch fsType {
	case "proc", "sysfs", "devpts", "tmpfs", "devtmpfs",
		"cgroup", "cgroup2", "pstore", "bpf", "debugfs",
		"tracefs", "hugetlbfs", "fusectl", "configfs",
		"securityfs", "mqueue", "binfmt_misc":
		return true
	}
	return false
}
