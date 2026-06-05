package grpcstream

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/pingan/hydra/internal/proto"
	"github.com/pingan/hydra/internal/buffer"
)

// Client manages the gRPC streaming connection to the Central Aggregator.
// It reads from a RingBuffer and sends MetricBatches over the stream.
type Client struct {
	centralAddr string
	region      string
	collectorID string
	buffer      *buffer.RingBuffer

	mu       sync.Mutex
	seq      int64
	running  bool
	stopCh   chan struct{}
}

// NewClient creates a gRPC stream client.
func NewClient(centralAddr, region, collectorID string, buf *buffer.RingBuffer) *Client {
	return &Client{
		centralAddr: centralAddr,
		region:      region,
		collectorID: collectorID,
		buffer:      buf,
		stopCh:      make(chan struct{}),
	}
}

// Start begins the sender goroutine that drains the buffer and sends batches.
func (c *Client) Start(ctx context.Context) {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	go c.sendLoop(ctx)
	log.Printf("gRPC stream client started (central: %s)", c.centralAddr)
}

// Stop signals the sender goroutine to shut down.
func (c *Client) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		close(c.stopCh)
		c.running = false
	}
}

// sendLoop periodically drains the ring buffer and sends batches.
// In production, this would use a real gRPC bidirectional stream.
// For now it simulates batch delivery via direct function call.
func (c *Client) sendLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			c.flush()
			log.Println("gRPC stream client stopped")
			return
		case <-ticker.C:
			c.flush()
		case <-c.buffer.NotEmpty():
			if c.buffer.Len() >= 100 {
				c.flush()
			}
		case <-ctx.Done():
			c.flush()
			return
		}
	}
}

func (c *Client) flush() {
	batch := c.buffer.Drain(1000)
	if len(batch) == 0 {
		return
	}

	c.mu.Lock()
	c.seq++
	seq := c.seq
	c.mu.Unlock()

	mb := &pb.MetricBatch{
		Region:      c.region,
		CollectorID: c.collectorID,
		Samples:     batch,
		BatchSeq:    seq,
	}

	// In production this would be sent via gRPC stream.Send(mb)
	// For now we log the flush for verification.
	log.Printf("[stream] flushed batch #%d: %d samples to %s (region=%s)",
		seq, len(batch), c.centralAddr, c.region)
	_ = mb
}

// SendBatch sends a single batch (non-streaming fallback).
func (c *Client) SendBatch(batch *pb.MetricBatch) error {
	c.mu.Lock()
	c.seq++
	batch.BatchSeq = c.seq
	batch.CollectorID = c.collectorID
	batch.Region = c.region
	c.mu.Unlock()

	// In production: gRPC unary call
	log.Printf("[stream] send batch #%d: %d samples", batch.BatchSeq, len(batch.Samples))
	_ = batch
	return fmt.Errorf("gRPC not connected (demo mode)")
}
