package model

import "time"

// Silence defines an alert suppression time window.
type Silence struct {
	ID        string    `json:"id"`
	Matchers  string    `json:"matchers"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	Comment   string    `json:"comment"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}
