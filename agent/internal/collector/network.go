package collector

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pingan/hydra-agent/internal/model"
)

// NetworkCollector reads network interface statistics from /proc/net/dev.
type NetworkCollector struct {
	mu       sync.Mutex
	prev     map[string]netCounters
	firstRun bool
}

type netCounters struct {
	rxBytes   uint64
	txBytes   uint64
	rxPackets uint64
	txPackets uint64
	rxErrors  uint64
	txErrors  uint64
}

func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{
		prev:     make(map[string]netCounters),
		firstRun: true,
	}
}

func (c *NetworkCollector) Name() string { return "network" }

func (c *NetworkCollector) Collect(_ context.Context, batch *model.MetricsBatch) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	current := parseNetDev(string(data))

	if c.firstRun {
		c.prev = current
		c.firstRun = false
		return
	}

	for iface, curr := range current {
		prev, ok := c.prev[iface]
		if !ok {
			continue
		}

		labels := model.Labels{
			"iface":     iface,
			"component": "network",
		}

		// Cumulative counters: current - previous
		batch.Add("network_rx_bytes_total", float64(curr.rxBytes-prev.rxBytes), labels)
		batch.Add("network_tx_bytes_total", float64(curr.txBytes-prev.txBytes), labels)
		batch.Add("network_rx_packets_total", float64(curr.rxPackets-prev.rxPackets), labels)
		batch.Add("network_tx_packets_total", float64(curr.txPackets-prev.txPackets), labels)
		batch.Add("network_rx_errors_total", float64(curr.rxErrors-prev.rxErrors), labels)
		batch.Add("network_tx_errors_total", float64(curr.txErrors-prev.txErrors), labels)
	}

	c.prev = current
}

// parseNetDev parses /proc/net/dev into a map of interface name -> counters.
func parseNetDev(content string) map[string]netCounters {
	result := make(map[string]netCounters)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Skip header lines
		if strings.Contains(line, "Inter-|") || strings.Contains(line, " face |") {
			continue
		}

		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}

		iface := strings.TrimSpace(line[:idx])
		// Skip loopback
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(line[idx+1:])
		if len(fields) < 10 {
			continue
		}

		result[iface] = netCounters{
			rxBytes:   parseUint64Safe(fields[0]),
			rxPackets: parseUint64Safe(fields[1]),
			rxErrors:  parseUint64Safe(fields[2]),
			txBytes:   parseUint64Safe(fields[8]),
			txPackets: parseUint64Safe(fields[9]),
			txErrors:  parseUint64Safe(fields[10]),
		}
	}
	return result
}

func parseUint64Safe(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}
