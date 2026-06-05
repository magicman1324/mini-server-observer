package notify

import (
	"log"
	"time"
)

// RetryWithBackoff executes a function with exponential backoff retries.
func RetryWithBackoff(name string, maxRetries int, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("notify: retry %d/%d for %s after %v", attempt, maxRetries, name, backoff)
			time.Sleep(backoff)
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		log.Printf("notify: attempt %d/%d for %s failed: %v", attempt+1, maxRetries+1, name, lastErr)
	}

	log.Printf("notify: all %d retries exhausted for %s", maxRetries+1, name)
	return lastErr
}
