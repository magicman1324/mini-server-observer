package collector

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/pingan/hydra-agent/internal/cgroup"
	"github.com/pingan/hydra-agent/internal/model"
)

// MemoryCollector reads memory usage from /proc/meminfo (or cgroup v2 memory.* files).
type MemoryCollector struct{}

func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

func (c *MemoryCollector) Name() string { return "memory" }

func (c *MemoryCollector) Collect(_ context.Context, batch *model.MetricsBatch) {
	if cgroup.IsV2() {
		c.collectV2(batch)
	} else {
		c.collectV1(batch)
	}
}

func (c *MemoryCollector) collectV2(batch *model.MetricsBatch) {
	current, err := readUint64(cgroup.MemoryCurrentPath())
	if err != nil {
		return
	}
	batch.Add("memory_usage_bytes", float64(current), model.Labels{"component": "memory"})

	maxPath := cgroup.MemoryMaxPath()
	if maxPath != "" {
		maxBytes, err := readUint64(maxPath)
		if err == nil && maxBytes > 0 {
			batch.Add("memory_limit_bytes", float64(maxBytes), nil)
			percent := float64(current) / float64(maxBytes) * 100.0
			batch.Add("memory_usage_percent", percent, nil)
		}
	}

	// OOM kill counter from memory.events
	eventsPath := cgroup.MemoryStatPath()
	if eventsPath != "" {
		oomKills := c.parseOOMKills(eventsPath)
		batch.Add("memory_oom_kills_total", float64(oomKills), model.Labels{"component": "memory"})
	}
}

func (c *MemoryCollector) collectV1(batch *model.MetricsBatch) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}
	info := parseMeminfo(string(data))

	total := info["MemTotal"]
	free := info["MemFree"]
	buffers := info["Buffers"]
	cached := info["Cached"]
	sReclaimable := info["SReclaimable"]

	used := total - free - buffers - cached - sReclaimable

	batch.Add("memory_total_bytes", float64(total*1024), model.Labels{"component": "memory"})
	batch.Add("memory_usage_bytes", float64(used*1024), model.Labels{"component": "memory"})
	if total > 0 {
		batch.Add("memory_usage_percent", float64(used)/float64(total)*100.0, nil)
	}
}

func (c *MemoryCollector) parseOOMKills(path string) uint64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "oom_kill ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				v, _ := strconv.ParseUint(parts[1], 10, 64)
				return v
			}
		}
	}
	return 0
}

func readUint64(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	// cgroup v2 files often contain "max" for unlimited
	text := strings.TrimSpace(string(data))
	if text == "max" {
		return 0, nil
	}
	return strconv.ParseUint(text, 10, 64)
}

func parseMeminfo(content string) map[string]uint64 {
	result := make(map[string]uint64)
	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(parts[1])
		valStr = strings.TrimSuffix(valStr, " kB")
		valStr = strings.TrimSpace(valStr)
		v, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			continue
		}
		result[key] = v
	}
	return result
}
