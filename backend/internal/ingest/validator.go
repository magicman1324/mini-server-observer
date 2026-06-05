package ingest

import (
	"fmt"
	"math"
	"strings"

	pb "github.com/pingan/hydra/internal/proto"
)

// ValidationError describes a rejected metric sample.
type ValidationError struct {
	SampleName string
	Reason     string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid sample %q: %s", e.SampleName, e.Reason)
}

// ValidateSample checks a metric sample for correctness.
func ValidateSample(s *pb.MetricSample) error {
	if s.Name == "" {
		return &ValidationError{SampleName: "(empty)", Reason: "metric name is required"}
	}
	if len(s.Name) > 256 {
		return &ValidationError{SampleName: s.Name, Reason: "metric name too long (max 256)"}
	}
	if math.IsNaN(s.Value) || math.IsInf(s.Value, 0) {
		return &ValidationError{SampleName: s.Name, Reason: "value is NaN or Inf"}
	}
	if s.Host == "" {
		return &ValidationError{SampleName: s.Name, Reason: "host is required"}
	}
	if len(s.Labels) > 50 {
		return &ValidationError{SampleName: s.Name, Reason: "too many labels (max 50)"}
	}
	for k := range s.Labels {
		if strings.HasPrefix(k, "__") {
			return &ValidationError{SampleName: s.Name, Reason: "label keys must not start with __"}
		}
	}
	return nil
}

// ValidateBatch validates all samples in a batch and returns the accepted/rejected counts.
func ValidateBatch(batch *pb.MetricBatch) (accepted, rejected int, errors []error) {
	for _, s := range batch.Samples {
		if err := ValidateSample(s); err != nil {
			rejected++
			errors = append(errors, err)
		} else {
			accepted++
		}
	}
	return
}
