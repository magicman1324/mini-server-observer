package collector

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pingan/hydra-agent/internal/cgroup"
	"github.com/pingan/hydra-agent/internal/model"
)

// CPUCollector reads CPU usage from /proc/stat (or cgroup cpu.stat on v2).
// It tracks cumulative CPU ticks between samples to compute a delta-based percentage.
type CPUCollector struct {
	mu          sync.Mutex
	prevTotal   uint64
	prevIdle    uint64
	initialized bool
}

func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

func (c *CPUCollector) Name() string { return "cpu" }

func (c *CPUCollector) Collect(_ context.Context, batch *model.MetricsBatch) {
	path := cgroup.CPUStatPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	total, idle := c.parseCPU(string(data))
	if !c.initialized {
		c.prevTotal = total
		c.prevIdle = idle
		c.initialized = true
		return
	}

	deltaTotal := total - c.prevTotal
	deltaIdle := idle - c.prevIdle
	c.prevTotal = total
	c.prevIdle = idle

	if deltaTotal == 0 {
		return
	}

	percentUsed := float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100.0

	batch.Add("cpu_usage_percent", percentUsed, model.Labels{"component": "cpu"})
}

// parseCPU extracts total and idle CPU tick counts from /proc/stat.
// Format: cpu  user nice system idle iowait irq softirq steal ...
func (c *CPUCollector) parseCPU(content string) (total, idle uint64) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		// fields[0] = "cpu", fields[1..] = tick counts
		nums := fields[1:]
		for i, s := range nums {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return 0, 0
			}
			total += v
			// fields[3] = idle, fields[4] = iowait (also considered idle in standard Linux accounting)
			if i == 3 || i == 4 {
				idle += v
			}
		}
		return total, idle
	}
	return 0, 0
}
