package ingest

import (
	"log"
	"sync"

	pb "github.com/pingan/hydra/internal/proto"
)

// StreamServer handles incoming MetricBatch submissions from Regional Collectors.
type StreamServer struct {
	mu     sync.Mutex
	fanout *Fanout
}

// NewStreamServer creates a StreamServer backed by the given fanout.
func NewStreamServer(fanout *Fanout) *StreamServer {
	return &StreamServer{fanout: fanout}
}

// HandleBatch processes a single MetricBatch and returns an acknowledgement.
func (s *StreamServer) HandleBatch(batch *pb.MetricBatch) *pb.IngestAck {
	if batch == nil || len(batch.Samples) == 0 {
		return &pb.IngestAck{Success: true, Message: "empty batch"}
	}

	log.Printf("stream: received batch from %s/%s with %d samples",
		batch.Region, batch.CollectorID, len(batch.Samples))

	return s.fanout.ProcessBatch(batch)
}

// IngestCount returns the total number of batches processed (for metrics).
func (s *StreamServer) IngestCount() int64 {
	// Simplistic counter; production would track this properly
	return 0
}
