package model

import "time"

// Host represents a registered monitoring target.
type Host struct {
	Host      string            `json:"host"`
	Region    string            `json:"region"`
	Labels    map[string]string `json:"labels,omitempty"`
	FirstSeen time.Time         `json:"first_seen"`
	LastSeen  time.Time         `json:"last_seen"`
}
