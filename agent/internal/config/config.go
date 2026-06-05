package config

import "os"

// Config holds agent configuration sourced from environment variables.
type Config struct {
	ListenAddr string
	Region     string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		ListenAddr: getEnv("HYDRA_LISTEN_ADDR", ":9100"),
		Region:     getEnv("HYDRA_REGION", "default"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
