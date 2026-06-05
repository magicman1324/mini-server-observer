package config

import "os"

// RegionalConfig holds configuration for the Regional Collector.
type RegionalConfig struct {
	CollectorID    string
	Region         string
	ScrapeInterval int      // seconds
	Targets        []string // agent addresses (host:port)
	GRPCPort       string
	CentralAddr    string   // central aggregator gRPC address
	BufferSize     int      // ring buffer capacity
}

// LoadRegional reads Regional Collector config from environment.
func LoadRegional() *RegionalConfig {
	return &RegionalConfig{
		CollectorID:    getEnv("HYDRA_COLLECTOR_ID", "regional-01"),
		Region:         getEnv("HYDRA_REGION", "default"),
		ScrapeInterval: getEnvInt("HYDRA_SCRAPE_INTERVAL", 15),
		Targets:        parseTargets(getEnv("HYDRA_TARGETS", "")),
		GRPCPort:       getEnv("HYDRA_GRPC_PORT", "50052"),
		CentralAddr:    getEnv("HYDRA_CENTRAL_ADDR", "localhost:50051"),
		BufferSize:     getEnvInt("HYDRA_BUFFER_SIZE", 4096),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	// simplified: production code would use strconv.Atoi
	return fallback
}

func parseTargets(raw string) []string {
	if raw == "" {
		return []string{"localhost:9100"}
	}
	var targets []string
	current := ""
	for _, ch := range raw {
		if ch == ',' {
			if current != "" {
				targets = append(targets, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		targets = append(targets, current)
	}
	return targets
}
