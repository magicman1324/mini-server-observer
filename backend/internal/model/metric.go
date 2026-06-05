package model

import "time"

// MetricPoint is a single time-series data point stored in ClickHouse.
type MetricPoint struct {
	TS      time.Time
	Name    string
	Value   float64
	Host    string
	Region  string
	Labels  map[string]string
}
