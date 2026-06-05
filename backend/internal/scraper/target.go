package scraper

// Target represents a scrape target (an agent's /metrics endpoint).
type Target struct {
	Host   string
	Port   string
	Labels map[string]string
}

// Addr returns the address string (host:port) for the target.
func (t *Target) Addr() string {
	return t.Host + ":" + t.Port
}
