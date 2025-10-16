package retry

import (
	"context"
	"time"
)

// checkContext verifies if the context is still valid (not cancelled or timed out)
// Returns the context error if cancelled/expired, nil otherwise
func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// waitWithContext waits for the specified duration or until context is cancelled
// Returns the context error if cancelled during wait, nil if wait completed normally
func waitWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
