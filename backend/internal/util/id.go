package util

import (
	"crypto/rand"
	"fmt"
	"sync/atomic"
)

var counter uint64

// NewID generates a unique identifier string.
func NewID() string {
	// Read 8 random bytes + atomic counter for uniqueness
	b := make([]byte, 8)
	rand.Read(b)
	seq := atomic.AddUint64(&counter, 1)
	return fmt.Sprintf("%x-%x", b, seq)
}
