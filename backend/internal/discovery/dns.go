package discovery

import (
	"fmt"
	"net"
	"sync"

	"github.com/pingan/hydra/internal/scraper"
)

// DNSDiscovery resolves targets via DNS SRV records.
type DNSDiscovery struct {
	mu       sync.RWMutex
	service  string
	port     string
	fallback []*scraper.Target
}

// NewDNS creates a DNSDiscovery for the given SRV service name.
func NewDNS(service, defaultPort string, fallback []*scraper.Target) *DNSDiscovery {
	return &DNSDiscovery{
		service:  service,
		port:     defaultPort,
		fallback: fallback,
	}
}

// Targets resolves SRV records and returns discovered targets.
func (d *DNSDiscovery) Targets() []*scraper.Target {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, addrs, err := net.LookupSRV("", "", d.service)
	if err != nil || len(addrs) == 0 {
		return d.fallback
	}

	var targets []*scraper.Target
	for _, srv := range addrs {
		port := d.port
		if srv.Port != 0 {
			port = fmt.Sprintf("%d", srv.Port)
		}
		targets = append(targets, &scraper.Target{
			Host: srv.Target,
			Port: port,
		})
	}
	return targets
}
