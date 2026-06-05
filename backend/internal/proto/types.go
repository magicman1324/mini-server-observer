package proto

// MetricSample represents a single metric data point.
type MetricSample struct {
	TimestampMs int64             `json:"timestamp_ms"`
	Name        string            `json:"name"`
	Value       float64           `json:"value"`
	Host        string            `json:"host"`
	Region      string            `json:"region"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// MetricBatch is a collection of metric samples from a regional collector.
type MetricBatch struct {
	Region      string          `json:"region"`
	CollectorID string          `json:"collector_id"`
	Samples     []*MetricSample `json:"samples"`
	BatchSeq    int64           `json:"batch_seq"`
}

// IngestAck acknowledges a batch ingestion.
type IngestAck struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Accepted int32  `json:"accepted"`
	Rejected int32  `json:"rejected"`
}

// Rule represents an alert rule definition.
type Rule struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Metric       string            `json:"metric"`
	Operator     string            `json:"operator"`
	Threshold    float64           `json:"threshold"`
	DurationSec  uint32            `json:"duration_sec"`
	Severity     string            `json:"severity"`
	Expression   string            `json:"expression,omitempty"`
	Enabled      bool              `json:"enabled"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

// Alert represents a triggered alert event.
type Alert struct {
	ID         string            `json:"id"`
	RuleID     string            `json:"rule_id"`
	RuleName   string            `json:"rule_name"`
	Host       string            `json:"host"`
	Region     string            `json:"region"`
	Severity   string            `json:"severity"`
	Metric     string            `json:"metric"`
	Value      float64           `json:"value"`
	Threshold  float64           `json:"threshold"`
	Message    string            `json:"message"`
	Status     string            `json:"status"`
	FiredAt    string            `json:"fired_at"`
	ResolvedAt string            `json:"resolved_at,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// Silence represents an alert suppression window.
type Silence struct {
	ID        string `json:"id"`
	Matchers  string `json:"matchers"`
	StartsAt  string `json:"starts_at"`
	EndsAt    string `json:"ends_at"`
	Comment   string `json:"comment"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

// HostInfo represents a registered monitoring target.
type HostInfo struct {
	Host      string            `json:"host"`
	Region    string            `json:"region"`
	FirstSeen string            `json:"first_seen"`
	LastSeen  string            `json:"last_seen"`
	Labels    map[string]string `json:"labels,omitempty"`
}
