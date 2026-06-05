package ingest

import (
	"log"
	"time"

	pb "github.com/pingan/hydra/internal/proto"
	"github.com/pingan/hydra/internal/model"
	"github.com/pingan/hydra/internal/repository"
)

// Fanout distributes incoming metric batches to the storage writer and alert engine.
type Fanout struct {
	metricRepo  repository.MetricRepository
	alertCh     chan<- []*model.MetricPoint
}

// NewFanout creates a Fanout that writes to the given repos and forwards to the alert channel.
func NewFanout(metricRepo repository.MetricRepository, alertCh chan<- []*model.MetricPoint) *Fanout {
	return &Fanout{
		metricRepo: metricRepo,
		alertCh:    alertCh,
	}
}

// ProcessBatch validates, converts, and fans out a MetricBatch.
func (f *Fanout) ProcessBatch(batch *pb.MetricBatch) *pb.IngestAck {
	accepted, rejected, errs := ValidateBatch(batch)

	for _, err := range errs {
		log.Printf("ingest validation: %v", err)
	}

	if accepted == 0 {
		return &pb.IngestAck{
			Success:  rejected == 0,
			Message:  "all samples rejected",
			Accepted: 0,
			Rejected: int32(rejected),
		}
	}

	// Convert proto samples to model MetricPoints
	now := time.Now()
	var points []*model.MetricPoint
	for _, s := range batch.Samples {
		if err := ValidateSample(s); err != nil {
			continue
		}
		points = append(points, &model.MetricPoint{
			TS:     now,
			Name:   s.Name,
			Value:  s.Value,
			Host:   s.Host,
			Region: s.Region,
			Labels: s.Labels,
		})
	}

	// Write to storage
	if err := f.metricRepo.InsertBatch(points); err != nil {
		log.Printf("ingest: failed to insert batch: %v", err)
		return &pb.IngestAck{
			Success:  false,
			Message:  "storage write failed",
			Accepted: 0,
			Rejected: int32(len(batch.Samples)),
		}
	}

	// Forward to alert engine (non-blocking)
	if f.alertCh != nil {
		select {
		case f.alertCh <- points:
		default:
			log.Printf("ingest: alert channel full, dropping %d points", len(points))
		}
	}

	log.Printf("ingest: accepted %d samples from region=%s collector=%s seq=%d (rejected=%d)",
		accepted, batch.Region, batch.CollectorID, batch.BatchSeq, rejected)

	return &pb.IngestAck{
		Success:  true,
		Message:  "ok",
		Accepted: int32(accepted),
		Rejected: int32(rejected),
	}
}
