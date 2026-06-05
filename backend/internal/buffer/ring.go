package buffer

import (
	"sync"

	pb "github.com/pingan/hydra/internal/proto"
)

// RingBuffer is a lock-free(ish) circular buffer for batching metrics before
// sending to the central aggregator via gRPC stream.
type RingBuffer struct {
	mu       sync.Mutex
	buf      []*pb.MetricSample
	head     int
	tail     int
	size     int
	count    int
	notEmpty chan struct{}
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buf:      make([]*pb.MetricSample, capacity),
		size:     capacity,
		notEmpty: make(chan struct{}, 1),
	}
}

// Push adds a metric sample to the buffer. Returns false if the buffer is full.
func (rb *RingBuffer) Push(sample *pb.MetricSample) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count >= rb.size {
		return false
	}

	rb.buf[rb.tail] = sample
	rb.tail = (rb.tail + 1) % rb.size
	rb.count++

	select {
	case rb.notEmpty <- struct{}{}:
	default:
	}

	return true
}

// Pop removes and returns the oldest sample from the buffer.
// Returns nil if the buffer is empty.
func (rb *RingBuffer) Pop() *pb.MetricSample {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return nil
	}

	sample := rb.buf[rb.head]
	rb.buf[rb.head] = nil
	rb.head = (rb.head + 1) % rb.size
	rb.count--

	return sample
}

// Drain removes up to N samples from the buffer in a batch.
func (rb *RingBuffer) Drain(n int) []*pb.MetricSample {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return nil
	}

	drain := n
	if drain > rb.count {
		drain = rb.count
	}

	batch := make([]*pb.MetricSample, drain)
	for i := 0; i < drain; i++ {
		batch[i] = rb.buf[rb.head]
		rb.buf[rb.head] = nil
		rb.head = (rb.head + 1) % rb.size
	}
	rb.count -= drain

	return batch
}

// Len returns the current number of samples in the buffer.
func (rb *RingBuffer) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

// NotEmpty returns a channel that signals when data is available.
func (rb *RingBuffer) NotEmpty() <-chan struct{} {
	return rb.notEmpty
}
