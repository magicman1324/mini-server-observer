package scraper

import (
	"math"
	"strconv"
	"strings"
	"time"

	pb "github.com/pingan/hydra/internal/proto"
)

// ParsePrometheusText parses Prometheus exposition format text into MetricSamples.
// It handles the HELP, TYPE, and metric lines.
func ParsePrometheusText(text string, host, region string) []*pb.MetricSample {
	var samples []*pb.MetricSample
	ts := time.Now().UnixMilli()

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: metric_name{labels} value [timestamp]
		// or:     metric_name value
		name, labels, value, ok := parseMetricLine(line)
		if !ok {
			continue
		}

		labels["__host__"] = host
		labels["__region__"] = region

		samples = append(samples, &pb.MetricSample{
			TimestampMs: ts,
			Name:        name,
			Value:       value,
			Host:        host,
			Region:      region,
			Labels:      cleanLabels(labels),
		})
	}

	return samples
}

// parseMetricLine extracts name, labels, and value from a Prometheus metric line.
func parseMetricLine(line string) (name string, labels map[string]string, value float64, ok bool) {
	labels = make(map[string]string)

	braceIdx := strings.Index(line, "{")
	spaceIdx := strings.LastIndex(line, " ")
	if spaceIdx < 0 {
		return "", nil, 0, false
	}

	valStr := strings.TrimSpace(line[spaceIdx+1:])
	v, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return "", nil, 0, false
	}

	if braceIdx >= 0 && braceIdx < spaceIdx {
		name = line[:braceIdx]
		labelStr := line[braceIdx+1 : strings.LastIndex(line[:spaceIdx], "}")]
		labels = parseLabels(labelStr)
	} else {
		name = line[:spaceIdx]
	}

	return name, labels, v, true
}

// parseLabels parses Prometheus label pairs: key1="val1",key2="val2"
func parseLabels(s string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(kv[1], `"`)
		result[key] = val
	}
	return result
}

// cleanLabels removes internal labels (prefixed with __) from the output.
func cleanLabels(labels map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range labels {
		if strings.HasPrefix(k, "__") {
			continue
		}
		result[k] = v
	}
	return result
}

// SafeFloat64 ensures the value is a valid float64 (not NaN or Inf).
func SafeFloat64(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
