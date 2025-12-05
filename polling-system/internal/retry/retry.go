package retry

import (
	"context"
	"time"
)

// DoWithRetry executes fn up to attempts times with exponential backoff.
// It stops early if the context is canceled.
func DoWithRetry(ctx context.Context, attempts int, baseDelay time.Duration, fn func() error) error {
	var err error
	delay := baseDelay

	for i := 0; i < attempts; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err = fn(); err == nil {
			return nil
		}

		if i == attempts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}
	return err
}
