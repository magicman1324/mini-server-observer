package discovery

import "github.com/pingan/hydra/internal/scraper"

// StaticDiscovery resolves targets from a hardcoded list.
type StaticDiscovery struct {
	targets []*scraper.Target
}

// NewStatic creates a StaticDiscovery from a list of "host:port" strings.
func NewStatic(addrs []string) *StaticDiscovery {
	var targets []*scraper.Target
	for _, addr := range addrs {
		host, port := splitAddr(addr)
		targets = append(targets, &scraper.Target{
			Host: host,
			Port: port,
		})
	}
	return &StaticDiscovery{targets: targets}
}

// Targets returns the current list of scrape targets.
func (d *StaticDiscovery) Targets() []*scraper.Target {
	return d.targets
}

func splitAddr(addr string) (host, port string) {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i], addr[i+1:]
		}
	}
	return addr, "9100"
}
