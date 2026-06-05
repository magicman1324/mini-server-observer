package buffer

import (
	"sync"
	"testing"

	pb "github.com/pingan/hydra/internal/proto"
)

func TestRingBufferPushPop(t *testing.T) {
	rb := NewRingBuffer(10)

	s := &pb.MetricSample{Name: "cpu", Value: 42.0, Host: "web-01"}
	if !rb.Push(s) {
		t.Fatal("push should succeed on non-full buffer")
	}
	if rb.Len() != 1 {
		t.Fatalf("len = %d, want 1", rb.Len())
	}

	popped := rb.Pop()
	if popped == nil {
		t.Fatal("pop should return a sample")
	}
	if popped.Name != "cpu" {
		t.Errorf("name = %q, want cpu", popped.Name)
	}
	if rb.Len() != 0 {
		t.Fatalf("len = %d, want 0 after pop", rb.Len())
	}
}

func TestRingBufferDrain(t *testing.T) {
	rb := NewRingBuffer(100)
	for i := 0; i < 50; i++ {
		rb.Push(&pb.MetricSample{Name: "m", Value: float64(i)})
	}

	batch := rb.Drain(30)
	if len(batch) != 30 {
		t.Fatalf("drained %d, want 30", len(batch))
	}
	if rb.Len() != 20 {
		t.Fatalf("remaining %d, want 20", rb.Len())
	}

	// Drain more than remaining
	batch = rb.Drain(100)
	if len(batch) != 20 {
		t.Fatalf("drained %d, want 20", len(batch))
	}
	if rb.Len() != 0 {
		t.Fatalf("remaining %d, want 0", rb.Len())
	}
}

func TestRingBufferOverflow(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Push(&pb.MetricSample{Name: "a", Value: 1})
	rb.Push(&pb.MetricSample{Name: "b", Value: 2})
	rb.Push(&pb.MetricSample{Name: "c", Value: 3})

	if rb.Push(&pb.MetricSample{Name: "d", Value: 4}) {
		t.Error("push should fail on full buffer")
	}
	if rb.Len() != 3 {
		t.Errorf("len = %d, want 3", rb.Len())
	}
}

func TestRingBufferConcurrency(t *testing.T) {
	rb := NewRingBuffer(2000)
	var wg sync.WaitGroup

	// 10 producers, 100 samples each
	for p := 0; p < 10; p++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				rb.Push(&pb.MetricSample{Name: "m", Value: float64(i)})
			}
		}()
	}
	wg.Wait()

	if rb.Len() != 1000 {
		t.Errorf("len = %d, want 1000", rb.Len())
	}

	// Single consumer drains all
	drained := rb.Drain(1000)
	if len(drained) != 1000 {
		t.Errorf("drained %d, want 1000", len(drained))
	}
	if rb.Len() != 0 {
		t.Errorf("len = %d, want 0", rb.Len())
	}
}

func TestRingBufferPopEmpty(t *testing.T) {
	rb := NewRingBuffer(10)
	if s := rb.Pop(); s != nil {
		t.Error("pop on empty buffer should return nil")
	}
}

func TestRingBufferDrainEmpty(t *testing.T) {
	rb := NewRingBuffer(10)
	if batch := rb.Drain(5); len(batch) != 0 {
		t.Errorf("drain on empty buffer should return empty slice, got %d", len(batch))
	}
}
