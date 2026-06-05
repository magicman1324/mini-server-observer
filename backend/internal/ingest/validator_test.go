package ingest

import (
	"testing"

	pb "github.com/pingan/hydra/internal/proto"
)

func TestValidateSample_OK(t *testing.T) {
	s := &pb.MetricSample{
		Name:  "cpu_usage_percent",
		Value: 42.5,
		Host:  "web-01",
	}
	if err := ValidateSample(s); err != nil {
		t.Errorf("expected valid sample, got: %v", err)
	}
}

func TestValidateSample_EmptyName(t *testing.T) {
	s := &pb.MetricSample{Name: "", Value: 1.0, Host: "h"}
	if err := ValidateSample(s); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestValidateSample_EmptyHost(t *testing.T) {
	s := &pb.MetricSample{Name: "cpu", Value: 1.0, Host: ""}
	if err := ValidateSample(s); err == nil {
		t.Error("expected error for empty host")
	}
}

func TestValidateSample_LongName(t *testing.T) {
	// Build a name that exceeds 256 characters
	longName := ""
	for i := 0; i < 300; i++ {
		longName += "a"
	}
	s := &pb.MetricSample{Name: longName, Value: 1.0, Host: "h"}
	if err := ValidateSample(s); err == nil {
		t.Error("expected error for long name")
	}
}

func TestValidateSample_TooManyLabels(t *testing.T) {
	labels := make(map[string]string)
	for i := 0; i < 51; i++ {
		labels[string(rune('a'+i%26))+string(rune('0'+i/26))] = "v"
	}
	s := &pb.MetricSample{Name: "cpu", Value: 1.0, Host: "h", Labels: labels}
	if err := ValidateSample(s); err == nil {
		t.Error("expected error for too many labels")
	}
}

func TestValidateSample_InternalLabelPrefix(t *testing.T) {
	s := &pb.MetricSample{
		Name:   "cpu",
		Value:  1.0,
		Host:   "h",
		Labels: map[string]string{"__internal": "bad"},
	}
	if err := ValidateSample(s); err == nil {
		t.Error("expected error for label with __ prefix")
	}
}

func TestValidateBatch(t *testing.T) {
	batch := &pb.MetricBatch{
		Region:      "ap-sh-1",
		CollectorID: "test-collector",
		Samples: []*pb.MetricSample{
			{Name: "cpu", Value: 42.0, Host: "web-01"},
			{Name: "", Value: 1.0, Host: "bad"}, // invalid
			{Name: "mem", Value: 8e9, Host: "web-01"},
			{Name: "disk", Value: 72.0, Host: ""}, // invalid
		},
	}

	accepted, rejected, errs := ValidateBatch(batch)
	if accepted != 2 {
		t.Errorf("accepted = %d, want 2", accepted)
	}
	if rejected != 2 {
		t.Errorf("rejected = %d, want 2", rejected)
	}
	if len(errs) != 2 {
		t.Errorf("errors = %d, want 2", len(errs))
	}
}
